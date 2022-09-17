package bytes

import "fmt"

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
	AccessOutOfBounds = FileError("cannot access the hex file out of the length of the encoded program")
)
