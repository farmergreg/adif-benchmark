package adif

import (
	"github.com/hamradiolog-net/spec/v6/adifield"
)

var _ ADIFRecord = (*adiRecord)(nil)

// adiRecord represents a single ADI record.
// The field keys MUST be in UPPERCASE!
type adiRecord struct {
	r                  map[adifield.ADIField]string
	fields             []adifield.ADIField
	isHeader           bool
	isFieldStructDirty bool
}

// NewADIRecord creates a new Record with the default initial capacity.
func NewADIRecord() *adiRecord {
	return NewADIRecordWithCapacity(-1)
}

// NewADIRecordWithCapacity creates a new Record with a specific initial capacity.
func NewADIRecordWithCapacity(initialCapacity int) *adiRecord {
	if initialCapacity < 1 {
		initialCapacity = 7
	}
	r := adiRecord{
		isFieldStructDirty: true,
		r:                  make(map[adifield.ADIField]string, initialCapacity),
	}
	return &r
}

// IsHeader returns whether the record is a header record.
func (r *adiRecord) IsHeader() bool {
	return r.isHeader
}

// SetIsHeader sets whether the record is a header record.
func (r *adiRecord) SetIsHeader(isHeader bool) {
	r.isHeader = isHeader
}

// Get returns the value for the specified field, or an empty string if the field is not present.
func (r *adiRecord) Get(field adifield.ADIField) string {
	return r.r[field]
}

// Set sets the value for the specified field.
func (r *adiRecord) Set(field adifield.ADIField, value string) {
	r.isFieldStructDirty = true
	if value == "" {
		delete(r.r, field)
		return
	}
	r.r[field] = value
}

func (r *adiRecord) Fields() []adifield.ADIField {
	if !r.isFieldStructDirty {
		return r.fields
	}
	r.isFieldStructDirty = false
	r.fields = make([]adifield.ADIField, 0, len(r.r))
	for k := range r.r {
		r.fields = append(r.fields, k)
	}
	return r.fields
}
