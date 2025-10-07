package adif

import (
	"io"
	"math"
	"strconv"
	"sync"

	"github.com/farmergreg/spec/v6/adifield"
)

var _ DocumentWriter = (*adiWriter)(nil)

const adiHeaderPreamble = "                    AM✠DG\nK9CTS High Performance ADIF Processing Library\n   https://github.com/farmergreg/adif\n\n"

// adiWriterPriorityFieldOrder defines the order of priority fields when writing ADIF records.
// These fields are written first, in this order.
var adiWriterPriorityFieldOrder = [...]adifield.Field{

	// Minimum "required" fields:
	// https://www.adif.org/315/ADIF_315_Resources.htm#ADIFImplementationNotesMinimumFields
	adifield.QSO_DATE,
	adifield.TIME_ON,
	adifield.BAND,
	adifield.MODE,
	adifield.SUBMODE, // not technically part of the "required" fields, but it logically belongs here.
	adifield.CALL,    // placing CALL last so that similar records are vertically aligned and easy to read in ADI format.

	// Additional Fields.
	// Currently, we match LoTW's list:
	// https://lotw.arrl.org/lotw-help/developer-submit-qsos/
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

// adiWriterPriorityFieldMap is used for quick lookups to determine if a field is a priority field
var adiWriterPriorityFieldMap = make(map[adifield.Field]struct{}, len(adiWriterPriorityFieldOrder))

func init() {
	for _, field := range adiWriterPriorityFieldOrder {
		adiWriterPriorityFieldMap[field] = struct{}{}
	}
}

type adiWriter struct {
	w io.Writer

	// This text will be added to the start of a file if there is a header record.
	// To satisfy the ADIF specification which states:
	// If the first character in an ADI file is <, it contains no Header.
	headerPreamble  string
	isHeaderWritten bool
}

var adiWriterBufferPool = sync.Pool{
	New: func() any {
		b := make([]byte, 0, 1024)
		return &b
	},
}

// NewADIDocumentWriter returns an ADIFDocumentWriter that can write ADIF *.adi formatted records.
func NewADIDocumentWriter(w io.Writer) DocumentWriter {
	return NewADIDocumentWriterWithPreamble(w, adiHeaderPreamble)
}

// NewADIDocumentWriterWithPreamble returns an ADIFDocumentWriter that can write ADIF *.adi formatted records with a custom preamble for header records.
func NewADIDocumentWriterWithPreamble(w io.Writer, adiPreamble string) DocumentWriter {
	return &adiWriter{
		w:              w,
		headerPreamble: adiPreamble,
	}
}

func (w *adiWriter) WriteHeader(r Record) error {
	if w.isHeaderWritten {
		return ErrHeaderAlreadyWritten
	}
	w.isHeaderWritten = true

	if w.headerPreamble == "" {
		// preamble is mandatory per the ADIF specification.
		w.w.Write([]byte{'\n'})
	} else {
		w.w.Write([]byte(w.headerPreamble))
	}
	return w.writeInternal(r, true)
}

func (w *adiWriter) WriteRecord(r Record) error {
	w.isHeaderWritten = true
	return w.writeInternal(r, false)
}

func (w *adiWriter) writeInternal(r Record, isHeader bool) error {
	adiLength := appendADIFRecordAsADIPreCalculate(r)
	bufPtr := adiWriterBufferPool.Get().(*[]byte)
	buf := *bufPtr

	if cap(buf) < adiLength {
		buf = make([]byte, 0, adiLength)
		*bufPtr = buf
	}
	buf = buf[:0]

	buf = appendAsADI(r, isHeader, buf)
	_, err := w.w.Write(buf)

	adiWriterBufferPool.Put(bufPtr)
	return err
}

// appendAsADI writes the ADI formatted QSO record to the provided buffer.
// The buffer should have sufficient capacity to avoid reallocations.
// You should use appendAsADIPreCalculate() to determine the required buffer capacity.
// Field order is NOT guaranteed to be stable.
func appendAsADI(r Record, isHeader bool, buf []byte) []byte {
	if r.FieldCount() == 0 {
		return buf
	}

	// Priority fields first (in order, but nothing about this is guaranteed to be stable between versions of this library)
	for _, field := range adiWriterPriorityFieldOrder {
		buf = appendADIFRecordAsADI(buf, field, r.Get(field))
	}

	// Remaining fields
	for field, value := range r.Fields() {
		if _, isPriority := adiWriterPriorityFieldMap[field]; isPriority {
			continue
		}
		buf = appendADIFRecordAsADI(buf, field, value)
	}

	if isHeader {
		buf = append(buf, "<EOH>\n"...)
	} else {
		buf = append(buf, "<EOR>\n"...)
	}

	return buf
}

// appendADIFRecordAsADI adds a single ADIF field to the buffer
func appendADIFRecordAsADI(buf []byte, field adifield.Field, value string) []byte {
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

// appendADIFRecordAsADIPreCalculate returns the length of the record in bytes when exported to ADI format by the appendAsADI method.
func appendADIFRecordAsADIPreCalculate(r Record) (adiLength int) {
	for field, value := range r.Fields() {
		valueLength := len(value)
		if valueLength == 0 {
			continue
		}
		adiLength += 3 + valueLength + len(field) // 3 for '<', ':', '>'

		// Avoid strconv.Itoa string allocation by calculating number of base 10 digits mathematically
		switch {
		case valueLength < 10:
			adiLength += 1
		case valueLength < 100:
			adiLength += 2
		case valueLength < 1_000:
			adiLength += 3
		case valueLength < 10_000:
			adiLength += 4
		case valueLength < 100_000:
			adiLength += 5
		default:
			adiLength += int(math.Log10(float64(valueLength))) + 1
		}
	}

	return adiLength + 6 // "<eor>\n" / "<eoh>\n"
}

func (w *adiWriter) Flush() error {
	return nil
}
