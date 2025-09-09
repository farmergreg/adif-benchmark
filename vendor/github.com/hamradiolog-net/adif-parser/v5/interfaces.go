package adif

import "github.com/hamradiolog-net/adif-spec/v6/adifield"

// ADIFRecord represents a single ADIF record, which may be either a header or a QSO record.
type ADIFRecord interface {
	IsHeader() bool                            // IsHeader returns true if the record is a header record.
	SetIsHeader(isHeader bool)                 // SetIsHeader sets whether the record is a header record.
	Get(field adifield.ADIField) string        // Get returns the value for the specified field, or an empty string if the field is not present.
	Set(field adifield.ADIField, value string) // Set sets the value for the specified field.
	Fields() []adifield.ADIField               // Fields returns a slice of all field available in the record.
}

// ADIFRecordReader processes Amateur Data Interchange Format (ADIF) records sequentially.
type ADIFRecordReader interface {

	// Next reads and returns the next Record in the input.
	// It returns io.EOF when no more records are available.
	Next() (record ADIFRecord, err error)
}

// ADIFRecordWriter writes Amateur Data Interchange Format (ADIF) records sequentially.
type ADIFRecordWriter interface {
	// Write writes ADIF record(s) to the output.
	Write(record []ADIFRecord) error
}
