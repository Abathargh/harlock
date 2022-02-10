package hex

import (
	"fmt"
	"io"
)

type File struct {
	base    uint32
	records []*Record
}

func ReadAll(in io.ByteReader) (*File, error) {
	// TODO add check for only one eof record
	var records []*Record
	rec, err := ParseRecord(in)
	for ; err == nil; rec, err = ParseRecord(in) {
		records = append(records, rec)
	}

	if err == NoMoreRecordsErr && records[len(records)-1].rType == EOFRecord {
		return &File{base: 0, records: records}, nil
	}
	return nil, err
}

func (hf *File) ReadAt(pos uint32, size int) ([]byte, error) {
	if size < 1 {
		return nil, fmt.Errorf("cannot read less than one byte")
	}

	for idx, record := range hf.records {
		switch record.rType {
		case StartSegmentAddrRecord:
			// Do nothing
		case ExtendedSegmentAddrRecord:
			data, err := hexToInt(record.ReadData(), true)
			if err != nil {
				return nil, fmt.Errorf("invalid record data: %w", err)
			}
			hf.base += uint32(data) * 16
		case StartLinearAddrRecord:
			// Do nothing
		case ExtendedLinearAddrRecord:
			data, err := hexToInt(record.ReadData(), true)
			if err != nil {
				return nil, fmt.Errorf("invalid record data: %w", err)
			}
			extendedBase := uint32(data)
			hf.base = extendedBase << 16
			// TODO exiting here => base =hf.base + start next addr of data record
			// TODO use a flag that is checked at the beginning of a data record
		case EOFRecord:
			// Do nothing
		case DataRecord:
			uLen := uint32(record.length)
			if pos >= hf.base && pos <= hf.base+uLen {
				// this record contains the start of
				// the data to read
				retData := make([]byte, size)
				start := pos - hf.base
				end := start + uint32(size)
				if end > uLen {
					end = uLen
				}
				copy(retData, record.ReadData()[start:end])
				written := int(end - start)
				for written < size && idx != len(hf.records)-1 {
					idx++
					current := hf.records[idx]
					if current.rType != DataRecord {
						return nil, fmt.Errorf("no data with %d size found at @%d", size, pos)
					}

					if current.length > size-written {
						copy(retData[written:], current.ReadData()[:size-written])
						written = size
					} else {
						copy(retData[written:], current.ReadData())
						written += current.length
					}
				}

				if written < size {
					// TODO const err
					return nil, fmt.Errorf("out of bounds read")
				}
				return retData, nil
			}
			pos += uint32(record.length)
		}
	}
	// TODO add err
	return nil, nil
}

func (hf *File) WriteAt(pos int, data []byte) ([]byte, error) {
	return nil, nil
}

func (hf *File) Contains(data []byte) bool {
	return false
}
