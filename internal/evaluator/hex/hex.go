package main

import (
	"fmt"
	"io"
)

type HexFile struct {
	base    uint32
	records []*Record
}

func ReadAll(in io.ByteReader) (*HexFile, error) {
	// TODO add check for only one eof record
	var records []*Record
	rec, err := parseRecord(in)
	for ; err == nil; rec, err = parseRecord(in) {
		records = append(records, rec)
	}

	if err == NoMoreRecordsErr && records[len(records)-1].rType == EOFRecord {
		return &HexFile{base: 0, records: records}, nil
	}
	return nil, err
}

func (hf *HexFile) ReadAt(pos uint32, size int) ([]byte, error) {
	if size < 1 {
		return nil, fmt.Errorf("cannot read less than one byte")
	}

	for idx, record := range hf.records {
		switch record.rType {
		case StartSegmentAddrRecord:
			// Do nothing
		case ExtendedSegmentAddrRecord:
			data, err := hexToInt(record.readData(), false)
			if err != nil {
				return nil, fmt.Errorf("invalid record data: %w", err)
			}
			hf.base += uint32(data) * 16
		case StartLinearAddrRecord:
			// Do nothing
		case ExtendedLinearAddrRecord:
			data, err := hexToInt(record.readData(), false)
			if err != nil {
				return nil, fmt.Errorf("invalid record data: %w", err)
			}
			extendedBase := uint32(data)
			hf.base = extendedBase << 16
		case EOFRecord:
			// Do nothing
		case DataRecord:
			uLen := uint32(record.length)
			if pos >= hf.base && pos <= hf.base+uLen {
				retData := make([]byte, size)
				start := pos - hf.base
				end := start + uint32(size)
				if end > uLen {
					end = uLen
				}
				copy(retData, record.readData()[start:end])
				written := int(end - start)
				for written < size && idx != len(hf.records)-1 {
					idx++
					current := hf.records[idx]
					if current.rType != DataRecord {
						return nil, fmt.Errorf("no data with %d size found at @%d", size, pos)
					}

					if current.length > size-written {
						copy(retData[written:], current.readData()[:size-written])
						written = size
					} else {
						copy(retData[written:], current.readData())
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
	return nil, nil
}

func (hf *HexFile) WriteAt(pos int, data []byte) ([]byte, error) {
	return nil, nil
}

func (hf *HexFile) Contains(data []byte) bool {
	return false
}
