package adif

import (
	"github.com/farmergreg/spec/v6/adifield"
)

// Record represents a single ADIF record.
// It may be either a Header, or a QSO.
type Record interface {
	Get(field adifield.Field) string        // Get returns the value for the specified field, or an empty string if the field is not present.
	Set(field adifield.Field, value string) // Set sets the value for the specified field.

	All() func(func(adifield.Field, string) bool) // All returns an iterator that yields field-value pairs for all fields in the record.
	Count() int                                   // Count returns the number of fields in the record.
}

// RecordReader reads Amateur Data Interchange Format (ADIF) records sequentially.
type RecordReader interface {

	// Next reads and returns the next Record in the input.
	// It returns io.EOF when no more records are available.
	// isHeader indicates if the record is a header record.
	Next() (record Record, isHeader bool, err error)
}

// RecordWriter writes Amateur Data Interchange Format (ADIF) records sequentially.
type RecordWriter interface {
	// Write writes ADIF record(s) to the output.
	// isHeader indicates if the record is a header record.
	// When writing a header record, it MUST be the first record written.
	Write(record Record, isHeader bool) error
}

// RecordWriteFlusher writes Amateur Data Interchange Format (ADIF) records sequentially.
// When all records are written, Close() MUST be called.
type RecordWriteFlusher interface {
	RecordWriter
	Flush() error // Flush flushes buffered data to the underlying writer.
}
