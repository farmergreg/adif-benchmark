package main

import (
	"strings"
	"testing"

	_ "embed"

	matir "github.com/Matir/adifparser"
	flwyd "github.com/flwyd/adif-multitool/adif"
	farmergreg "github.com/farmergreg/adif/v5"
)

func BenchmarkWriteFarmerGreg(b *testing.B) {
	qsoListNative := loadTestData()
	b.ResetTimer()
	for b.Loop() {
		var sb strings.Builder
		w := farmergreg.NewADIRecordWriter(&sb)
		for _, qso := range qsoListNative {
			err := w.Write(qso, false)
			if err != nil {
				b.Fatal(err)
			}
		}
		_ = sb.String()
	}
}

func BenchmarkWriteMatir(b *testing.B) {
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

func BenchmarkWriteFLWYD(b *testing.B) {
	p := flwyd.NewADIIO()
	doc, err := p.Read(strings.NewReader(benchmarkFile))
	if err != nil {
		b.Fatal(err)
	}

	for b.Loop() {
		p = flwyd.NewADIIO()
		var sb strings.Builder
		p.Write(doc, &sb)
	}
}

/*
// Eminlin does not support writing adi files
func BenchmarkWriteEminlin(b *testing.B) {
}
*/
