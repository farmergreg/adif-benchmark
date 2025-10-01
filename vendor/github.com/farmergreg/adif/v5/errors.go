package adif

import "errors"

var (
	// ErrAdiReaderMalformedADI is returned when the ADI formatted data does not conform to the ADIF specification.
	ErrAdiReaderMalformedADI = errors.New("adi reader: data is malformed")

	// ErrAdiReaderNilReader is returned when the reader passed to NewADIDocumentReader is nil.
	ErrAdiWriterNilWriter = errors.New("adi writer: nil writer")

	// ErrHeaderAlreadyWritten is returned when attempting to write more than one header record.
	ErrHeaderAlreadyWritten = errors.New("header record already written")
)
