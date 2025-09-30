package main

import (
	"io"
	"strings"
	"testing"

	eminlin "github.com/Eminlin/GoADIFLog"
	eminlinformat "github.com/Eminlin/GoADIFLog/format"
	matir "github.com/Matir/adifparser"
	flwyd "github.com/flwyd/adif-multitool/adif"
	farmergreg "github.com/farmergreg/adif/v5"
)

func BenchmarkReadFarmerGregADI(b *testing.B) {
	var qsoList []farmergreg.Record

	for b.Loop() {
		qsoList = make([]farmergreg.Record, 0, 10000)
		p := farmergreg.NewADIRecordReader(strings.NewReader(benchmarkFile), false)
		for {
			q, _, err := p.Next()
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

func BenchmarkReadFLWYD(b *testing.B) {
	var log *flwyd.Logfile
	var err error

	for b.Loop() {
		p := flwyd.NewADIIO()
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
