package main

import (
	"encoding/json"
	"io"
	"strings"
	"testing"

	eminlin "github.com/Eminlin/GoADIFLog"
	eminlinformat "github.com/Eminlin/GoADIFLog/format"
	matir "github.com/Matir/adifparser"
	"github.com/hamradiolog-net/adif-parser/v5"
)

func BenchmarkReadHamRadioLogDotNet(b *testing.B) {
	var qsoList []adif.ADIFRecord

	for b.Loop() {
		qsoList = make([]adif.ADIFRecord, 0, 10000)
		p := adif.NewADIReader(strings.NewReader(benchmarkFile), false)
		for {
			q, err := p.Next()
			if err == io.EOF {
				break
			}
			qsoList = append(qsoList, q)
		}
	}

	_ = len(qsoList)
}

func BenchmarkReadMatir(b *testing.B) {
	var qsoList []matir.ADIFRecord

	for b.Loop() {
		qsoList = make([]matir.ADIFRecord, 0, 10000)
		r := matir.NewADIFReader(strings.NewReader(benchmarkFile))
		for {
			q, err := r.ReadRecord()
			if err != nil {
				break
			}
			qsoList = append(qsoList, q)
		}
	}

	_ = len(qsoList)
}

func BenchmarkReadEminlin(b *testing.B) {
	var log []eminlinformat.CQLog
	var err error

	for b.Loop() {
		log, err = eminlin.ParseAdifFromString(benchmarkFile)
		if err != nil {
			b.Fatal(err)
		}
	}

	_ = len(log)
}

// This benchmark works directly on JSON data, without using this library for reading the JSON ADIF data.
// It is meant as a reference point to compare the performance to a known standard (the go stdlib JSON parser).
func BenchmarkReadJSONReference(b *testing.B) {
	jsonRecords := benchmarkFileAsJSON()

	b.ResetTimer()
	document := adifDocument{}
	for b.Loop() {
		err := json.Unmarshal(jsonRecords, &document)
		if err != nil {
			b.Fatal(err)
		}
	}
}
