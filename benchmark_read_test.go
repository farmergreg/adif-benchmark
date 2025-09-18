package main

import (
	"io"
	"strings"
	"testing"

	eminlin "github.com/Eminlin/GoADIFLog"
	eminlinformat "github.com/Eminlin/GoADIFLog/format"
	matir "github.com/Matir/adifparser"
	multitool "github.com/flwyd/adif-multitool/adif"
	hrln "github.com/farmergreg/adif/v5"
)

func BenchmarkReadFarmerGreg(b *testing.B) {
	var qsoList []hrln.Record

	for b.Loop() {
		qsoList = make([]hrln.Record, 0, 10000)
		p := hrln.NewADIRecordReader(strings.NewReader(benchmarkFile), false)
		for {
			q, err := p.Next()
			if err == io.EOF {
				break
			}
			if err != nil {
				b.Fatal(err)
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
			if err == io.EOF {
				break
			}
			if err != nil {
				b.Fatal(err)
			}
			qsoList = append(qsoList, q)
			_ = len(qsoList)
		}
	}
	_ = len(qsoList)
}

func BenchmarkReadAdifMultitool(b *testing.B) {
	var log *multitool.Logfile
	var err error

	for b.Loop() {
		p := multitool.NewADIIO()
		log, err = p.Read(strings.NewReader(benchmarkFile))
		if err != nil {
			b.Fatal(err)
		}
		_ = len(log.Records)
	}
	_ = len(log.Records)
}

func BenchmarkReadEminlin(b *testing.B) {
	var log []eminlinformat.CQLog
	var err error

	for b.Loop() {
		log, err = eminlin.ParseAdifFromString(benchmarkFile)
		if err != nil {
			b.Fatal(err)
		}
		_ = len(log)
	}
	_ = len(log)
}
