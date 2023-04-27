package hex

import (
	"bytes"
	"errors"
	"reflect"
	"strings"
	"testing"
)

func TestIterator(t *testing.T) {
	test := `:04000000FA00000200
:020000021000EC
:10C20000E0A5E6F6FDFFE0AEE00FE6FCFDFFE6FD93
:10C21000FFFFF6F50EFE4B66F2FA0CFEF2F40EFE90
:10C22000F04EF05FF06CF07DCA0050C2F086F097DF
:10C23000F04AF054BCF5204830592D02E018BB03F9
:020000022000DC
:04000000FA00000200
:00000001FF
`
	file, err := ReadAll(bytes.NewBufferString(test))
	if err != nil {
		t.Errorf("Expected valid hex file got %s", err)
	}

	count := 0
	splitted := strings.Split(test, "\n")
	it := file.Iterator()
	for record := range it {
		currData := record.AsBytes()
		textData := append([]byte(splitted[count]), 13, 10)
		if !bytes.Equal(currData, textData) {
			t.Errorf("Hex record mismatch, expected %v, got %v", textData, currData)
		}
		count++
	}

	if count != len(splitted)-1 { // -1 since splitting the backtick-escaped text generates an empty line
		t.Errorf("Expected %d records, got %d", count, len(splitted))
	}
}

func TestSize(t *testing.T) {
	test := `:04000000FA00000200
:020000021000EC
:10C20000E0A5E6F6FDFFE0AEE00FE6FCFDFFE6FD93
:10C21000FFFFF6F50EFE4B66F2FA0CFEF2F40EFE90
:10C22000F04EF05FF06CF07DCA0050C2F086F097DF
:10C23000F04AF054BCF5204830592D02E018BB03F9
:020000022000DC
:04000000FA00000200
:00000001FF
`
	file, err := ReadAll(bytes.NewBufferString(test))
	if err != nil {
		t.Errorf("Expected valid hex file got %s", err)
	}

	splittedSize := len(strings.Split(test, "\n")) - 1 // -1: same as before
	size := file.Size()

	if size != splittedSize {
		t.Errorf("Expected #records = %d, got %d", splittedSize, size)
	}
}

func TestBinarySize(t *testing.T) {
	test := `:04000000FA00000200
:020000021000EC
:10C20000E0A5E6F6FDFFE0AEE00FE6FCFDFFE6FD93
:10C21000FFFFF6F50EFE4B66F2FA0CFEF2F40EFE90
:10C22000F04EF05FF06CF07DCA0050C2F086F097DF
:10C23000F04AF054BCF5204830592D02E018BB03F9
:020000022000DC
:04000000FA00000200
:00000001FF
`
	testSize := 72

	file, err := ReadAll(bytes.NewBufferString(test))
	if err != nil {
		t.Errorf("Expected valid hex file got %s", err)
	}

	size := file.BinarySize()
	if size != testSize {
		t.Errorf("Expected bin size = %d, got %d", testSize, size)
	}
}

func TestRecord(t *testing.T) {
	test := `:04000000FA00000200
:020000021000EC
:10C20000E0A5E6F6FDFFE0AEE00FE6FCFDFFE6FD93
:10C21000FFFFF6F50EFE4B66F2FA0CFEF2F40EFE90
:10C22000F04EF05FF06CF07DCA0050C2F086F097DF
:10C23000F04AF054BCF5204830592D02E018BB03F9
:020000022000DC
:04000000FA00000200
:00000001FF
`

	file, err := ReadAll(bytes.NewBufferString(test))
	if err != nil {
		t.Errorf("Expected valid hex file got %s", err)
	}

	_, errLower := file.Record(-1)
	if errLower == nil {
		t.Errorf("Expected an out of bounds error for negative indexes")
	}

	_, errUpper := file.Record(100)
	if errUpper == nil {
		t.Errorf("Expected an out of bounds error for indexes >= size")
	}

	splitted := strings.Split(test, "\n")
	recordNum := len(splitted) - 1 // -1: same as before

	for idx := 0; idx < recordNum; idx++ {
		rec, err := file.Record(idx)
		if err != nil {
			t.Errorf("Got an error when getting a record with a valid index: %d", idx)
		}

		if rec.AsString() != splitted[idx] {
			t.Errorf("Expected record '%s', got '%s'", splitted[idx], rec.AsString())
		}
	}
}

func TestReadAll(t *testing.T) {
	tests := []struct {
		input    string
		expected any
	}{
		{"\x01\r\n", WrongRecordFormatErr},
		{`:020000021000EC
:10C20000E0A5E6F6FDFFE0AEE00FE6FCFDFFE6FD93
:10C21000FFFFF6F50EFE4B66F2FA0CFEF2F40EFE90
:10C22000F04EF05FF06CF07DCA0050C2F086F097DF
:10C23000F04AF054BCF5204830592D02E018BB03F9
:020000020000FC
:04000000FA00000200
:00000001FF
`, 8}, {`:020000021000EC
:10C20000E0A5E6F6FDFFE0AEE00FE6FCFDFFE6FD93
:10C21000FFFFF6F50EFE4B66F2FA0CFEF2F40EFE90
:10C22000F04EF05FF06CF07DCA0050C2F086F097DF
:10C23000F04AF054BCF5204830592D02E018BB03F9
:020000020000FC
:04000000FA00000200
`, NoEofRecordErr}, {`:020000021000EC
:10C20000E0A5E6F6FDFFE0AEE00FE6FCFDFFE6FD93
:10C21000FFFFF6F50EFE4B66F2FA0CFEF2F40EFE90
:10C22000F04EF05FF06CF07DCA0050C2F086F097DF
:10C23000F04AF054BCF5204830592D02E018BB03F9
:020000020000FC
:00000001FF
:04000000FA00000200
:00000001FF
`, MultipleEofErr},
		{"", NoEofRecordErr},
	}

	for _, testCase := range tests {
		file, err := ReadAll(bytes.NewBufferString(testCase.input))

		switch expected := testCase.expected.(type) {
		case FileError:
			if !errors.Is(err, expected) {
				t.Errorf("expected %s error, got %s", expected, err)
			}
		case int:
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if len(file.records) != expected {
				t.Fatalf("expected %q records, got %q", expected, len(file.records))
			}
		}
	}
}

func TestFile_ReadAt(t *testing.T) {
	hexFile := `:10000000FFAEAEFF00000000000000000000000096
:04000000FA00000200
:020000021000EC
:10C20000E0A5E6F6FDFFE0AEE00FE6FCFDFFE6FD93
:10C21000FFFFF6F50EFE4B66F2FA0CFEF2F40EFE90
:10C22000F04EF05FF06CF07DCA0050C2F086F097DF
:10C23000F04AF054BCF5204830592D02E018BB03F9
:020000022000DC
:04000000FA00000200
:00000001FF
`
	tests := []struct {
		pos      uint32
		size     int
		expected any
	}{
		{0, 21, AccessOutOfBounds},
		{1, 3, []byte{0xAE, 0xAE, 0xFF}},
		{0, 4, []byte{0xFF, 0xAE, 0xAE, 0xFF}},
		{0, 16, []byte{0xFF, 0xAE, 0xAE, 0xFF, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}},
		{
			0x1000*16 + 0xC200,
			16,
			[]byte{0xE0, 0xA5, 0xE6, 0xF6, 0xFD, 0xFF, 0xE0, 0xAE, 0xE0, 0x0F, 0xE6, 0xFC, 0xFD, 0xFF, 0xE6, 0xFD},
		},
		{
			0x1000*16 + 0xC200,
			14,
			[]byte{0xE0, 0xA5, 0xE6, 0xF6, 0xFD, 0xFF, 0xE0, 0xAE, 0xE0, 0x0F, 0xE6, 0xFC, 0xFD, 0xFF},
		},
		{
			0x1000*16 + 0xC200,
			18,
			[]byte{0xE0, 0xA5, 0xE6, 0xF6, 0xFD, 0xFF, 0xE0, 0xAE, 0xE0, 0x0F, 0xE6, 0xFC, 0xFD, 0xFF, 0xE6, 0xFD, 0xFF, 0xFF},
		},
		{
			0x1000*16 + 0xC202,
			16,
			[]byte{0xE6, 0xF6, 0xFD, 0xFF, 0xE0, 0xAE, 0xE0, 0x0F, 0xE6, 0xFC, 0xFD, 0xFF, 0xE6, 0xFD, 0xFF, 0xFF},
		},
		{
			0x2000*16 - 2,
			4,
			AccessOutOfBounds,
		},
		{
			0x2000 * 16,
			6,
			AccessOutOfBounds,
		},
		{
			0x2000 * 16,
			4,
			[]byte{0xFA, 0x00, 0x00, 0x02},
		},
	}

	file, _ := ReadAll(bytes.NewBufferString(hexFile))

	for _, testCase := range tests {
		readData, err := file.ReadAt(testCase.pos, testCase.size)

		switch expected := testCase.expected.(type) {
		case FileError:
			if !errors.Is(err, expected) {
				t.Errorf("expected %s error, got %s", expected, err)
			}
		case []byte:
			if err != nil {
				t.Fatalf("unexpected error: %s, expecting %v", err, expected)
			}

			if !reflect.DeepEqual(readData, expected) {
				t.Errorf("expected read data to be %v, got %v", expected, readData)
			}
		}
	}
}

func TestFile_WriteAt(t *testing.T) {

	hexFile := `:04000000FA00000200
:020000021000EC
:10C20000E0A5E6F6FDFFE0AEE00FE6FCFDFFE6FD93
:10C21000FFFFF6F50EFE4B66F2FA0CFEF2F40EFE90
:10C22000F04EF05FF06CF07DCA0050C2F086F097DF
:10C23000F04AF054BCF5204830592D02E018BB03F9
:020000022000DC
:04000000FA00000200
:00000001FF
`
	tests := []struct {
		pos           uint32
		input         []byte
		expectedError error
	}{
		{0, []byte{}, nil},
		{0, []byte{0x0C, 0xAF, 0xFE}, nil},
		{1, []byte{0xAF, 0xFE}, nil},
		{
			0x1000*16 + 0xC200,
			[]byte{0x00, 0xEE, 0xAE, 0xBC, 0x01, 0x02, 0x03, 0x04, 0xCC, 0x05, 0x60, 0x71, 0x44, 0x12, 0xF7, 0xA1},
			nil,
		},
		{
			0x1000*16 + 0xC200,
			[]byte{0xAA, 0xBD, 0x1C},
			nil,
		},
		{
			0x1000*16 + 0xC200,
			[]byte{0x00, 0xEE, 0xAE, 0xBC, 0x01, 0x02, 0x03, 0x04, 0xCC, 0x05, 0x60, 0x71, 0x44, 0x12, 0xF7, 0xA1, 0xFF, 0xFD},
			nil,
		},
		{
			0x1000*16 + 0xC202,
			[]byte{0x00, 0xEE, 0xAE, 0xBC, 0x01, 0x02, 0x03, 0x04, 0xCC, 0x05, 0x60, 0x71, 0x44, 0x12, 0xF7, 0xA1},
			nil,
		},
		{
			0x1000*16 + 0xC202,
			[]byte{0x00, 0xEE, 0xAE, 0xBC, 0x01, 0x02, 0x03, 0x04, 0xCC, 0x05, 0x60, 0x71, 0x44, 0x12, 0xF7, 0xA1, 0x01, 0x09, 0x21, 0x23},
			nil,
		},
		{
			0x2000*16 - 2,
			[]byte{0xAA, 0xBD, 0x1C, 0x2C},
			AccessOutOfBounds,
		},
		{
			0x2000 * 16,
			[]byte{0xAA, 0xBD, 0x1C, 0x2C, 0x00, 0xFE},
			AccessOutOfBounds,
		},
		{
			0x2000 * 16,
			[]byte{0xAA, 0xBD, 0x1C, 0x2C},
			nil,
		},
	}

	for _, testCase := range tests {
		file, _ := ReadAll(bytes.NewBufferString(hexFile))
		err := file.WriteAt(testCase.pos, testCase.input)

		switch testCase.expectedError {
		case AccessOutOfBounds:
			if !errors.Is(err, testCase.expectedError) {
				t.Errorf("expected %q error, got %v", testCase.expectedError, err)
			}
		case nil:
			if err != nil {
				t.Fatalf("unexpected error: %s, expecting %v", err, testCase.expectedError)
			}

			readData, err := file.ReadAt(testCase.pos, len(testCase.input))
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			for _, record := range file.records {
				if isValid, _, _ := validateRecord(record); !isValid {
					t.Fatalf("invalid record after write: %s", record.AsString())
				}
			}

			if !reflect.DeepEqual(readData, testCase.input) {
				t.Errorf("expected read data @%X to be %v, got %v", testCase.pos, testCase.input, readData)
			}
		}
	}
}
