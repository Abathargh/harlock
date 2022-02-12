package hex

import (
	"encoding/hex"
	"io"
)

type File struct {
	records []*Record
}

type recordView struct {
	start    int
	firstIdx int
	records  []Record
}

func ReadAll(in io.ByteReader) (*File, error) {
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

	if err == NoMoreRecordsErr && records != nil && records[len(records)-1].rType == EOFRecord {
		return &File{records: records}, nil
	}
	return nil, NoEofRecordErr
}

func (hf *File) ReadAt(pos uint32, size int) ([]byte, error) {
	block, err := hf.accessFileAt(pos, size)
	if err != nil {
		return nil, err
	}

	written := 0
	hexSize := size * 2
	hexData := make([]byte, hexSize)

	for idx, record := range block.records {
		recordData := record.ReadData()
		if idx == 0 && block.start != 0 {
			if block.start+hexSize < len(recordData) {
				copy(hexData, recordData[block.start:block.start+hexSize])
				break
			}
			copy(hexData, recordData[block.start:])
			written += len(recordData) - block.start
			continue
		}

		if record.length > hexSize-written {
			copy(hexData[written:], recordData[:hexSize-written])
			break
		} else {
			copy(hexData[written:], recordData)
			written += record.length * 2
		}
	}

	byteData := make([]byte, len(hexData)/2)
	_, _ = hex.Decode(byteData, hexData)
	return byteData, nil
}

func (hf *File) WriteAt(pos uint32, data []byte) error {
	return nil
}

func (hf *File) Contains(data []byte) bool {
	return false
}

func (hf *File) accessFileAt(pos uint32, size int) (*recordView, error) {
	if size < 1 {
		// Empty array => no op
		return &recordView{}, nil
	}

	// we are reading hex digits, 2 hex digits = 1 byte
	size *= 2
	base := uint32(0)
	block := &recordView{}

	fromExtendedRec := false
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
			fromExtendedRec = true
		case StartLinearAddrRecord:
			// Do nothing
		case ExtendedLinearAddrRecord:
			data, err := hexToInt[uint16](record.ReadData(), false)
			if err != nil {
				return nil, RecordErr
			}
			extendedBase := uint32(data)
			base = extendedBase << 16
			fromExtendedRec = true
		case EOFRecord:
			// Do nothing
		case DataRecord:
			// if the earlier record is an extended record the cursor must
			// re-based onto the start of this data address
			if fromExtendedRec {
				base += uint32(record.Address())
				fromExtendedRec = false
			}
			uLen := uint32(record.length) * 2

			// Found the record where the access should begin
			if pos >= base && pos <= base+uLen {
				// these checks are needed to know if the access
				// should stop at the first record
				start := (pos - base) * 2
				end := start + uint32(size)
				if end > uLen {
					end = uLen
				}

				// put the first record in the view
				block.start = int((pos - base) * 2)
				block.firstIdx = idx
				block.records = append(block.records, *record)

				alreadyAccessedLen := int(end - start)

				// the access operation is not finished with the current record
				for alreadyAccessedLen < size && idx != len(hf.records)-1 {
					idx++
					current := hf.records[idx]
					// bad access: trying to access data with holes in it
					if current.rType != DataRecord {
						return nil, CustomError(AccessOutOfBounds,
							"no data with %d size found at @%d", size, pos)
					}
					block.records = append(block.records, *current)
					alreadyAccessedLen += current.length
				}

				// bad access: trying to access more than what is there on the hex hf.
				if alreadyAccessedLen < size {
					return nil, AccessOutOfBounds
				}

				// This should never hf. since the hex hf. is validated

				return block, nil
			}
			base += uint32(record.length)
		}
	}
	return nil, AccessOutOfBounds
}
