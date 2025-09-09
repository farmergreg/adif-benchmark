# ⚡ World's Fastest ADI Parser for Go

This library provides high-performance processing of [ADIF](https://adif.org/) (Amateur Data Interchange Format) ADI files used for ham radio logs.
It's idiomatic, developer-friendly API seamlessly integrates with your codebase and the go standard library.

[![Tests](https://github.com/hamradiolog-net/adif/actions/workflows/test.yml/badge.svg)](https://github.com/hamradiolog-net/adif/actions/workflows/test.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/hamradiolog-net/adif)](https://goreportcard.com/report/github.com/hamradiolog-net/adif)
[![Go Reference](https://pkg.go.dev/badge/github.com/hamradiolog-net/adif.svg)](https://pkg.go.dev/github.com/hamradiolog-net/adif)
[![Go Version](https://img.shields.io/github/go-mod/go-version/hamradiolog-net/adif)](https://github.com/hamradiolog-net/adif/blob/main/go.mod)
[![License](https://img.shields.io/github/license/hamradiolog-net/adif)](https://github.com/hamradiolog-net/adif/blob/main/LICENSE)

## ✨ Features

- 🔧 **Developer Friendly**: Clean, idiomatic, mock friendly interfaces
- 🚀 **Blazing Fast**: 3-20x faster than other ADI libraries; more than 2x faster than Go's native JSON marshaling!
- 💡 **Memory Efficient**: Up to 1400x fewer memory allocations than alternatives
- 🔬 **Tested**: This library has 100% test coverage!

## 🚀 Quick Start

```bash
go get github.com/hamradiolog-net/adif@latest
```

1) [Example Processing ADI Files](./example_adi_test.go): If you're not sure, use this!
2) ADX XML Processing: Not implemented. PR(s) welcome!
3) [Example Processing ADIJ Files](./example_adij.go): ADIJ is a proposed ADI container format that represents ADIF data as json.

## Benchmarks

TLDR: The ADI parser is very fast, possibly the fastest in the world.
Please see the [Go ADIF Parser Benchmarks](https://github.com/hamradiolog-net/adif-benchmark) project for benchmarks.

## 🔧 Technical Deep Dive (ADI Parser)

The ADI parser achieves high performance through the following optimizations:

### Performance Optimizations

- Leverages stdlib I/O operations with SSE/SIMD acceleration depending upon your CPU architecture
- Smart buffer pre-allocation based on discovered record sizes
- Optimized base-10 integer parsing for ADIF field lengths

### Memory Management

- Zero-copy techniques minimize memory operations
- String interning of repeated field names greatly reduces both allocations and memory use
- Minimal temporary allocations during field parsing
- Dynamic buffer sizing based on learned field counts
- Buffer pooling

## Related Projects

If you found this library useful, you may also be interested in the following projects:

- [Go ADIF Parser Benchmarks](https://github.com/hamradiolog-net/adif-benchmark)
- [Go ADIF Specification](https://github.com/hamradiolog-net/spec)
- [g3zod/CreateADIFTestFiles](https://github.com/g3zod/CreateADIFTestFiles) ADI Test Files
- [g3zod/CreateADIFExportFiles](https://github.com/g3zod/CreateADIFExportFiles) ADIF Workgroup Specification Export Tool

## 📝 License

This project is licensed under the BSD 3-Clause License - see the [LICENSE](LICENSE) file for details.
