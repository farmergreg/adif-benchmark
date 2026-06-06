#!/usr/bin/env sh

echo "================================================"
echo Read Benchmark
echo "================================================"
go test -run=^$ -bench=^BenchmarkRead -benchtime=10s -benchmem ./...
echo
echo
echo "================================================"
echo Write Benchmark
echo "================================================"
go test -run=^$ -bench=^BenchmarkWrite -benchtime=10s -benchmem ./...