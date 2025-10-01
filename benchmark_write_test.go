package main

import (
	"encoding/json"
	"strings"
	"testing"

	_ "embed"

	matir "github.com/Matir/adifparser"
	flwyd "github.com/flwyd/adif-multitool/adif"
	farmergreg "github.com/farmergreg/adif/v5"
)

func BenchmarkWriteFarmerGregADI(b *testing.B) {
	qsoListNative := loadTestData()
	b.ResetTimer()
	for b.Loop() {
		var sb strings.Builder
		w := farmergreg.NewADIDocumentWriter(&sb)
		for _, qso := range qsoListNative {
			err := w.WriteRecord(qso)
			if err != nil {
				b.Fatal(err)
			}
		}
		_ = sb.String()
	}
}

func BenchmarkWriteFarmerGregJSON(b *testing.B) {
	qsoListNative := loadTestData()
	b.ResetTimer()
	for b.Loop() {
		var sb strings.Builder
		w := farmergreg.NewJSONDocumentWriter(&sb, "")
		for _, qso := range qsoListNative {
			err := w.WriteRecord(qso)
			if err != nil {
				b.Fatal(err)
			}
		}
		if err := w.Close(); err!=nil {
			b.Fatal(err)
		}
		_ = sb.String()
	}
}

func BenchmarkWriteMatirADI(b *testing.B) {
	// Setup Matir test data
	var qsoListMatir []matir.ADIFRecord
	r := matir.NewADIFReader(strings.NewReader(benchmarkFile))
	for {
		qso, err := r.ReadRecord()
		if err != nil {
			break
		}
		qsoListMatir = append(qsoListMatir, qso)
	}

	for b.Loop() {
		var sb strings.Builder
		mw := matir.NewADIFWriter(&sb)
		for _, qso := range qsoListMatir {
			mw.WriteRecord(qso)
		}
		_ = sb.String()
	}
}

func BenchmarkWriteFlwydADI(b *testing.B) {
	p := flwyd.NewADIIO()
	doc, err := p.Read(strings.NewReader(benchmarkFile))
	if err != nil {
		b.Fatal(err)
	}

	for b.Loop() {
		p = flwyd.NewADIIO()
		var sb strings.Builder
		p.Write(doc, &sb)
		_ = sb.String()
	}
}

/*
// Eminlin does not support writing adi files
func BenchmarkWriteEminlinADI(b *testing.B) {
}
*/

// BenchmarkWriteGoStdLibDirectJSON benchmarks writing ADIF data using the Go standard library's encoding/json package.
func BenchmarkWriteGoStdLibDirectJSON(b *testing.B) {
	doc := &jsonDocument{}
	decoder := json.NewDecoder(strings.NewReader(benchmarkFileAsJSON))
	err := decoder.Decode(doc)
	if err != nil {
		b.Fatal(err)
	}

	for b.Loop() {
		var sb strings.Builder
		encoder := json.NewEncoder(&sb)
		err := encoder.Encode(&doc)
		if err != nil {
			b.Fatal(err)
		}
		_ = sb.String()
	}
}