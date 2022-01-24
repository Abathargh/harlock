package main

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
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

func Checksum(record []byte) ([2]byte, error) {
	var checksum uint64
	var hexStringChecksum [2]byte
	start := 0
	if record[0] == startCode {
		start = 1
	}

	decoded := make([]byte, hex.DecodedLen(len(record[start:])))
	if _, err := hex.Decode(decoded, record[start:]); err != nil {
		return hexStringChecksum, fmt.Errorf("%s: %w", WrongRecordFormatErr, err)
	}

	hexChecksum := uint64(0)
	for _, hexVal := range decoded {
		hexChecksum += uint64(hexVal)
	}

	checksumCont := make([]byte, 8)
	checksum = uint64(^byte(hexChecksum&0x0F) + 1)
	binary.LittleEndian.PutUint64(checksumCont, checksum)
	hex.Encode(hexStringChecksum[:], checksumCont[0:1])

	return hexStringChecksum, nil
}
