package adif

import (
	"fmt"
	"io"
	"math"
	"strconv"
	"sync"

	"github.com/hamradiolog-net/adif-spec/v6/adifield"
)

// Interface implementations
var (
	_ io.WriterTo   = &Record{}
	_ io.ReaderFrom = &Record{}
	_ fmt.Stringer  = &Record{}
)

var recordBufferPool = sync.Pool{
	New: func() any {
		b := make([]byte, 0, 1024)
		return &b
	},
}

// recordPriorityFieldOrder defines the order of priority fields when writing ADIF records.
// These fields are written first, in this order.
var recordPriorityFieldOrder = [...]adifield.ADIField{

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

// recordPriorityFields is used for quick lookups to determine if a field is a priority field
var recordPriorityFields = make(map[adifield.ADIField]struct{}, len(recordPriorityFieldOrder))

func init() {
	for _, field := range recordPriorityFieldOrder {
		recordPriorityFields[field] = struct{}{}
	}
}

// NewRecord creates a new Record with the default initial capacity.
func NewRecord() Record {
	return NewRecordWithCapacity(-1)
}

// NewRecordWithCapacity creates a new Record with a specific initial capacity.
func NewRecordWithCapacity(initialCapacity int) Record {
	if initialCapacity < 1 {
		initialCapacity = 7
	}
	return make(Record, initialCapacity)
}

// Reset clears the record of all fields.
func (r Record) Reset() {
	clear(r)
}

// ReadFrom reads an ADIF formatted record from the provided io.Reader.
// It returns the number of bytes read and any error encountered.
// io.EOF is returned when no more records are available.
// Existing fields will be updated and new fields added as necessary.
// Header records are SKIPPED.
//
// n.b. This method is best used to read a single record.
// When reading multiple records, create an ADIFReader using NewADIReader().
// Use its Next() method to obtain maximum speed and memory efficiency while iterating over records.
func (r *Record) ReadFrom(src io.Reader) (int64, error) {
	p := NewADIReader(src, true)
	var n int64

	record, _, c, err := p.Next()
	n += c
	if err != nil {
		return n, err
	}

	*r = record
	return n, nil
}

// WriteTo writes ADI formatted record data to the provided io.Writer.
// It WILL NOT write the <EOR> / <EOH> tag because it has no way to tell what type of Record it is.
// It returns the number of bytes written and any error encountered.
func (r *Record) WriteTo(dest io.Writer) (int64, error) {
	adiLength := r.appendAsADIPreCalculate()
	bufPtr := recordBufferPool.Get().(*[]byte)
	buf := *bufPtr

	if cap(buf) < adiLength {
		buf = make([]byte, 0, adiLength)
		*bufPtr = buf
	}
	buf = buf[:0]

	buf = r.appendAsADI(buf)
	n, err := dest.Write(buf)

	recordBufferPool.Put(bufPtr)
	return int64(n), err
}

// appendAsADI writes the ADI formatted QSO record to the provided buffer.
// The buffer should have sufficient capacity to avoid reallocations.
// You should use appendAsADIPreCalculate() to determine the required buffer capacity.
// Field order is NOT guaranteed to be stable.
func (r *Record) appendAsADI(buf []byte) []byte {
	if len(*r) == 0 {
		return buf
	}

	// Priority fields first (in order, but nothing about this is guaranteed to be stable between versions of this library)
	for _, field := range recordPriorityFieldOrder {
		buf = r.appendField(buf, field)
	}

	// Remaining fields
	for field := range *r {
		if _, isPriority := recordPriorityFields[field]; isPriority {
			continue
		}
		buf = r.appendField(buf, field)
	}

	return buf
}

// appendField adds a single ADIF field to the buffer
func (r *Record) appendField(buf []byte, field adifield.ADIField) []byte {
	value, ok := (*r)[field]
	if !ok {
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

// appendAsADIPreCalculate returns the length of the record in bytes when exported to ADI format by the AppendAsADI method.
func (r *Record) appendAsADIPreCalculate() (adiLength int) {
	if len(*r) == 0 {
		return 0
	}

	for field, value := range *r {
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
		default:
			adiLength += int(math.Log10(float64(valueLength))) + 1
		}
	}

	return adiLength
}

// String returns the ADIF record as a string.
// The field order is NOT guaranteed to be stable.
func (r *Record) String() string {
	adiLength := r.appendAsADIPreCalculate()
	bufPtr := recordBufferPool.Get().(*[]byte)
	buf := *bufPtr

	if cap(buf) < adiLength {
		buf = make([]byte, 0, adiLength)
		*bufPtr = buf
	}
	buf = buf[:0]

	buf = r.appendAsADI(buf)
	s := string(buf)

	recordBufferPool.Put(bufPtr)
	return s
}
