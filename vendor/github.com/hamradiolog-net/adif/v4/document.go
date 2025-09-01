package adif

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

var (
	_ io.WriterTo   = &Document{}
	_ io.ReaderFrom = &Document{}
	_ fmt.Stringer  = &Document{}
)

// NewDocument creates a new Document with default initial capacity.
func NewDocument() *Document {
	return NewDocumentWithOptions(0, "")
}

// NewDocumentWithOptions creates a new Document with the specified capacity and header preamble.
// If capacity is less than 1, it defaults to 32.
// If headerPreamble is empty, it defaults to the standard ADIF header preamble.
func NewDocumentWithOptions(capacity int, headerPreamble string) *Document {
	if capacity < 1 {
		capacity = 32
	}

	return &Document{
		Records:        make([]Record, 0, capacity),
		headerPreamble: headerPreamble,
	}
}

// WriteTo writes the document in ADI format to the given writer.
func (f *Document) WriteTo(w io.Writer) (n int64, err error) {
	bw, ok := w.(*bufio.Writer)
	if !ok {
		bw = bufio.NewWriter(w)
	}

	if len(f.Header) > 0 {
		if f.headerPreamble == "" {
			f.headerPreamble = adiHeaderPreamble
		}
		ci, err := bw.WriteString(f.headerPreamble)
		n += int64(ci)
		if err != nil {
			return handleFlush(bw, n, err)
		}

		var c64 int64
		c64, err = f.Header.WriteTo(bw)
		n += c64
		if err != nil {
			return handleFlush(bw, n, err)
		}

		ci, err = bw.WriteString(TagEOH + "\n")
		n += int64(ci)
		if err != nil {
			return handleFlush(bw, n, err)
		}
	}

	for _, record := range f.Records {
		c, err := record.WriteTo(bw)
		n += c
		if err != nil {
			return handleFlush(bw, n, err)
		}

		cc, err := bw.WriteString(TagEOR + "\n")
		n += int64(cc)
		if err != nil {
			return handleFlush(bw, n, err)
		}
	}

	return handleFlush(bw, n, err)
}

func handleFlush(bw *bufio.Writer, n int64, err error) (int64, error) {
	if bwerr := bw.Flush(); bwerr != nil {
		return n, fmt.Errorf("error flushing writer: %w", bwerr)
	}
	return n, err
}

// ReadFrom reads an ADI document from the given reader.
func (f *Document) ReadFrom(r io.Reader) (n int64, err error) {
	p := NewADIReader(r, false)

	firstRecord, isHeader, n, err := p.Next()
	if err != nil {
		return n, err
	}

	if isHeader {
		f.Header = firstRecord
	} else {
		f.Records = append(f.Records, firstRecord)
	}

	for {
		record, _, c, err := p.Next()
		n += c
		if err != nil {
			if err == io.EOF {
				break
			}
			return n, err
		}
		f.Records = append(f.Records, record)
	}

	return n, nil
}

// String returns the document as an ADI string.
// Returns an empty string if the receiver is nil.
func (f *Document) String() string {
	if f == nil || len(f.Records) == 0 {
		return ""
	}

	sb := strings.Builder{}

	_, err := f.WriteTo(&sb)
	if err != nil {
		return fmt.Sprintf("error while building adi string: %v", err)
	}

	return sb.String()
}
