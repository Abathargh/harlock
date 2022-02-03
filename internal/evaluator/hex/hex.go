package main

import (
	"io"
)

type HexFile struct {
	base    int
	records []*Record
}

func ReadAll(in io.ByteReader) (*HexFile, error) {
	// add check for only one eof record
	var records []*Record
	rec, err := parseRecord(in)
	for ; err == nil; rec, err = parseRecord(in) {
		if !validateRecord(rec) {
			return nil, WrongRecordFormatErr
		}
		records = append(records, rec)
	}

	if err == NoMoreRecordsErr {
		return &HexFile{base: 0, records: records}, nil
	}
	return nil, err
}
