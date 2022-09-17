package bytes

import (
	"bytes"
	"errors"
	"testing"
)

func TestFile_WriteAt(t *testing.T) {
	tests := []struct {
		input        []byte
		data         []byte
		position     int
		expectedErr  error
		expectedRead []byte
	}{
		{[]byte{1, 2, 3, 4}, []byte{6, 7, 8, 9}, 0, nil, []byte{6, 7, 8, 9}},
		{[]byte{1, 2, 3, 4}, []byte{6, 7}, 2, nil, []byte{1, 2, 6, 7}},
		{[]byte{1, 2, 3, 4}, []byte{6}, 3, nil, []byte{1, 2, 3, 6}},
		{[]byte{1, 2, 3, 4}, []byte{6, 7, 8, 9, 10}, 0, AccessOutOfBounds, nil},
		{[]byte{1, 2, 3, 4}, []byte{6, 7, 8, 9, 10}, 2, AccessOutOfBounds, nil},
		{[]byte{1, 2, 3, 4}, []byte{6, 7, 8, 9, 10}, 3, AccessOutOfBounds, nil},
		{[]byte{1, 2, 3, 4}, []byte{6, 7, 8, 9, 10}, 4, AccessOutOfBounds, nil},
	}

	for idx, testCase := range tests {
		bytesFile, err := ReadAll(bytes.NewReader(testCase.input))
		if err != nil {
			t.Errorf("unexpected error, got %v for case '%d'", err, idx)
			continue
		}

		werr := bytesFile.WriteAt(testCase.position, testCase.data)
		switch testCase.expectedErr {
		case AccessOutOfBounds:
			if !errors.Is(werr, testCase.expectedErr) {
				t.Errorf("expected err %q got %v", testCase.expectedErr, werr)
			}
		case nil:
			if werr != nil {
				t.Errorf("unexpected err %v", werr)
				continue
			}

			if !bytes.Equal(bytesFile.bytes, testCase.expectedRead) {
				t.Errorf("unexpected data after write: got %v, expected %v", bytesFile.bytes, testCase.expectedRead)
			}
		}
	}
}

func TestFile_ReadAt(t *testing.T) {
	tests := []struct {
		input        []byte
		position     int
		size         int
		expectedErr  error
		expectedRead []byte
	}{
		{[]byte{1, 2, 3, 4}, 0, 4, nil, []byte{1, 2, 3, 4}},
		{[]byte{0xca, 0xff, 0xe0, 0xaa, 0xa1, 0xa2}, 2, 2, nil, []byte{0xe0, 0xaa}},
		{[]byte{0xca, 0xff, 0xe0, 0xaa, 0xa1, 0xa2}, 5, 1, nil, []byte{0xa2}},
		{[]byte{0xca, 0xff, 0xe0}, 0, 4, AccessOutOfBounds, nil},
		{[]byte{0xca, 0xff, 0xe0}, 1, 3, AccessOutOfBounds, nil},
		{[]byte{0xca, 0xff, 0xe0}, 2, 3, AccessOutOfBounds, nil},
		{[]byte{0xca, 0xff, 0xe0}, 3, 3, AccessOutOfBounds, nil},
	}

	for idx, testCase := range tests {
		bytesFile, err := ReadAll(bytes.NewReader(testCase.input))
		if err != nil {
			t.Errorf("unexpected error, got %v for case '%d'", err, idx)
			continue
		}

		readData, rerr := bytesFile.ReadAt(testCase.position, testCase.size)
		switch testCase.expectedErr {
		case AccessOutOfBounds:
			if !errors.Is(rerr, testCase.expectedErr) {
				t.Errorf("expected err %q got %v", testCase.expectedErr, rerr)
			}
		case nil:
			if rerr != nil {
				t.Errorf("unexpected err %v", rerr)
				continue
			}

			if !bytes.Equal(readData, testCase.expectedRead) {
				t.Errorf("unexpected data after write: got %v, expected %v", readData, testCase.expectedRead)
			}
		}
	}
}
