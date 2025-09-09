package adif

import (
	"io"
	"math"
	"slices"
	"strconv"
	"sync"

	"github.com/hamradiolog-net/adif-spec/v6/adifield"
)

var _ ADIFRecordWriter = (*adiWriter)(nil)

const adiHeaderPreamble = "                    AM✠DG\nK9CTS High Performance ADIF Processing Library\nhttps://github.com/hamradiolog-net/adif-parser\n\n"

// adiWriterPriorityFieldOrder defines the order of priority fields when writing ADIF records.
// These fields are written first, in this order.
var adiWriterPriorityFieldOrder = [...]adifield.ADIField{

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
var adiWriterPriorityFieldMap = make(map[adifield.ADIField]struct{}, len(adiWriterPriorityFieldOrder))

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
	headerPreamble string
}

var adiWriterBufferPool = sync.Pool{
	New: func() any {
		b := make([]byte, 0, 1024)
		return &b
	},
}

// NewADIWriter returns an ADIFRecordWriter that can write ADIF *.adi formatted records.
func NewADIWriter(w io.Writer) *adiWriter {
	return NewADIWriterWithPreamble(w, adiHeaderPreamble)
}

// NewADIWriterWithPreamble returns an ADIFRecordWriter that can write ADIF *.adi formatted records with a custom preamble for header records.
func NewADIWriterWithPreamble(w io.Writer, adiPreamble string) *adiWriter {
	return &adiWriter{
		w:              w,
		headerPreamble: adiPreamble,
	}
}

func (w *adiWriter) Write(r []ADIFRecord) error {
	if len(r) > 0 && r[0].IsHeader() {
		if w.headerPreamble == "" {
			// preamble is mandatory per the ADIF specification.
			w.w.Write([]byte{'\n'})
		} else {
			w.w.Write([]byte(w.headerPreamble))
		}
	}

	for _, record := range r {
		err := w.writeInternal(record)
		if err != nil {
			return err
		}
	}
	return nil
}

func (w *adiWriter) writeInternal(r ADIFRecord) error {
	adiLength := appendADIFRecordAsADIPreCalculate(r)
	bufPtr := adiWriterBufferPool.Get().(*[]byte)
	buf := *bufPtr

	if cap(buf) < adiLength {
		buf = make([]byte, 0, adiLength)
		*bufPtr = buf
	}
	buf = buf[:0]

	buf = appendAsADI(r, buf)
	_, err := w.w.Write(buf)

	adiWriterBufferPool.Put(bufPtr)
	return err
}

// appendAsADI writes the ADI formatted QSO record to the provided buffer.
// The buffer should have sufficient capacity to avoid reallocations.
// You should use appendAsADIPreCalculate() to determine the required buffer capacity.
// Field order is NOT guaranteed to be stable.
func appendAsADI(r ADIFRecord, buf []byte) []byte {
	if len(r.Fields()) == 0 {
		return buf
	}

	// Priority fields first (in order, but nothing about this is guaranteed to be stable between versions of this library)
	for _, field := range adiWriterPriorityFieldOrder {
		buf = appendADIFRecordAsADI(r, buf, field)
	}

	// Remaining fields
	fields := r.Fields()
	slices.Sort(fields)
	for _, field := range fields {
		if _, isPriority := adiWriterPriorityFieldMap[field]; isPriority {
			continue
		}
		buf = appendADIFRecordAsADI(r, buf, field)
	}

	if r.IsHeader() {
		buf = append(buf, "<EOH>\n"...)
	} else {
		buf = append(buf, "<EOR>\n"...)
	}

	return buf
}

// appendADIFRecordAsADI adds a single ADIF field to the buffer
func appendADIFRecordAsADI(r ADIFRecord, buf []byte, field adifield.ADIField) []byte {
	value := r.Get(field)
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

// appendADIFRecordAsADIPreCalculate returns the length of the record in bytes when exported to ADI format by the AppendAsADI method.
func appendADIFRecordAsADIPreCalculate(r ADIFRecord) (adiLength int) {
	for _, field := range r.Fields() {
		valueLength := len(r.Get(field))
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

	return adiLength + 6 // "<EOR>\n" / "<EOH>\n"
}
