package hex

import (
	"bytes"
	"errors"
	"reflect"
	"testing"
)

func TestParseRecord(t *testing.T) {
	tests := []struct {
		input    string
		expected any
	}{
		{"", NoMoreRecordsErr},
		{"\r\n", MissingStartCodeErr},
		{":\r\n", WrongRecordFormatErr},
		{`:\r`, WrongRecordFormatErr},
		{`:\r`, WrongRecordFormatErr},
		{`:0001\r\n`, WrongRecordFormatErr},
		{`:0001\r\n`, WrongRecordFormatErr},
		{`:00\r\n`, WrongRecordFormatErr},
		{":020000021000EC\r\n", &Record{
			length: 2,
			rType:  ExtendedSegmentAddrRecord,
			data:   []byte{':', '0', '2', '0', '0', '0', '0', '0', '2', '1', '0', '0', '0', 'E', 'C'},
		}},
		{":06058000000A000000006B\r\n", &Record{
			length: 6,
			rType:  DataRecord,
			data:   []byte{':', '0', '6', '0', '5', '8', '0', '0', '0', '0', '0', '0', 'A', '0', '0', '0', '0', '0', '0', '0', '0', '6', 'B'},
		}},
		{":00000001FF\r\n", &Record{
			length: 0,
			rType:  EOFRecord,
			data:   []byte{':', '0', '0', '0', '0', '0', '0', '0', '1', 'F', 'F'},
		}},
	}

	for _, testCase := range tests {
		rec, err := ParseRecord(bytes.NewBufferString(testCase.input))
		switch expected := testCase.expected.(type) {
		case RecordError:
			if !errors.Is(err, expected) {
				t.Errorf("expected %q error, got %q", expected, err)
			}
		case *Record:
			testRecordEqual(t, rec, expected)
		}
	}
}

func TestAsString(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{":020000021000EC\r\n", ":020000021000EC"},
		{":06058000000A000000006B\r\n", ":06058000000A000000006B"},
		{":00000001FF\r\n", ":00000001FF"},
	}
	empty := &Record{}
	if empty.AsString() != "" {
		t.Errorf("expected non initialized record to have an empty string repr, got %s", empty.AsString())
	}

	for _, testCase := range tests {
		rec, err := ParseRecord(bytes.NewBufferString(testCase.input))
		if err != nil {

		}
		strRep := rec.AsString()
		if strRep != testCase.expected {
			t.Errorf("expected str repr = %q, got %q", testCase.expected, strRep)
		}
	}
}

func TestByteCount(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{":020000021000EC\r\n", 2},
		{":06058000000A000000006B\r\n", 6},
		{":00000001FF\r\n", 0},
	}
	empty := &Record{}
	if empty.ByteCount() != 0 {
		t.Errorf("expected non initialized record to have len 0, got %d", empty.ByteCount())
	}

	for _, testCase := range tests {
		rec, err := ParseRecord(bytes.NewBufferString(testCase.input))
		if err != nil {

		}
		count := rec.ByteCount()
		if count != testCase.expected {
			t.Errorf("expected byte count = %d, got %d", testCase.expected, count)
		}
	}
}

func TestAddress(t *testing.T) {
	tests := []struct {
		input         string
		expected      uint16
		expectedBytes []byte
	}{
		{":020000021000EC\r\n", 0, []byte{'0', '0', '0', '0'}},
		{":06058000000A000000006B\r\n", 1408, []byte{'0', '5', '8', '0'}},
		{":00000001FF\r\n", 0, []byte{'0', '0', '0', '0'}},
	}
	empty := &Record{}

	if empty.Address() != 0 {
		t.Errorf("expected non initialized record to have addr 0, got %d", empty.Address())
	}

	if empty.AddressBytes() != nil {
		t.Errorf("expected non initialized record to return nil, got %d", empty.AddressBytes())
	}

	for _, testCase := range tests {
		rec, err := ParseRecord(bytes.NewBufferString(testCase.input))
		if err != nil {

		}
		addr := rec.Address()
		if addr != testCase.expected {
			t.Errorf("expected addr = %d, got %d", testCase.expected, addr)
		}

		addrBytes := rec.AddressBytes()
		if !reflect.DeepEqual(addrBytes, testCase.expectedBytes) {
			t.Errorf("expected addr = %v, got %v", testCase.expectedBytes, addrBytes)
		}
	}
}

func TestType(t *testing.T) {
	tests := []struct {
		input    string
		expected RecordType
	}{
		{":020000021000EC\r\n", ExtendedSegmentAddrRecord},
		{":06058000000A000000006B\r\n", DataRecord},
		{":00000001FF\r\n", EOFRecord},
	}
	empty := &Record{}

	if empty.Type() != InvalidRecord {
		t.Errorf("expected non initialized record to have InvalidType, got %v", empty.Type())
	}

	for _, testCase := range tests {
		rec, _ := ParseRecord(bytes.NewBufferString(testCase.input))
		rType := rec.Type()
		if rType != testCase.expected {
			t.Errorf("expected type = %d, got %d", testCase.expected, rType)
		}
	}
}
func TestReadData(t *testing.T) {
	tests := []struct {
		input    string
		expected []byte
	}{
		{":020000021000EC\r\n", []byte{'1', '0', '0', '0'}},
		{":06058000000A000000006B\r\n", []byte{'0', '0', '0', 'A', '0', '0', '0', '0', '0', '0', '0', '0'}},
		{":00000001FF\r\n", []byte{}},
	}
	empty := &Record{}

	if empty.ReadData() != nil {
		t.Errorf("expected non initialized record to have nil data, got %v", empty.ReadData())
	}

	for _, testCase := range tests {
		rec, _ := ParseRecord(bytes.NewBufferString(testCase.input))
		data := rec.ReadData()
		if !reflect.DeepEqual(data, testCase.expected) {
			t.Errorf("expected data = %v, got %v", testCase.expected, data)
		}
	}
}

func TestChecksum(t *testing.T) {
	tests := []struct {
		input    string
		expected []byte
	}{
		{":020000021000EC\r\n", []byte{'E', 'C'}},
		{":06058000000A000000006B\r\n", []byte{'6', 'B'}},
		{":00000001FF\r\n", []byte{'F', 'F'}},
	}
	empty := &Record{}

	if empty.Checksum() != nil {
		t.Errorf("expected non initialized record to have nil checksum, got %d", empty.Checksum())
	}

	for _, testCase := range tests {
		rec, _ := ParseRecord(bytes.NewBufferString(testCase.input))
		cSum := rec.Checksum()
		if !reflect.DeepEqual(cSum, testCase.expected) {
			t.Errorf("expected checksum = %v, got %v", testCase.expected, cSum)
		}
	}
}

func TestWriteData(t *testing.T) {
	tests := []struct {
		input    string
		start    int
		wData    []byte
		expected any
	}{
		{":020000021000EC\r\n", -1, nil, DataOutOfBounds},
		{":020000021000EC\r\n", 5, []byte{}, DataOutOfBounds},
		{":020000021000EC\r\n", 0, []byte{'\n', ':'}, InvalidHexDigit},
		{":020000021000EC\r\n", 0, []byte{'1', '0', '0', '0', '0'}, DataOutOfBounds},
		{":020000021000EC\r\n", 0, []byte{'3', '4', '5', '6'}, []byte{'3', '4', '5', '6'}},
		{":06058000000A000000006B\r\n", 2, []byte{'A', 'E'}, []byte{'0', '0', 'A', 'E', '0', '0', '0', '0', '0', '0', '0', '0'}},
		{":00000001FF\r\n", 0, []byte{}, []byte{}},
	}
	empty := &Record{}

	if err := empty.WriteData(0, []byte{}); err != DataOutOfBounds {
		t.Errorf("expected non initialized record to return OutOfBound on write, got %v", err)
	}

	for _, testCase := range tests {
		rec, _ := ParseRecord(bytes.NewBufferString(testCase.input))
		err := rec.WriteData(testCase.start, testCase.wData)
		switch expected := testCase.expected.(type) {
		case RecordError:
			if !errors.Is(err, expected) {
				t.Errorf("expected %q error, got %q", expected, err)
			}
		case []byte:
			if !reflect.DeepEqual(rec.ReadData(), expected) {
				t.Fatalf("expected data = %v, got %v", expected, rec.ReadData())
			}

			okSum, _ := checksum(rec.data)
			actualSum, _ := hexToInt[uint8](rec.Checksum(), true)
			if okSum != actualSum {
				t.Errorf("expected new checksum = %d, got %d", okSum, actualSum)
			}
		}
	}
}

func testRecordEqual(t *testing.T, rec, expected *Record) {
	if rec == nil || rec.length != expected.length ||
		rec.rType != expected.rType ||
		!reflect.DeepEqual(rec, expected) {
		t.Errorf("expected %+v, got %+v", expected, rec)
	}
}
