package adif

import (
	"github.com/farmergreg/spec/v6/adifield"
)

// Record represents a single ADIF record.
// It represents both Header and QSO records.
type Record interface {
	Get(field adifield.Field) string        // Get returns the value for the specified field, or an empty string if the field is not present.
	Set(field adifield.Field, value string) // Set sets the value for the specified field.

	Fields() func(func(adifield.Field, string) bool) // Fields returns an iterator that yields field-value pairs for all fields in the record.
	FieldCount() int                                 // FieldCount returns the number of fields in the record.
}

// DocumentReader reads Amateur Data Interchange Format (ADIF) records sequentially.
type DocumentReader interface {
	// Next reads and returns the next Record.
	// It returns io.EOF when no more records are available.
	// isHeader indicates if the record is a header record.
	Next() (record Record, isHeader bool, err error)
}

// DocumentWriter writes Amateur Data Interchange Format (ADIF) records sequentially.
type DocumentWriter interface {
	// WriteHeader writes the ADIF header record to the output.
	// It MUST be called before using WriteRecord.
	// If WriteHeader is called more than once, it returns an error.
	WriteHeader(record Record) error

	// WriteRecord writes ADIF record(s) to the output.
	// When writing a header record, it MUST be the first record written.
	WriteRecord(record Record) error

	// Flush writes buffered data to the underlying writer.
	// IMPORTANT: This MUST be called once the header and records have been written to ensure all data is properly written.
	Flush() error
}
