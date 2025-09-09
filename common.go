package main

import (
	"bytes"
	"io"
	"strings"

	_ "embed"

	"github.com/hamradiolog-net/adif-parser/v5"
)

//go:embed testdata/N3FJP-AClogAdif.adi
var benchmarkFile string

func benchmarkFileAsJSON() []byte {
	var buffer bytes.Buffer
	src := adif.NewADIReader(strings.NewReader(benchmarkFile), false)
	dst := adif.NewADIJWriter(&buffer)
	srcRecords := make([]adif.ADIFRecord, 0, 10000)
	for {
		record, err := src.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			panic(err)
		}
		srcRecords = append(srcRecords, record)
	}
	dst.Write(srcRecords)
	return buffer.Bytes()
}

func loadTestData() []adif.ADIFRecord {
	var qsoListNative []adif.ADIFRecord
	p := adif.NewADIReader(strings.NewReader(benchmarkFile), false)
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
