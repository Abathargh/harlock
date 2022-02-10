package hex

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
)

const (
	startCode = 0x3A // ':'
	dataIndex = 9

	startCodeLen  = 1
	byteCountLen  = 2
	addressLen    = 4
	recordTypeLen = 2
	checksumLen   = 2

	// Indexing the Byte Count field
	lenIdx = 1
	lenEnd = lenIdx + byteCountLen

	// Indexing the Address field
	addrIdx = lenEnd
	addrEnd = addrIdx + addressLen

	//Indexing the Record Type
	typeIdx = addrEnd
	typeEnd = typeIdx + recordTypeLen

	minLength = startCodeLen + byteCountLen + addressLen + recordTypeLen + checksumLen
)

type RecordError string

func (r RecordError) Error() string {
	return string(r)
}

const (
	MissingStartCodeErr  = RecordError("the passed record does not start with the correct start code")
	WrongRecordFormatErr = RecordError("the passed record is not a correct hex record")
	DataOutOfBounds      = RecordError("the passed byte slice cannot be held by this record")
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

// Record is an HEX Record that has been validated.
// Instantiate only via ParseRecord
type Record struct {
	length int
	rType  RecordType
	data   []byte
}

// ByteCount returned as an integer
func (r *Record) ByteCount() int {
	if r.data == nil {
		return 0
	}
	return r.length
}

// Address is the record address value
func (r *Record) Address() []byte {
	if r.data == nil {
		return nil
	}
	return r.data[addrIdx:addrEnd]
}

// ReadData returns the data section of the record
func (r *Record) ReadData() []byte {
	if r.data == nil {
		return nil
	}
	return r.data[dataIndex : dataIndex+(r.length*2)]
}

// Checksum of the current record
func (r *Record) Checksum() []byte {
	if r.data == nil {
		return nil
	}
	return r.data[dataIndex+(r.length*2):]
}

// WriteData is used to rewrite the data section of the record.
// This method re-computes the record checksum automatically.
func (r *Record) WriteData(start int, data []byte) error {
	if r.data == nil || start < 0 || start+len(data) > r.length {
		return DataOutOfBounds
	}

	// TODO recalculate checksum
	for idx, b := range data {
		r.data[start+idx] = b
	}
	return nil
}

// ParseRecord initializes a new Record reading from a ByteReader.
// This function returns an error if the byte stream that is read
// does not represent a valid Record.
func ParseRecord(input io.ByteReader) (*Record, error) {
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

	isValid, rType, length := validateRecord(record)
	if !isValid {
		return nil, WrongRecordFormatErr
	}

	record.rType = rType
	record.length = length
	return record, nil
}

func validateRecord(rec *Record) (bool, RecordType, int) {
	recordLen := len(rec.data)
	if recordLen < minLength {
		return false, InvalidRecord, 0
	}

	dataLenBytes := make([]byte, 2)
	_, err := hex.Decode(dataLenBytes, rec.data[lenIdx:lenEnd])
	if err != nil {
		return false, InvalidRecord, 0
	}

	dataLen := binary.LittleEndian.Uint16(dataLenBytes)
	if recordLen != int(minLength+(dataLen*2)) {
		return false, InvalidRecord, 0
	}

	c, err := checksum(rec.data)
	if err != nil {
		return false, InvalidRecord, 0
	}

	h, err := hexToInt(rec.data[dataIndex+(rec.length*2):], true)
	if err != nil || c != h {
		return false, InvalidRecord, 0
	}

	rTypeUint, err := hexToInt(rec.data[typeIdx:typeEnd], true)
	if err != nil || rTypeUint > uint64(InvalidRecord) {
		return false, InvalidRecord, 0
	}

	rType := RecordType(rTypeUint)
	switch rType {
	case ExtendedSegmentAddrRecord:
		fallthrough
	case ExtendedLinearAddrRecord:
		if dataLen != 2 {
			return false, InvalidRecord, 0
		}
	case StartSegmentAddrRecord:
		fallthrough
	case StartLinearAddrRecord:
		addr, err := hexToInt(rec.Address(), true)
		if err != nil || dataLen != 4 || addr != 0 {
			return false, InvalidRecord, 0
		}
	}

	length, _ := hexToInt(rec.data[lenIdx:lenEnd], true)

	return true, rType, int(length)
}

func hexToInt(data []byte, littleEndian bool) (uint64, error) {
	dataLen := len(data)
	if data == nil || dataLen > 8 {
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
