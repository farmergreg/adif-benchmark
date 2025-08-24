// Package adif implements a high performance ADIF library for Go.
// It provides types, structs and methods for managing ADIF Records.
// Idiomatic interfaces for reading and writing ADI formatted data
// make integration with other Go libraries fast and easy.
package adif

const (
	TagEOH = string("<EOH>") // TagEOH is the end of header ADIF tag: <EOH>
	TagEOR = string("<EOR>") // TagEOR is the end of record ADIF tag: <EOR>
)

const adiHeaderPreamble = "K9CTS AM✠DG ADIF Library\n"
