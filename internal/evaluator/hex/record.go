package main

import (
	"encoding/binary"
	"encoding/hex"
)

const (
	startCode = 0x3A // ':'

	startCodeLen  = 1
	byteCountLen  = 2
	addressLen    = 4
	recordTypeLen = 2
	checksumLen   = 2

	minLength = startCodeLen + byteCountLen + addressLen + recordTypeLen + checksumLen
)

type HexadecimalError string

func (h HexadecimalError) Error() string {
	return string(h)
}

const (
	MissingStartCodeErr  = HexadecimalError("the passed record does not start with the correct start code")
	WrongRecordFormatErr = HexadecimalError("the passed record is not a correct hex record")
	NoMoreRecordsErr     = HexadecimalError("no more records")
)

type RecordType uint

const (
	DataRecord RecordType = iota
	EOFRecord
	ExtendedSegmentAddrRecord
	StartSegmentAddrRecord
	ExtendedLinearAddrRecord
	StartLinearAddrRecord
	InvalidRecord
)

type Record []byte

func (r Record) RecordType() RecordType {
	rType, err := hexToInt(r[7:9], false)
	if err != nil || rType > uint64(InvalidRecord) {
		return InvalidRecord
	}
	return RecordType(rType)
}

func hexToInt(data []byte, littleEndian bool) (uint64, error) {
	dataLen := len(data)
	if dataLen%2 == 1 {
		return 0, WrongRecordFormatErr
	}
	dataLenBytes := make([]byte, 8)
	_, err := hex.Decode(dataLenBytes, data)
	if err != nil {
		return 0, err
	}

	var endianness binary.ByteOrder
	if littleEndian {
		endianness = binary.LittleEndian
	} else {
		endianness = binary.BigEndian
	}

	return endianness.Uint64(dataLenBytes), nil
}
