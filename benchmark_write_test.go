package main

import (
	"encoding/json"
	"io"
	"strings"
	"testing"

	_ "embed"

	matir "github.com/Matir/adifparser"
	farmergreg "github.com/farmergreg/adif/v5"
	flwyd "github.com/flwyd/adif-multitool/adif"
)

func BenchmarkWriteFarmerGregADIFast(b *testing.B) {
	qsoListNative := loadTestData()
	b.ResetTimer()
	for b.Loop() {
		var sb strings.Builder
		w := farmergreg.NewWriter(&sb).SetWriteMode(farmergreg.WriteModeFast)
		for _, qso := range qsoListNative {
			if err := w.Write(qso); err != nil {
				b.Fatal(err)
			}
		}
		if err := w.Flush(); err != nil {
			b.Fatal(err)
		}
		_ = sb.String()
	}
}

func BenchmarkWriteFarmerGregADIPretty(b *testing.B) {
	qsoListNative := loadTestData()
	b.ResetTimer()
	for b.Loop() {
		var sb strings.Builder
		w := farmergreg.NewWriter(&sb)
		for _, qso := range qsoListNative {
			if err := w.Write(qso); err != nil {
				b.Fatal(err)
			}
		}
		if err := w.Flush(); err != nil {
			b.Fatal(err)
		}
		_ = sb.String()
	}
}

func BenchmarkWriteFarmerGregJSON(b *testing.B) {
	doc := loadFarmerGregDocument()
	b.ResetTimer()
	for b.Loop() {
		var sb strings.Builder
		if err := json.NewEncoder(&sb).Encode(doc); err != nil {
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
		if err == io.EOF {
			break
		}
		if err != nil {
			b.Fatal(err)
		}
		qsoListMatir = append(qsoListMatir, qso)
	}
	b.ResetTimer()

	for b.Loop() {
		var sb strings.Builder
		mw := matir.NewADIFWriter(&sb)
		for _, qso := range qsoListMatir {
			mw.WriteRecord(qso)
		}
		if err := mw.Flush(); err != nil {
			b.Fatal(err)
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
	b.ResetTimer()

	for b.Loop() {
		p = flwyd.NewADIIO()
		var sb strings.Builder
		if err := p.Write(doc, &sb); err != nil {
			b.Fatal(err)
		}
		_ = sb.String()
	}
}

/*
// Eminlin does not support writing adi files
func BenchmarkWriteEminlinADI(b *testing.B) {
}
*/
