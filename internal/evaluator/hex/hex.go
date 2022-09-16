package hex

import (
	"encoding/hex"
	"io"
)

// File implements an Intel Hex-encoded file
type File struct {
	records []*Record
}

// recordView is an internal struct used to
// abstract data accesses to the hex file
type recordView struct {
	start    int
	firstIdx int
	records  []*Record
}

// ReadAll initializes a hex file by reading every byte
// from its source, parsing the records and validating them
func ReadAll(in io.ByteScanner) (*File, error) {
	eof := false
	var records []*Record
	rec, err := ParseRecord(in)
	for ; err == nil; rec, err = ParseRecord(in) {
		if eof && rec.Type() == EOFRecord {
			return nil, MultipleEofErr
		}
		records = append(records, rec)
		if rec.Type() == EOFRecord {
			eof = true
		}
	}

	if err == NoMoreRecordsErr {
		if records != nil && records[len(records)-1].rType == EOFRecord {
			return &File{records: records}, nil
		}
		return nil, NoEofRecordErr
	}

	return nil, err
}

func (hf *File) Iterator() <-chan *Record {
	ch := make(chan *Record)
	go func(recs []*Record, channel chan *Record) {
		for _, rec := range recs {
			ch <- rec
		}
		close(ch)
	}(hf.records, ch)
	return ch
}

// Size returns the number of records in the file
func (hf *File) Size() int {
	return len(hf.records)
}

// Record returns the idx-th record or nil if it does not exist
func (hf *File) Record(idx int) *Record {
	if idx < 0 || idx >= len(hf.records) {
		return nil
	}
	return hf.records[idx]
}

// ReadAt reads size bytes starting from pos position in the
// hex-encoded file. This implements a sort of random access
// to the data mapped in hex-format.
func (hf *File) ReadAt(pos uint32, size int) ([]byte, error) {
	block, err := hf.accessAt(pos, size)
	if err != nil {
		return nil, err
	}

	written := 0
	hexSize := size * 2
	hexData := make([]byte, hexSize)

	for idx, record := range block.records {
		recData := record.ReadData()
		if idx == 0 && block.start != 0 {
			if block.start+hexSize < len(recData) {
				copy(hexData, recData[block.start:block.start+hexSize])
				break
			}
			copy(hexData, recData[block.start:])
			written += len(recData) - block.start
			continue
		}

		if record.length*2 >= hexSize-written {
			copy(hexData[written:], recData[:hexSize-written])
			break
		}

		copy(hexData[written:], recData)
		written += record.length * 2
	}

	byteData := make([]byte, len(hexData)/2)
	_, _ = hex.Decode(byteData, hexData)
	return byteData, nil
}

// WriteAt writes len(data) bytes starting from pos position
// onto the hex-encoded file. The written bytes are passed
// through the data parameter.
func (hf *File) WriteAt(pos uint32, data []byte) error {
	block, err := hf.accessAt(pos, len(data))
	if err != nil {
		return err
	}

	written := 0
	hexSize := len(data) * 2
	hexData := make([]byte, hexSize)
	hex.Encode(hexData, data)

	for idx, record := range block.records {
		recData := record.ReadData()
		if idx == 0 && block.start != 0 {
			if block.start+hexSize < len(recData) {
				copy(recData[block.start:], hexData[:])
				updateChecksum(record)
				break
			}
			copy(recData[block.start:], hexData[:len(recData)-block.start])
			written += len(recData) - block.start
			updateChecksum(record)
			continue
		}

		// from here on, write always start from the
		// first byte of the record, hence no block.start

		// if the current record is bigger
		// than what remains to be written,
		// write the whole remaining buf
		if record.length*2 > hexSize-written {
			copy(recData, hexData[written:])
			updateChecksum(record)
			break
		}

		// there is a part of the buf that will be
		// written on the next record(s)
		copy(recData, hexData[written:written+(record.length*2)])
		written += record.length * 2
		updateChecksum(record)
	}
	return nil
}

// accessAt implements a generic random access feature for hex files
// by returning a recordView that refers to a block of contiguous
// records that span through the [pos; pos+size] interval.
func (hf *File) accessAt(pos uint32, size int) (*recordView, error) {
	if size < 1 {
		// Empty array => no op
		return &recordView{}, nil
	}

	// we are reading hex digits, 2 hex digits = 1 byte
	hexSize := size * 2
	base := uint32(0)
	block := &recordView{}

	for idx, record := range hf.records {
		switch record.rType {
		case StartSegmentAddrRecord:
			// Do nothing
		case ExtendedSegmentAddrRecord:
			data, err := hexToInt[uint16](record.ReadData(), false)
			if err != nil {
				return nil, RecordErr
			}
			base = uint32(data) * 16
		case StartLinearAddrRecord:
			// Do nothing
		case ExtendedLinearAddrRecord:
			data, err := hexToInt[uint16](record.ReadData(), false)
			if err != nil {
				return nil, RecordErr
			}
			extendedBase := uint32(data)
			base = extendedBase << 16
		case EOFRecord:
			// Do nothing
		case DataRecord:
			uLen := uint32(record.length) * 2
			recordBase := uint32(record.Address()) + base

			// Found the record where the access should begin
			if pos >= recordBase && pos < recordBase+uLen {
				// these checks are needed to know if the access
				// should stop at the first record
				start := (pos - recordBase) * 2
				end := start + uint32(hexSize)
				if end > uLen {
					end = uLen
				}

				// put the first record in the view
				block.start = int((pos - recordBase) * 2)
				block.firstIdx = idx
				block.records = append(block.records, record)

				alreadyAccessedLen := int(end - start)

				// the access operation is not finished with the current record
				for alreadyAccessedLen < hexSize && idx != len(hf.records)-1 {
					idx++
					current := hf.records[idx]
					// bad access: trying to access data with holes in it
					if current.rType != DataRecord {
						return nil, CustomError(AccessOutOfBounds,
							"no data with %d size found at @%d", size, pos)
					}
					block.records = append(block.records, current)
					alreadyAccessedLen += current.length
				}

				// bad access: trying to access more than what is there on the hex hf.
				if alreadyAccessedLen < hexSize {
					return nil, AccessOutOfBounds
				}

				// This should never hf. since the hex hf. is validated

				return block, nil
			}
		}
	}
	return nil, AccessOutOfBounds
}

// updateChecksum is a helper function used to fix checksums
// of modified records
func updateChecksum(record *Record) {
	recordChecksum, _ := checksum(record.data)
	var hexChecksum [2]byte
	hex.Encode(hexChecksum[:], []byte{recordChecksum})
	copy(record.Checksum(), hexChecksum[:])
}
