package adif

import (
	"github.com/farmergreg/spec/v6/adifield"
)

// jsonDocument represents an ADIF document in a JSON container.
type jsonDocument struct {
	Header  map[adifield.Field]string   `json:"header,omitempty"`
	Records []map[adifield.Field]string `json:"records"`
}
