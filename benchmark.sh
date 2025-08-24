#!/usr/bin/env sh

echo "================================================"
echo Read Benchmark
echo "================================================"
go test -bench=BenchmarkRead -benchmem
echo
echo
echo "================================================"
echo Write Benchmark
echo "================================================"
go test -bench=BenchmarkWrite -benchmem
