package main

import (
	"encoding/json"
	"strings"
	"testing"

	_ "embed"

	matir "github.com/Matir/adifparser"
	"github.com/hamradiolog-net/adif-parser/v5"
)
func BenchmarkWriteHamRadioLogDotNet(b *testing.B) {
	qsoListNative := loadTestData()
	b.ResetTimer()
	for b.Loop() {
		var sb strings.Builder
		w := adif.NewADIWriter(&sb)
		w.Write(qsoListNative)
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

/*
// Eminlin does not support writing adi files
func BenchmarkWriteEminlin(b *testing.B) {
}
*/

func BenchmarkWriteJSONReference(b *testing.B) {
	jsonRecords := benchmarkFileAsJSON()
	document := adifDocument{}
	err := json.Unmarshal(jsonRecords, &document)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for b.Loop() {
		_, err := json.Marshal(document)
		if err != nil {
			panic(err)
		}
	}
}
