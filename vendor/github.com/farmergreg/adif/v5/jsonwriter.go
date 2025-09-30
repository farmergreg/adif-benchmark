package adif

import (
	"encoding/json"
	"io"
	"maps"
)

var _ RecordWriteFlusher = (*jsonWriter)(nil)

// jsonWriter implements ADIFRecordWriter for writing ADIF records in ADIJ format.
type jsonWriter struct {
	w      io.Writer
	doc    *jsonDocument
	indent string
}

// NewJSONRecordWriter creates a new ADIFRecordWriter that writes ADIJ JSON to the provided io.Writer.
// The indent parameter specifies the string to use for indentation (e.g. "\t" or "  ").
// An empty string means no indentation.
// JSON is not an official ADIF document container format.
// It is, however, useful for interoperability with other systems.
func NewJSONRecordWriter(w io.Writer, indent string) RecordWriteFlusher {
	return &jsonWriter{
		w:      w,
		doc:    &jsonDocument{},
		indent: indent,
	}
}

// Write implements ADIFRecordWriter.Write for writing ADIF records in ADIJ format.
func (j *jsonWriter) Write(record Record, isHeader bool) error {
	r := maps.Collect(record.All())
	if isHeader {
		if j.doc.Header != nil {
			return ErrHeaderAlreadyWritten
		}
		j.doc.Header = r
		return nil
	}
	j.doc.Records = append(j.doc.Records, r)
	return nil
}

// Flush implements RecordWriteFlusher.Flush
func (j *jsonWriter) Flush() error {
	encoder := json.NewEncoder(j.w)
	encoder.SetIndent("", j.indent)
	err := encoder.Encode(j.doc)
	if err != nil {
		return err
	}
	return nil
}
