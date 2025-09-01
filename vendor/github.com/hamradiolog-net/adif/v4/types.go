package adif

import (
	"github.com/hamradiolog-net/adif-spec/v6/adifield"
)

// Document represents a complete ADIF document.
//
// Future Work:
// This type intentionally resembles the ADX XML structure even though XML is not currently supported by this library.
type Document struct {
	// Header is nil when there is no header.
	// Otherwise it will be a Record with header fields inside.
	Header Record `json:"HEADER,omitempty"`

	// Records is a slice of Record.
	// It contains the QSO records.
	Records []Record `json:"RECORDS"`

	headerPreamble string
}

// Record is a map of ADIF fields to their values, representing EITHER a Header record or a QSO record.
// The field keys must be UPPERCASE strings of type adifield.Field.
type Record map[adifield.ADIField]string
