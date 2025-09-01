# ⚡ World's Fastest ADI Parser for Go

This library provides high-performance processing of [ADIF](https://adif.org/) (Amateur Data Interchange Format) ADI files used for ham radio logs.
It's idiomatic, developer-friendly API seamlessly integrates with your codebase and the go standard library.

[![Tests](https://github.com/hamradiolog-net/adif/actions/workflows/test.yml/badge.svg)](https://github.com/hamradiolog-net/adif/actions/workflows/test.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/hamradiolog-net/adif)](https://goreportcard.com/report/github.com/hamradiolog-net/adif)
[![Go Reference](https://pkg.go.dev/badge/github.com/hamradiolog-net/adif.svg)](https://pkg.go.dev/github.com/hamradiolog-net/adif)
[![Go Version](https://img.shields.io/github/go-mod/go-version/hamradiolog-net/adif)](https://github.com/hamradiolog-net/adif/blob/main/go.mod)
[![License](https://img.shields.io/github/license/hamradiolog-net/adif)](https://github.com/hamradiolog-net/adif/blob/main/LICENSE)

## ✨ Features

- 🔧 **Developer Friendly**: Clean, idiomatic Go interfaces
- 🚀 **Blazing Fast**: 3-20x faster than other ADI libraries; more than 2x faster than Go's native JSON marshaling!
- 💡 **Memory Efficient**: Up to 1400x fewer memory allocations than alternatives
- 🔬 **Tested**: Comprehensive test coverage ensures reliability

## 🚀 Quick Start

```bash
go get github.com/hamradiolog-net/adif@latest
```

1) [Document](./example_document_test.go): Complete ADI file operations using `io.Reader`/`io.Writer`
2) [ADIFReader](./example_adireader_test.go): Stream-based parsing of ADI records using `io.Reader`
3) [Record](./example_record_test.go): Single record operations using `io.Reader`/`io.Writer`

## Benchmarks

JSON marshaling is included as a baseline for comparison.

| Benchmark  (AMD Ryzen 9 7950X)             | Iterations | Time/op (ns) | Bytes/op    | Allocs/op   |
|--------------------------------------------|----------:|---------------:|------------:|-----------:|
| ▲ Higher is better / ▼ Lower is better     |         ▲ |              ▼ |           ▼ |          ▼ |
| **Read Operations**                        |           |                |             |            |
| This Library                               | **1,461** |    **819,922** |   673,421   | **8,757**  |
| JSON                                       |     622   |    1,915,720   | **402,803** |   25,601   |
| Matir                                      |     417   |    2,895,274   | 2,037,004   |   66,535   |
| Eminlin                                    |      68   |   16,453,839   |13,127,877   |  193,083   |
| **Write Operations**                       |           |                |             |            |
| This Library                               | **1,800** |    **666,157** | **514,418** |     **20** |
| JSON                                       |     796   |    1,488,265   |   966,487   |   17,805   |
| Matir                                      |     399   |    2,994,459   | 1,490,840   |   28,673   |
| Eminlin                                    |     N/A   |          N/A   |       N/A   |      N/A   |

## 🔧 Technical Deep Dive

This parser achieves high performance through the following optimizations:

### Performance Optimizations

- Leverages stdlib I/O operations with SSE/SIMD acceleration depending upon your CPU architecture
- Smart buffer pre-allocation based on discovered record sizes
- Optimized ASCII case conversion using bitwise operations
- Optimized base-10 integer parsing for ADIF field lengths

### Memory Management

- Zero-copy techniques minimize memory operations
- String interning of repeated field names greatly reduces allocations and memory use
- Minimal temporary allocations during field parsing
- Dynamic buffer sizing based on learned field counts

## Related Projects

If you found this library useful, you may also be interested in the following projects:

- [Go ADIF Parser Benchmarks](https://github.com/hamradiolog-net/adif-benchmark)
- [Go ADIF Specification](https://github.com/hamradiolog-net/adif-spec)
- [g3zod/CreateADIFTestFiles](https://github.com/g3zod/CreateADIFTestFiles) ADI Test Files
- [g3zod/CreateADIFExportFiles](https://github.com/g3zod/CreateADIFExportFiles) ADIF Workgroup Specification Export Tool

## 📝 License

This project is licensed under the BSD 3-Clause License - see the [LICENSE](LICENSE) file for details.
