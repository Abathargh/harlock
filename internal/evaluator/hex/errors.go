package hex

import "fmt"

// RecordError identifies an error related to a hex record
type RecordError string

// Error returns a string representation of a RecordError
func (r RecordError) Error() string {
	return string(r)
}

const (
	MissingStartCodeErr  = RecordError("the passed record does not start with the correct start code")
	WrongRecordFormatErr = RecordError("the passed record is not a correct hex record")
	DataOutOfBounds      = RecordError("the passed byte slice cannot be held by this record")
	InvalidHexDigit      = RecordError("the passed byte represents a character that is not an hex digit")
	NoMoreRecordsErr     = RecordError("no more records")
)

// FileError identifies an error related to a hex file
type FileError string

// Error returns a string representation of a FileError
func (r FileError) Error() string {
	return string(r)
}

// CustomError returns FileError that can use the classic fmt message/varargs.
func CustomError(original FileError, msg string, args ...any) error {
	nested := fmt.Sprintf(msg, args...)
	return fmt.Errorf("%w: %s", original, nested)
}

const (
	MultipleEofErr    = FileError("the passed hex file contains more than one EOF records")
	NoEofRecordErr    = FileError("the passed hex file does not contain an EOF record")
	AccessOutOfBounds = FileError("cannot access the hex file out of the length of the encoded program")
	RecordErr         = FileError("faulty record")
)
