#!/usr/bin/env bash

go test -bench=BenchmarkParse -cpuprofile cpu.out
go tool pprof -http=:8000 cpu.out
