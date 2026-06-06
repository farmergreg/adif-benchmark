package adif

import (
	"io"
	"strings"
)

// Document holds a complete ADIF document in memory.
// Header is nil when the source contains no header record.
//
// For large files that should not be fully loaded into memory, use Scanner instead.
type Document struct {
	// Header is the ADIF header record, or nil if no header is present.
	Header Record `json:"HEADER,omitempty"`

	// Records contains all QSO records.
	Records []Record `json:"RECORDS"`
}

// NewDocument returns an empty Document.
func NewDocument() *Document {
	return &Document{
		Records: make([]Record, 0, 256),
	}
}

// ReadFrom reads an ADI document from r, appending its records to this Document.
// Implements io.ReaderFrom.
func (d *Document) ReadFrom(r io.Reader) (int64, error) {
	cr := &countingReader{r: r}
	s := NewScanner(cr)
	for s.Scan() {
		if s.IsHeader() {
			if d.Header != nil || len(d.Records) > 0 {
				return cr.n, ErrUnexpectedHeader
			}
			d.Header = s.Record()
		} else {
			d.Records = append(d.Records, s.Record())
		}
	}
	return cr.n, s.Err()
}

// WriteTo writes the document in ADI format to w.
// Implements io.WriterTo.
func (d *Document) WriteTo(w io.Writer) (int64, error) {
	cw := &countingWriter{w: w}
	wr := NewWriter(cw)
	if d.Header != nil {
		if err := wr.WriteHeader(d.Header); err != nil {
			return cw.n, err
		}
	}
	for _, record := range d.Records {
		if err := wr.Write(record); err != nil {
			return cw.n, err
		}
	}
	// Writer adds no buffering of its own; flush the original writer directly.
	if err := flushWriter(w); err != nil {
		return cw.n, err
	}
	return cw.n, nil
}

// String returns the document serialized as an ADI string.
// Returns an empty string when the document has no header and no records.
// Implements fmt.Stringer.
func (d *Document) String() string {
	if d == nil || (d.Header == nil && len(d.Records) == 0) {
		return ""
	}
	var sb strings.Builder
	_, _ = d.WriteTo(&sb)
	return sb.String()
}

// countingReader wraps an io.Reader and tracks the total bytes read.
type countingReader struct {
	r io.Reader
	n int64
}

func (cr *countingReader) Read(p []byte) (n int, err error) {
	n, err = cr.r.Read(p)
	cr.n += int64(n)
	return n, err
}

// countingWriter wraps an io.Writer and tracks the total bytes written.
type countingWriter struct {
	w io.Writer
	n int64
}

func (cw *countingWriter) Write(p []byte) (n int, err error) {
	n, err = cw.w.Write(p)
	cw.n += int64(n)
	return n, err
}
