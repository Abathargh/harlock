package main

import (
	"encoding/binary"
	"encoding/hex"
	"io"
)

type HexFile struct {
	base    int
	records []Record
}

func ReadAll(in io.ByteReader) (*HexFile, error) {
	// add check for only one eof record
	var records []Record
	rec, err := parseRecord(in)
	for ; err == nil; rec, err = parseRecord(in) {
		if validateRecord(rec) {
			records = append(records, rec)
		}
	}

	if err == NoMoreRecordsErr {
		return &HexFile{base: 0, records: records}, nil
	}
	return nil, err
}

func parseRecord(input io.ByteReader) (Record, error) {
	var record Record

	curr, err := input.ReadByte()
	if err != nil {
		return Record{}, NoMoreRecordsErr
	}

	if curr != startCode {
		return Record{}, MissingStartCodeErr
	}

	for curr != '\r' && curr != '\n' {
		record = append(record, curr)
		curr, err = input.ReadByte()
		if err != nil {
			return Record{}, WrongRecordFormatErr
		}
	}

	if curr == '\r' {
		curr, err = input.ReadByte()
		if err != nil || curr != '\n' {
			return nil, WrongRecordFormatErr
		}
	}
	return record, nil
}

func validateRecord(rec Record) bool {
	recordLen := len(rec)
	if recordLen < minLength {
		return false
	}

	dataLenBytes := make([]byte, 2)
	_, err := hex.Decode(dataLenBytes, rec[1:3])
	if err != nil {
		return false
	}

	dataLen := binary.LittleEndian.Uint16(dataLenBytes)
	if recordLen != int(minLength+dataLen) {
		return false
	}
	return true
}
