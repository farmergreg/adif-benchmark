package adif

import (
	"encoding/json"
	"io"

	"github.com/hamradiolog-net/spec/v6/adifield"
)

var _ ADIFRecordWriter = (*jsonWriter)(nil)

// jsonWriter implements ADIFRecordWriter for writing ADIF records in ADIJ format.
type jsonWriter struct {
	w io.Writer
}

func NewJSONRecordWriter(w io.Writer) *jsonWriter {
	return &jsonWriter{w: w}
}

// Write implements ADIFRecordWriter.Write for writing ADIF records in ADIJ format.
func (aw *jsonWriter) Write(records []Record) error {
	doc := &adifDocument{}
	if len(records) > 0 && records[0].IsHeader() {
		doc.Header = adijRecordToMap(records[0])
		records = records[1:]
	}

	for _, record := range records {
		doc.Records = append(doc.Records, adijRecordToMap(record))
	}

	encoder := json.NewEncoder(aw.w)
	encoder.SetIndent("", "  ")
	err := encoder.Encode(doc)
	if err != nil {
		return err
	}

	return nil
}

func adijRecordToMap(record Record) map[adifield.Field]string {
	result := make(map[adifield.Field]string)
	for field, value := range record.All() {
		if value != "" {
			result[field] = value
		}
	}
	return result
}
