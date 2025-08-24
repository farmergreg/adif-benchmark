package main

import (
	"encoding/json"
	"io"
	"strings"
	"testing"

	eminlin "github.com/Eminlin/GoADIFLog"
	eminlinformat "github.com/Eminlin/GoADIFLog/format"
	matir "github.com/Matir/adifparser"
	"github.com/hamradiolog-net/adif"
)

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

func BenchmarkReadHamRadioLogDotNet(b *testing.B) {
	var qsoList []adif.Record

	for b.Loop() {
		qsoList = make([]adif.Record, 0, 10000)
		p := adif.NewADIReader(strings.NewReader(benchmarkFile), false)
		for {
			q, _, _, err := p.Next()
			if err == io.EOF {
				break
			}
			qsoList = append(qsoList, q)
		}
	}

	_ = len(qsoList)
}

func BenchmarkReadJSON(b *testing.B) {
	qsoListNative := loadTestData()
	jsonData, err := json.Marshal(qsoListNative)
	if err != nil {
		b.Fatal(err)
	}
	jsonString := string(jsonData)

	var records []adif.Record
	var readCountJSON int

	for b.Loop() {
		records = records[:0]
		// convoluted, but this is to match the other benchmarks which also work with string input...
		// in reality, this does not affect the speed of this benchmark...
		err := json.Unmarshal([]byte(jsonString), &records)
		if err != nil {
			b.Fatal(err)
		}
		readCountJSON = len(records)
	}

	if len(qsoListNative) != readCountJSON {
		b.Errorf("Read count mismatch: JSON %d, expected %d", readCountJSON, len(qsoListNative))
	}
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
