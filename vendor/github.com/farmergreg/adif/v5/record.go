package adif

import (
	"encoding/json"
	"io"
	"strings"

	"github.com/farmergreg/spec/v6/adifield"
)

// Record is a map of ADIF fields to their values, representing either a header or QSO record.
// Field keys are normalized to uppercase by the Scanner.
// Use adifield constants or adifield.New() to construct field names.
//
// Example:
//
//	r := adif.NewRecord()
//	r[adifield.CALL] = "K9CTS"
//	r[adifield.BAND] = "20m"
type Record map[adifield.Field]string

// NewRecord returns a new empty Record with a sensible default capacity.
func NewRecord() Record {
	return make(Record, 7)
}

// String returns the record's fields serialized in ADI format, without an EOR or EOH tag.
// Priority fields are written first in a fixed order; remaining fields follow in alphabetical order.
// Implements fmt.Stringer.
func (r Record) String() string {
	var sb strings.Builder
	r.WriteTo(&sb) //nolint:errcheck — strings.Builder.Write never returns an error
	return sb.String()
}

// WriteTo writes the record's fields in ADI format to w, without an EOR or EOH tag.
// Priority fields are written first in a fixed order; remaining fields follow in alphabetical order.
// Implements io.WriterTo.
func (r Record) WriteTo(w io.Writer) (int64, error) {
	return r.WriteToMode(w, WriteModePretty)
}

// WriteToMode writes the record's fields in ADI format to w using the given WriteMode, without an EOR or EOH tag.
func (r Record) WriteToMode(w io.Writer, mode WriteMode) (int64, error) {
	bufPtr := writerBufPool.Get().(*[]byte)
	buf := appendFieldsADI(r, (*bufPtr)[:0], mode)
	n, err := w.Write(buf)
	*bufPtr = buf
	writerBufPool.Put(bufPtr)
	return int64(n), err
}

// UnmarshalJSON implements json.Unmarshaler, allowing a Record to be unmarshaled from a JSON object with string keys and values.
// It ensures that field keys are normalized to uppercase and converted to adifield.Field types.
func (r *Record) UnmarshalJSON(data []byte) error {
	var raw map[string]string
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	*r = make(Record, len(raw))
	for k, v := range raw {
		(*r)[adifield.New(strings.ToUpper(k))] = v
	}
	return nil
}
