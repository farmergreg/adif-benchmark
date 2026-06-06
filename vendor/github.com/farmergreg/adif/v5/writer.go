package adif

import (
	"io"
	"slices"
	"strconv"
	"sync"

	"github.com/farmergreg/spec/v6/adifield"
)

// WriteMode controls the output format of the ADI Writer.
type WriteMode int

const (
	// WriteModePretty is the default write mode.
	// It produces human-readable output with common ADIF fields presented first followed by the remaining fields in alphabetical order.
	WriteModePretty WriteMode = iota

	// WriteModeFast is optimized for speed and does not guarantee field order.
	WriteModeFast
)

const adiHeaderPreamble = "                    AM✠DG\nK9CTS High Performance ADIF Processing Library\n   https://github.com/farmergreg/adif\n\n"

// adiWriterPriorityFieldOrder defines the order in which common QSO fields are written.
// These fields appear first in every record; all others follow in alphabetical order.
// This matches the minimum required fields from the ADIF spec and LoTW's submission list.
var adiWriterPriorityFieldOrder = [...]adifield.Field{
	// Minimum required fields per https://www.adif.org/315/ADIF_315_Resources.htm#ADIFImplementationNotesMinimumFields
	adifield.QSO_DATE,
	adifield.TIME_ON,
	adifield.BAND,
	adifield.MODE,
	adifield.SUBMODE, // not strictly required but belongs with MODE
	adifield.CALL,    // last so adjacent records are vertically aligned in the file

	// LoTW submission fields per https://lotw.arrl.org/lotw-help/developer-submit-qsos/
	adifield.FREQ,
	adifield.FREQ_RX,
	adifield.BAND_RX,
	adifield.PROP_MODE,
	adifield.SAT_NAME,
	adifield.STATION_CALLSIGN,
	adifield.OPERATOR,
	adifield.MY_DXCC,
	adifield.MY_STATE,
	adifield.MY_CNTY,
	adifield.GRIDSQUARE,
	adifield.VUCC_GRIDS,
	adifield.MY_CQ_ZONE,
	adifield.MY_ITU_ZONE,
}

var adiWriterPriorityFieldMap = make(map[adifield.Field]struct{}, len(adiWriterPriorityFieldOrder))

func init() {
	for _, field := range adiWriterPriorityFieldOrder {
		adiWriterPriorityFieldMap[field] = struct{}{}
	}
}

var writerBufPool = sync.Pool{
	New: func() any {
		b := make([]byte, 0, 1024)
		return &b
	},
}

var writerFieldScratchPool = sync.Pool{
	New: func() any {
		s := make([]adifield.Field, 0, 32)
		return &s
	},
}

// Writer writes ADIF records to an underlying io.Writer in ADI format.
// Obtain one with NewWriter, write records with WriteHeader and Write,
// then call Flush once to ensure any buffered data in the underlying writer is flushed.
//
// Writer does not add its own buffering. Wrap the destination with bufio.NewWriter
// for buffered output, and pass it to both NewWriter and Flush.
type Writer struct {
	w              io.Writer
	headerPreamble string
	mode           WriteMode
	wroteData      bool
}

// NewWriter returns a Writer that writes ADI records to w using the default header preamble.
func NewWriter(w io.Writer) *Writer {
	return NewWriterWithPreamble(w, adiHeaderPreamble)
}

// NewWriterWithPreamble returns a Writer with a custom preamble prepended to the header record.
// Pass an empty string to use a single newline, which satisfies the ADIF spec requirement
// that a header file must not start with '<'.
func NewWriterWithPreamble(w io.Writer, preamble string) *Writer {
	return &Writer{
		w:              w,
		headerPreamble: preamble,
	}
}

// WriteHeader writes the ADIF header record.
// The header must be written before any QSO records and may only be written once.
func (w *Writer) WriteHeader(r Record) error {
	if w.wroteData {
		return ErrHeaderAlreadyWritten
	}
	preamble := w.headerPreamble
	if preamble == "" {
		preamble = "\n" // minimal preamble required by the ADIF spec
	}
	if _, err := io.WriteString(w.w, preamble); err != nil {
		return err
	}
	w.wroteData = true
	return w.writeRecord(r, 'H')
}

// Write appends a QSO record to the output.
func (w *Writer) Write(r Record) error {
	w.wroteData = true
	return w.writeRecord(r, 'R')
}

// SetWriteMode sets the WriteMode for this Writer and returns the Writer for chaining.
func (w *Writer) SetWriteMode(mode WriteMode) *Writer {
	w.mode = mode
	return w
}

// Flush flushes the underlying io.Writer if it implements Flush() error (e.g. bufio.Writer).
// It is a no-op for writers that do not buffer.
func (w *Writer) Flush() error {
	return flushWriter(w.w)
}

// flushWriter calls w.Flush() if w implements the Flush() error interface, otherwise it is a no-op.
func flushWriter(w io.Writer) error {
	if f, ok := w.(interface{ Flush() error }); ok {
		return f.Flush()
	}
	return nil
}

func (w *Writer) writeRecord(r Record, endTag byte) error {
	bufPtr := writerBufPool.Get().(*[]byte)
	buf := (*bufPtr)[:0]

	buf = appendFieldsADI(r, buf, w.mode)
	if len(buf) == 0 {
		writerBufPool.Put(bufPtr)
		return nil
	}

	buf = append(buf, '<', 'E', 'O', endTag, '>', '\n')
	_, err := w.w.Write(buf)

	*bufPtr = buf
	writerBufPool.Put(bufPtr)
	return err
}

// appendFieldsADI writes all fields of r to buf in ADI format using the given WriteMode.
// It will not write the end tag.
func appendFieldsADI(r Record, buf []byte, mode WriteMode) []byte {
	if mode == WriteModeFast {
		return appendFieldsADIFast(r, buf)
	}
	return appendFieldsADIPretty(r, buf)
}

// appendFieldsADIFast writes all fields of r to buf in ADI format as quickly and efficiently as possible.
// It does not guarantee any particular field order.
// It will not write the end tag.
func appendFieldsADIFast(r Record, buf []byte) []byte {
	for field, value := range r {
		buf = appendField(buf, field, value)
	}
	return buf
}

// appendFieldsADIPretty writes all fields of r to buf in ADI format.
// First, it writes priority fields in a fixed order.
// Next, it writes remaining fields in alphabetical order.
// It will not write the end tag.
func appendFieldsADIPretty(r Record, buf []byte) []byte {
	for _, field := range adiWriterPriorityFieldOrder {
		buf = appendField(buf, field, r[field])
	}

	scratchPtr := writerFieldScratchPool.Get().(*[]adifield.Field)
	scratch := (*scratchPtr)[:0]
	for field := range r {
		if _, isPriority := adiWriterPriorityFieldMap[field]; !isPriority {
			scratch = append(scratch, field)
		}
	}
	slices.Sort(scratch)
	for _, field := range scratch {
		buf = appendField(buf, field, r[field])
	}
	*scratchPtr = scratch
	writerFieldScratchPool.Put(scratchPtr)

	return buf
}

// appendField appends a single ADIF field in ADI format to buf.
// Returns buf unchanged when value is empty.
func appendField(buf []byte, field adifield.Field, value string) []byte {
	if value == "" {
		return buf
	}
	buf = append(buf, '<')
	buf = append(buf, field...)
	buf = append(buf, ':')
	buf = strconv.AppendInt(buf, int64(len(value)), 10)
	buf = append(buf, '>')
	buf = append(buf, value...)
	return buf
}
