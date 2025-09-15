package adif

import (
	"github.com/hamradiolog-net/spec/v6/adifield"
)

// adifDocument represents an ADIF document.
// This may be used directly with the encoding/json package to marshal or unmarshal ADIJ (ADIF as JSON) data.
type adifDocument struct {
	// Header is nil when there is no header.
	// Otherwise it is a Record with header fields inside.
	Header map[adifield.Field]string `json:"HEADER,omitempty"`

	// Records is a slice of Record.
	// It contains zero or more QSO records.
	Records []map[adifield.Field]string `json:"RECORDS"`
}
