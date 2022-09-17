package elf

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
	FileOpenErr      = FileError("cannot open the file with the passed file name")
	NoSuchSectionErr = FileError("there is no such section in the passed elf file")
	OutOfBoundsErr   = FileError("attempting to write out of the section bounds")
)
