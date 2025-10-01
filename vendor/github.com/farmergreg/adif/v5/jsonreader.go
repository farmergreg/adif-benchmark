package adif

import (
	"encoding/json"
	"io"

	"github.com/farmergreg/spec/v6/adifield"
)

var _ DocumentReader = (*jsonReader)(nil)

type jsonReader struct {
	document     *jsonDocument
	currentIndex int
	skipHeader   bool
}

// NewJSONDocumentReader returns an ADIFDocumentReader that can parse ADIF records in ADIJ JSON format.
// If skipHeader is true, Next() will not return the header record if it exists.
func NewJSONDocumentReader(r io.Reader, skipHeader bool) (DocumentReader, error) {
	var doc jsonDocument
	decoder := json.NewDecoder(r)
	err := decoder.Decode(&doc)
	if err != nil {
		return nil, err
	}

	reader := &jsonReader{
		document:     &doc,
		currentIndex: -1,
		skipHeader:   skipHeader,
	}

	return reader, nil
}

// Next implements ADIFDocumentReader.
func (r *jsonReader) Next() (record Record, isHeader bool, err error) {
	if r.currentIndex >= len(r.document.Records) {
		return nil, false, io.EOF
	}
	if r.currentIndex == -1 {
		r.currentIndex = 0
		if !r.skipHeader && r.document.Header != nil {
			header := r.mapToRecord(r.document.Header)
			return header, true, nil
		}
		if len(r.document.Records) == 0 {
			return nil, false, io.EOF
		}
	}

	record = r.mapToRecord(r.document.Records[r.currentIndex])
	r.currentIndex++
	return record, false, nil
}

func (r *jsonReader) mapToRecord(fieldMap map[adifield.Field]string) Record {
	record := NewRecord()
	for field, value := range fieldMap {
		record.Set(field, value)
	}
	return record
}
