package adif

import "errors"

var (
	// ErrAdiReaderMalformedADI is returned when the ADI formatted data does not conform to the ADIF specification.
	ErrAdiReaderMalformedADI = errors.New("adi reader: data is malformed")

	// ErrAdiReaderInvalidFieldLength is returned when the field length is invalid, or would cause a large memory allocation.
	ErrAdiReaderInvalidFieldLength = errors.New("adi reader: invalid field length")

	// ErrAdiReaderNilReader is returned when the reader passed to NewADIRecordReader is nil.
	ErrAdiWriterNilWriter = errors.New("adi writer: nil writer")
)
