package main

import (
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
	s := adif.NewScanner(strings.NewReader(benchmarkFile))
	for s.Scan() {
		qsoListNative = append(qsoListNative, s.Record())
	}
	return qsoListNative
}

func loadFarmerGregDocument() *adif.Document {
	doc := adif.NewDocument()
	doc.ReadFrom(strings.NewReader(benchmarkFile)) //nolint:errcheck
	return doc
}

