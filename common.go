package main

import (
	"io"
	"strings"

	_ "embed"

	"github.com/hamradiolog-net/adif/v4"
)

//go:embed testdata/N3FJP-AClogAdif.adi
var benchmarkFile string

func loadTestData() []adif.Record {
	var qsoListNative []adif.Record
	p := adif.NewADIReader(strings.NewReader(benchmarkFile), false)
	for {
		record, _, _, err := p.Next()
		if err == io.EOF {
			break
		}
		qsoListNative = append(qsoListNative, record)
	}
	return qsoListNative
}
