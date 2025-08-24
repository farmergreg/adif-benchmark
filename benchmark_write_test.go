package main

import (
	"encoding/json"
	"strings"
	"testing"

	_ "embed"

	matir "github.com/Matir/adifparser"
)

/*
// Eminlin does not support writing adi files
func BenchmarkWriteEminlin(b *testing.B) {
}
*/

func BenchmarkWriteHamRadioLogDotNet(b *testing.B) {
	qsoListNative := loadTestData()
	var writeCountADI int

	for b.Loop() {
		var sb strings.Builder
		writeCountADI = 0
		for _, qso := range qsoListNative {
			qso.WriteTo(&sb)
			writeCountADI++
		}
		_ = sb.String()
	}

	if len(qsoListNative) != writeCountADI {
		b.Errorf("Write count mismatch: ADI %d, expected %d", writeCountADI, len(qsoListNative))
	}
}

func BenchmarkWriteJSON(b *testing.B) {
	qsoListNative := loadTestData()

	for b.Loop() {
		data, err := json.Marshal(qsoListNative)
		if err != nil {
			b.Fatal(err)
		}
		_ = string(data)
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
