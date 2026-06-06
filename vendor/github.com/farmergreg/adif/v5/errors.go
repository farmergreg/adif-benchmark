package adif

import "errors"

var (
	// ErrMalformedADI is returned when the ADI formatted data does not conform to the ADIF specification.
	ErrMalformedADI = errors.New("malformed ADI")

	// ErrTooManyUniqueFields is returned when the number of unique field names exceeds the internal limit.
	// This prevents denial of service attacks from malformed ADI files with unlimited unique field names.
	ErrTooManyUniqueFields = errors.New("too many unique field names")

	// ErrUnexpectedHeader is returned when a header is encountered after QSO records have already been read, or after a header has already been processed.
	ErrUnexpectedHeader = errors.New("unexpected header")

	// ErrHeaderAlreadyWritten is returned when attempting to write more than one header record.
	ErrHeaderAlreadyWritten = errors.New("header already written")
)
