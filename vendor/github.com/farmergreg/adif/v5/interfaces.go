package adif

import "github.com/farmergreg/spec/v6/adifield"

// Record represents a single ADIF record.
// It may be either a Header, or a QSO.
type Record interface {
	IsHeader() bool            // IsHeader returns true if the record is a header record.
	SetIsHeader(isHeader bool) // SetIsHeader sets whether the record is a header record.

	Get(field adifield.Field) string        // Get returns the value for the specified field, or an empty string if the field is not present.
	Set(field adifield.Field, value string) // Set sets the value for the specified field.

	All() func(func(adifield.Field, string) bool) // All returns an iterator that yields field-value pairs for all fields in the record.
	Count() int                                   // Count returns the number of fields in the record.
}

// ADIFRecordReader reads Amateur Data Interchange Format (ADIF) records sequentially.
type ADIFRecordReader interface {

	// Next reads and returns the next Record in the input.
	// It returns io.EOF when no more records are available.
	Next() (record Record, err error)
}

// ADIFRecordWriter writes Amateur Data Interchange Format (ADIF) records sequentially.
type ADIFRecordWriter interface {
	// Write writes ADIF record(s) to the output.
	Write(record Record) error
}
