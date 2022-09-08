package hex

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"regexp/syntax"
	"strings"
	"unsafe"
)

const (
	startCode = 0x3A // The ':' character

	// Length for the Start Code field
	startCodeLen = 1

	// Length and indexes for the Byte Count field
	countLen = 2
	countIdx = 1
	countEnd = countIdx + countLen

	// Length and indexes for  the Address field
	addrLen = 4
	addrIdx = countEnd
	addrEnd = addrIdx + addrLen

	// Length and indexes for  the Record field
	typeLen = 2
	typeIdx = addrEnd
	typeEnd = typeIdx + typeLen

	// Index for the Data field
	dataIdx = 9

	// Length for the Checksum field
	checksumLen = 2

	// Minimal Length for the whole word (e.g. EOF record)
	minLength = startCodeLen + countLen + addrLen + typeLen + checksumLen
)

var digits = map[byte]struct{}{
	'0':  {},
	'1':  {},
	'2':  {},
	'3':  {},
	'4':  {},
	'5':  {},
	'6':  {},
	'7':  {},
	'8':  {},
	'9':  {},
	'a':  {},
	'b':  {},
	'c':  {},
	'd':  {},
	'e':  {},
	'f':  {},
	'A':  {},
	'B':  {},
	'C':  {},
	'D':  {},
	'E':  {},
	'F':  {},
	'\n': {},
	'\r': {},
}

// RecordType identifies the type of hex record (Data, EOF, etc.)
type RecordType uint

const (
	DataRecord RecordType = iota // A DataRecord contains data in hex format
	EOFRecord                    // An EOFRecord identifies the end of a hex file
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

// AsString returns a string representation of the record
func (r *Record) AsString() string {
	if r.data == nil {
		return ""
	}
	return strings.ToUpper(string(r.data))
}

// ByteCount returned as an integer
func (r *Record) ByteCount() int {
	if r.data == nil {
		return 0
	}
	return r.length
}

// AddressBytes is the hex representation of
// the record address value
func (r *Record) AddressBytes() []byte {
	if r.data == nil {
		return nil
	}
	return r.data[addrIdx:addrEnd]
}

// Address is the record address value
func (r *Record) Address() uint16 {
	if r.data == nil {
		return 0
	}

	addr, err := hexToInt[uint16](r.data[addrIdx:addrEnd], false)
	if err != nil {
		return 0
	}
	return addr
}

// Type is the record type
func (r *Record) Type() RecordType {
	if r.data == nil {
		return InvalidRecord
	}
	return r.rType
}

// ReadData returns the data section of the record
func (r *Record) ReadData() []byte {
	if r.data == nil {
		return nil
	}
	return r.data[dataIdx : dataIdx+(r.length*2)]
}

// Checksum of the current record
func (r *Record) Checksum() []byte {
	if r.data == nil {
		return nil
	}
	return r.data[dataIdx+(r.length*2):]
}

// WriteData is used to rewrite the data section of the record.
// This method re-computes the record checksum automatically.
func (r *Record) WriteData(start int, data []byte) error {
	if r.data == nil || start < 0 || start+len(data) > (2*r.length) {
		return DataOutOfBounds
	}

	for _, b := range data {
		if !syntax.IsWordChar(rune(b)) {
			return InvalidHexDigit
		}
	}

	copy(r.data[dataIdx+start:], data)
	newSum, err := checksumBytes(r.data)
	if err != nil {
		return err
	}

	copy(r.data[dataIdx+(r.length*2):], newSum)

	return nil
}

// ParseRecord initializes a new Record reading from a ByteReader.
// This function returns an error if the byte stream that is read
// does not represent a valid Record.
func ParseRecord(input io.ByteScanner) (*Record, error) {
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
		_, ok := digits[curr]
		if !ok || err != nil {
			return nil, WrongRecordFormatErr
		}
	}

	// support \r, \n and \r\n as line terminators
	// wikipedia indicates that any of these are ok
	// microchip does too
	if curr == '\r' {
		curr, err = input.ReadByte()
		if err != nil || (curr != ':' && curr != '\n') {
			return nil, WrongRecordFormatErr
		}
		if curr == ':' {
			_ = input.UnreadByte()
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

// validateRecord validates a Record that is being parsed
func validateRecord(rec *Record) (bool, RecordType, int) {
	recordLen := len(rec.data)
	if recordLen < minLength {
		return false, InvalidRecord, 0
	}

	dataLenBytes := make([]byte, 2)
	_, err := hex.Decode(dataLenBytes, rec.data[countIdx:countEnd])
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

	h, err := hexToInt[uint8](rec.data[dataIdx+(dataLen*2):], true)
	if err != nil || c != h {
		return false, InvalidRecord, 0
	}

	rTypeUint, err := hexToInt[uint8](rec.data[typeIdx:typeEnd], true)
	if err != nil || rTypeUint > uint8(InvalidRecord) {
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
		addr := rec.Address()
		if err != nil || dataLen != 4 || addr != 0 {
			return false, InvalidRecord, 0
		}
	}

	byteCount, _ := hexToInt[uint8](rec.data[countIdx:countEnd], true)

	return true, rType, int(byteCount)
}

// Unsigned is a constraint for unsigned integers
// with explicit bit-width.
type Unsigned interface {
	uint8 | uint16 | uint32 | uint64
}

// hexToInt decodes a byte array with len < 16 into the corresponding
// unsigned integer passed as type parameter, with the specified
// endianness.
func hexToInt[U Unsigned](data []byte, littleEndian bool) (U, error) {
	dataLen := len(data)
	if data == nil || dataLen > 16 {
		return U(0), fmt.Errorf("err")
	}

	size := unsafe.Sizeof(U(0))
	dataLenBytes := make([]byte, size)
	_, err := hex.Decode(dataLenBytes, data)
	if err != nil {
		return U(0), err
	}

	ret := U(0)
	intSize := len(dataLenBytes)
	if littleEndian {
		for i := 0; i < intSize; i++ {
			ret |= U(dataLenBytes[i]) << (i * 8)
		}
	} else {
		for i := 0; i < intSize; i++ {
			ret |= U(dataLenBytes[intSize-1-i]) << (i * 8)
		}
	}
	return ret, nil
}

// checksum computes the checksum for a record
func checksum(record []byte) (byte, error) {
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
	return byte(cs), nil
}

func checksumBytes(record []byte) ([]byte, error) {
	hexSum := make([]byte, checksumLen)
	cs, err := checksum(record)
	if err != nil {
		return hexSum, err
	}
	hex.Encode(hexSum, []byte{cs})
	return hexSum, nil
}
