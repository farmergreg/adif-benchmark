# ⚡ World's Fastest ADI Parser for Go

This library provides high-performance processing of [ADIF](https://adif.org/) (Amateur Data Interchange Format) ADI files used for ham radio logs.
It's idiomatic, developer-friendly API seamlessly integrates with your codebase and the go standard library.

[![Tests](https://github.com/farmergreg/adif/actions/workflows/test.yml/badge.svg)](https://github.com/farmergreg/adif/actions/workflows/test.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/farmergreg/adif)](https://goreportcard.com/report/github.com/farmergreg/adif)
[![Go Reference](https://pkg.go.dev/badge/github.com/farmergreg/adif.svg)](https://pkg.go.dev/github.com/farmergreg/adif)
[![Go Version](https://img.shields.io/github/go-mod/go-version/farmergreg/adif)](https://github.com/farmergreg/adif/blob/main/go.mod)
[![License](https://img.shields.io/github/license/farmergreg/adif)](https://github.com/farmergreg/adif/blob/main/LICENSE)

## ✨ Features

- 🔬 **Tested**: 100% test coverage!
- 🔧 **Developer Friendly**: Clean, idiomatic, mock friendly interfaces
- 🚀 **Blazing Fast**: 2x-20x faster than other ADI libraries
- 💡 **Memory Efficient**: Uses 2x less memory and makes 1400x fewer allocations than other libraries.

## 🚀 Quick Start

```bash
go get github.com/farmergreg/adif/v5
```

1) [ADI Processing Example](./example_adi_test.go)
2) ADX XML Processing: Not implemented. PR(s) welcome!
3) [JSON Processing Example](./example_json_test.go): Experimental. Not optimized; not part of the ADIF Specification.

## Benchmarks

Please see the [Go ADIF Parser Benchmarks](https://github.com/farmergreg/adif-benchmark) project for benchmarks.

## 🔧 Technical Deep Dive (ADI Parser)

The ADI parser in this library achieves high performance through the following optimizations:

### Performance Optimizations

- Leverages stdlib I/O operations with SSE/SIMD acceleration depending upon your CPU architecture
- Smart buffer pre-allocation based on discovered record sizes
- Optimized base-10 integer parsing for ADIF field lengths

### Memory Management

- Zero-copy techniques minimize memory operations
- String interning of repeated field names greatly reduces copying, allocations, and memory use
- Minimal temporary allocations during field parsing
- Dynamic buffer sizing based on learned field counts
- Buffer pooling

## Related Projects

If you found this library useful, you may also be interested in the following projects:

- [Go ADIF Parser Benchmarks](https://github.com/farmergreg/adif-benchmark)
- [Go ADIF Specification](https://github.com/farmergreg/spec)
- [g3zod/CreateADIFTestFiles](https://github.com/g3zod/CreateADIFTestFiles) ADI Test Files
- [g3zod/CreateADIFExportFiles](https://github.com/g3zod/CreateADIFExportFiles) ADIF Workgroup Specification Export Tool

## 📝 License

This project is licensed under the BSD 3-Clause License - see the [LICENSE](LICENSE) file for details.
