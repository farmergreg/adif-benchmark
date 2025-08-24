package adif

import "errors"

var (
	// ErrMalformedADI is returned when the ADI formatted data does not conform to the ADIF specification.
	ErrMalformedADI = errors.New("adi reader: data is malformed")

	// ErrInvalidFieldLength is returned when the field length is invalid, or would cause a large memory allocation.
	ErrInvalidFieldLength = errors.New("adi reader: invalid field length")
)
