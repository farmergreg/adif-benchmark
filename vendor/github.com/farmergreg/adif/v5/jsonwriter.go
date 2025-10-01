package adif

import (
	"encoding/json"
	"io"
	"maps"
)

var _ DocumentWriter = (*jsonWriter)(nil)

// jsonWriter implements ADIFDocumentWriter for writing ADIF records in ADIJ format.
type jsonWriter struct {
	w      io.Writer
	doc    *jsonDocument
	indent string
}

// NewJSONDocumentWriter creates a new ADIFDocumentWriter that writes ADIJ JSON to the provided io.Writer.
// The indent parameter specifies the string to use for indentation (e.g. "\t" or "  ").
// An empty string means no indentation.
// JSON is not an official ADIF document container format.
// It is, however, useful for interoperability with other systems.
func NewJSONDocumentWriter(w io.Writer, indent string) DocumentWriter {
	return &jsonWriter{
		w:      w,
		doc:    &jsonDocument{},
		indent: indent,
	}
}

// WriteHeader implements ADIFDocumentWriter.WriteHeader for writing ADIF headers in ADIJ format.
func (j *jsonWriter) WriteHeader(record Record) error {
	if j.doc.Header != nil || len(j.doc.Records) > 0 {
		return ErrHeaderAlreadyWritten
	}
	j.doc.Header = maps.Collect(record.All())
	return nil
}

// WriteRecord implements ADIFDocumentWriter.WriteRecord for writing ADIF records in ADIJ format.
func (j *jsonWriter) WriteRecord(record Record) error {
	r := maps.Collect(record.All())
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
