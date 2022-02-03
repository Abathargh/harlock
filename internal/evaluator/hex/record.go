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

type RecordError string

func (r RecordError) Error() string {
	return string(r)
}

const (
	MissingStartCodeErr  = RecordError("the passed record does not start with the correct start code")
	WrongRecordFormatErr = RecordError("the passed record is not a correct hex record")
	WrongChecksumErr     = RecordError("the passed record has got an incorrect checksum")
	NoMoreRecordsErr     = RecordError("no more records")
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

type Record struct {
	length int
	rType  RecordType
	data   []byte
}

func (r Record) RecordType() RecordType {
	rType, err := hexToInt(r.data[7:9], false)
	if err != nil || rType > uint64(InvalidRecord) {
		return InvalidRecord
	}
	return RecordType(rType)
}

func parseRecord(input io.ByteReader) (*Record, error) {
	record := &Record{}

	curr, err := input.ReadByte()
	if err != nil {
		return nil, NoMoreRecordsErr
	}

	if curr != startCode {
		return nil, MissingStartCodeErr
	}

	for curr != '\r' && curr != '\n' {
		record.data = append(record.data, curr)
		curr, err = input.ReadByte()
		if err != nil {
			return nil, WrongRecordFormatErr
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

func validateRecord(rec *Record) bool {
	recordLen := len(rec.data)
	if recordLen < minLength {
		return false
	}

	dataLenBytes := make([]byte, 2)
	if _, err := hex.Decode(dataLenBytes, rec.data[1:3]); err != nil {
		return false
	}

	dataLen := binary.LittleEndian.Uint16(dataLenBytes)
	if recordLen != int(minLength+(dataLen*2)) {
		return false
	}

	c, err := checksum(rec.data)
	if err != nil {
		return false
	}

	h, err := hexToInt(rec.data[recordLen-2:recordLen], true)
	if err != nil || c != h {
		return false
	}
	return true
}

func hexToInt(data []byte, littleEndian bool) (uint64, error) {
	dataLen := len(data)
	if data == nil || dataLen > 8 {
		return 0, recordError(WrongRecordFormatErr, "wrong byte width: %s", dataLen)
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

func checksum(record []byte) (uint64, error) {
	var cs uint64
	start := 0
	end := len(record) - 2
	if record[0] == startCode {
		start = 1
	}

	decoded := make([]byte, hex.DecodedLen(len(record[start:end])))
	if _, err := hex.Decode(decoded, record[start:end]); err != nil {
		return 0, fmt.Errorf("%s: %w", WrongRecordFormatErr, err)
	}

	hexChecksum := uint64(0)
	for _, hexVal := range decoded {
		hexChecksum += uint64(hexVal)
	}

	cs = uint64(^byte(hexChecksum&0xFF) + 1)
	return cs, nil
}

func checksumBytes(record []byte) ([2]byte, error) {
	var hexStringChecksum [2]byte
	checksumCont := make([]byte, 8)
	cs, err := checksum(record)
	if err != nil {
		return hexStringChecksum, err
	}
	binary.LittleEndian.PutUint64(checksumCont, cs)
	hex.Encode(hexStringChecksum[:], checksumCont[0:1])
	return hexStringChecksum, nil
}

func recordError(err RecordError, msg string, args ...interface{}) error {
	msgErr := fmt.Errorf("%w: msg", err)
	return fmt.Errorf(msgErr.Error(), args...)
}
