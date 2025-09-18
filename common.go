package main

import (
	"io"
	"strings"

	_ "embed"

	"github.com/farmergreg/adif/v5"
)

//go:embed testdata/N3FJP-AClogAdif.adi
var benchmarkFile string

func loadTestData() []adif.Record {
	var qsoListNative []adif.Record
	p := adif.NewADIRecordReader(strings.NewReader(benchmarkFile), false)
	for {
		record, err := p.Next()
		if err == io.EOF {
			break
		}
		qsoListNative = append(qsoListNative, record)
	}
	return qsoListNative
}

// adifDocument represents an ADIF document.
// This may be used directly with the encoding/json package to marshal or unmarshal ADIJ (ADIF as JSON) data.
type adifDocument struct {
	// Header is nil when there is no header.
	// Otherwise it is a Record with header fields inside.
	Header map[string]string `json:"HEADER,omitempty"`

	// Records is a slice of Record.
	// It contains zero or more QSO records.
	Records []map[string]string `json:"RECORDS"`
}
