package hex

import (
	"encoding/hex"
	"fmt"
	"io"
)

type File struct {
	base    uint32
	records []*Record
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
		return &File{base: 0, records: records}, nil
	}
	return nil, NoEofRecordErr
}

func (hf *File) ReadAt(pos uint32, size int) ([]byte, error) {
	// TODO momentary hack: hf.base should be local and not a struct member
	defer func() { hf.base = 0 }()
	if size < 1 {
		return nil, fmt.Errorf("cannot read less than one byte")
	}

	// we are reading hex digits, 2 hex digits = 1 byte
	size *= 2

	fromExtendedRec := false
	for idx, record := range hf.records {
		switch record.rType {
		case StartSegmentAddrRecord:
			// Do nothing
		case ExtendedSegmentAddrRecord:
			data, err := hexToInt[uint16](record.ReadData(), false)
			if err != nil {
				return nil, fmt.Errorf("invalid record data: %w", err)
			}
			hf.base = uint32(data) * 16
			fromExtendedRec = true
		case StartLinearAddrRecord:
			// Do nothing
		case ExtendedLinearAddrRecord:
			data, err := hexToInt[uint16](record.ReadData(), false)
			if err != nil {
				return nil, fmt.Errorf("invalid record data: %w", err)
			}
			extendedBase := uint32(data)
			hf.base = extendedBase << 16
			fromExtendedRec = true
		case EOFRecord:
			// Do nothing
		case DataRecord:
			// if the earlier record is an extended record the cursor must
			// re-based onto the start of this data address
			if fromExtendedRec {
				hf.base += uint32(record.Address())
				fromExtendedRec = false
			}
			uLen := uint32(record.length) * 2

			// Found the record where the access should begin
			if pos >= hf.base && pos <= hf.base+uLen {
				retData := make([]byte, size)
				start := (pos - hf.base) * 2
				end := start + uint32(size)

				// if the user wants to read further, this record is read whole
				if end > uLen {
					end = uLen
				}

				// otherwise, read only as needed
				copy(retData, record.ReadData()[start:end])
				written := int(end - start)

				// the read operation is not finished with the current record
				for written < size && idx != len(hf.records)-1 {
					idx++
					current := hf.records[idx]
					// bad access: trying to read data with holes in it
					if current.rType != DataRecord {
						return nil, CustomError(AccessOutOfBounds,
							"no data with %d size found at @%d", size, pos)
					}

					// read accordingly to how much data is missing until completion
					if current.length > size-written {
						copy(retData[written:], current.ReadData()[:size-written])
						written = size
					} else {
						copy(retData[written:], current.ReadData())
						written += current.length
					}
				}

				// bad access: trying to read more than what is there on the hex file
				if written < size {
					return nil, AccessOutOfBounds
				}

				// This should never file since the hex file is validated
				byteData := make([]byte, len(retData)/2)
				_, _ = hex.Decode(byteData, retData)
				return byteData, nil
			}
			hf.base += uint32(record.length)
		}
	}
	return nil, AccessOutOfBounds
}

func (hf *File) WriteAt(pos int, data []byte) ([]byte, error) {
	return nil, nil
}

func (hf *File) Contains(data []byte) bool {
	return false
}
