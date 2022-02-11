package hex

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
