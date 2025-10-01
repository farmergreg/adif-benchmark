package adif

import (
	"slices"

	"github.com/farmergreg/spec/v6/adifield"
)

var _ Record = (*adiRecord)(nil)

// adiRecord represents a single ADI record.
// It may be either a Header, or a QSO.
type adiRecord struct {
	Fields   map[adifield.Field]string // map of all fields and their values
	allCache []adifield.Field          // all sorts the keys prior to iterating, this caches that work
}

// NewRecord creates a new Record with the default initial capacity.
func NewRecord() Record {
	return newRecordWithCapacity(-1)
}

// newRecordWithCapacity creates a new adiRecord with a specific initial capacity.
func newRecordWithCapacity(initialCapacity int) *adiRecord {
	if initialCapacity < 1 {
		initialCapacity = 7
	}
	r := adiRecord{
		Fields: make(map[adifield.Field]string, initialCapacity),
	}
	return &r
}

// reset clears the record for reuse.
func (r *adiRecord) reset() {
	clear(r.Fields)
	r.allCache = r.allCache[:0]
}

// Get implements ADIFRecord.Get
func (r *adiRecord) Get(field adifield.Field) string {
	return r.Fields[field]
}

// Set implements ADIFRecord.Set
func (r *adiRecord) Set(field adifield.Field, value string) {
	r.allCache = nil
	r.setInternal(field, value)
}

// setInternal sets the value for a field without modifying the field name or clearing the cache.
// It is used by the parser to avoid unnecessary allocations.
// It assumes the field name is already normalized (lowercase).
func (r *adiRecord) setInternal(field adifield.Field, value string) {
	if value == "" {
		delete(r.Fields, field)
	} else {
		r.Fields[field] = value
	}
}

// All implements ADIFRecord.All
func (r *adiRecord) All() func(func(adifield.Field, string) bool) {
	if r.allCache == nil {
		r.allCache = make([]adifield.Field, 0, len(r.Fields))
		for field := range r.Fields {
			r.allCache = append(r.allCache, field)
		}
		slices.Sort(r.allCache)
	}

	return func(yield func(adifield.Field, string) bool) {
		for _, field := range r.allCache {
			if !yield(field, r.Fields[field]) {
				return
			}
		}
	}
}

// Count implements ADIFRecord.Count
func (r *adiRecord) Count() int {
	return len(r.Fields)
}
