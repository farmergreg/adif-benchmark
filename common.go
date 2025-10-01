package main

import (
	"io"
	"strings"

	_ "embed"

	"github.com/farmergreg/adif/v5"
)

//go:embed testdata/N3FJP-AClogAdif.adi
var benchmarkFile string

//go:embed testdata/N3FJP-AClogAdif.adij
var benchmarkFileAsJSON string

func loadTestData() []adif.Record {
	var qsoListNative []adif.Record
	p := adif.NewADIDocumentReader(strings.NewReader(benchmarkFile), false)
	for {
		record, _, err := p.Next()
		if err == io.EOF {
			break
		}
		qsoListNative = append(qsoListNative, record)
	}
	return qsoListNative
}

// jsonDocument represents an ADIF document using a json container format.
type jsonDocument struct {
	// Header is nil when there is no header.
	// Otherwise it is a Record with header fields inside.
	Header map[string]string `json:"header,omitempty"`

	// Records is a slice of Record.
	// It contains zero or more QSO records.
	Records []map[string]string `json:"records"`
}
