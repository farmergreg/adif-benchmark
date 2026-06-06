package adif

import (
	"bufio"
	"bytes"
	"io"
	"strings"
	"unsafe"

	"github.com/farmergreg/spec/v6/adifield"
)

// scannerArenaChunkSize is the size of each value arena chunk.
// Field values are copied into the current chunk and referenced as strings without per-value allocation.
// A new chunk is allocated only when the next value does not fit, so committed bytes are never moved or mutated.
const scannerArenaChunkSize = 16384

// Scanner reads ADIF *.adi records sequentially from an io.Reader.
// Use NewScanner to create one, then call Scan in a loop.
// It follows the same pattern as bufio.Scanner.
//
//	s := adif.NewScanner(r)
//	for s.Scan() {
//	    if s.IsHeader() {
//	        // process header fields
//	        continue
//	    }
//	    record := s.Record()
//	}
//	if err := s.Err(); err != nil { ... }
type Scanner struct {
	r                 *bufio.Reader
	appFieldMap       map[string]adifield.Field
	arena             []byte
	preAllocateFields int
	current           Record
	isHeader          bool
	err               error
}

// NewScanner returns a Scanner that reads ADI records from r.
func NewScanner(r io.Reader) *Scanner {
	br, ok := r.(*bufio.Reader)
	if !ok {
		br = bufio.NewReader(r)
	}
	return &Scanner{
		r:                 br,
		preAllocateFields: 7,
		appFieldMap:       make(map[string]adifield.Field, 128),
		arena:             make([]byte, 0, scannerArenaChunkSize),
	}
}

// Scan advances to the next record and returns true if one was found.
// Call Record to retrieve it and IsHeader to determine its type.
// Returns false when no more records exist or an error occurred.
// After Scan returns false, call Err to check for any non-EOF error.
func (s *Scanner) Scan() bool {
	s.current, s.isHeader, s.err = s.next()
	return s.err == nil
}

// Record returns the record from the most recent successful Scan call.
func (s *Scanner) Record() Record { return s.current }

// IsHeader reports whether the record from the most recent Scan call is a header record.
func (s *Scanner) IsHeader() bool { return s.isHeader }

// Err returns the first non-EOF error encountered by the Scanner.
// Returns nil when Scan stopped due to io.EOF.
func (s *Scanner) Err() error {
	if s.err == io.EOF {
		return nil
	}
	return s.err
}

// next reads the next record from the underlying reader.
// It returns the record along with a boolean indicating whether it's a header record.
// It returns an error if the ADI is malformed or an I/O error occurs.
func (s *Scanner) next() (Record, bool, error) {
	result := make(Record, s.preAllocateFields)
	for {
		if err := s.discardUntilLessThan(); err != nil {
			return nil, false, err
		}

		field, value, err := s.parseOneField()
		if err != nil {
			return nil, false, err
		}

		switch field {
		case adifield.EOR:
			s.preAllocateFields = len(result)
			return result, false, nil
		case adifield.EOH:
			return result, true, nil
		}

		if value != "" {
			result[field] = value
		}
	}
}

// parseOneField reads the next field specifier and value from the underlying reader.
// It is heavily optimized for speed and memory use.
func (s *Scanner) parseOneField() (field adifield.Field, value string, err error) {
	// Step 1: Read "<fieldname:length:type>" removing the trailing '>'.
	volatileSpecifier, err := s.readDataSpecifierVolatile()
	if err != nil {
		return "", "", err
	}

	// Step 2: Split on the first colon to get field name and length.
	volatileField, volatileLength, foundFirstColon := bytes.Cut(volatileSpecifier, []byte(":"))
	if len(volatileField) == 0 {
		return "", "", ErrMalformedADI
	}

	// Step 2.1: Intern the field name string to avoid repeated allocations.
	var ok bool
	fieldStringUnsafe := unsafe.String(&volatileField[0], len(volatileField))
	if field, ok = s.appFieldMap[fieldStringUnsafe]; !ok {
		if len(s.appFieldMap) > 1024 {
			return "", "", ErrTooManyUniqueFields
		}
		fieldStringSafe := strings.Clone(fieldStringUnsafe)
		field = adifield.New(fieldStringSafe)
		s.appFieldMap[fieldStringSafe] = field
	}

	if !foundFirstColon {
		// EOH, EOR, and LoTW's non-standard APP_LOTW_EOF all lack a colon.
		return field, "", nil
	}

	// Step 3: Strip optional single-character data type indicator (e.g. ":S").
	if idx := len(volatileLength) - 2; idx > 0 && volatileLength[idx] == ':' {
		volatileLength = volatileLength[:idx]
	}

	// Step 4: Parse the field length.
	length, err := parseDataLength(volatileLength)
	if err != nil {
		return "", "", err
	}
	if length < 1 {
		return field, "", nil
	}

	// Step 5: Read exactly length bytes of field value into the arena.
	// Values are referenced as strings pointing into the arena, avoiding a per-value allocation.
	// A fresh chunk is allocated only when the value does not fit, so previously committed bytes never move.
	if cap(s.arena)-len(s.arena) < length {
		chunkSize := scannerArenaChunkSize
		if length > chunkSize {
			chunkSize = length
		}
		s.arena = make([]byte, 0, chunkSize)
	}
	start := len(s.arena)

	c, err := io.ReadFull(s.r, s.arena[start:start+length])
	if c != length {
		return "", "", ErrMalformedADI
	}
	s.arena = s.arena[:start+length]
	return field, unsafe.String(&s.arena[start], length), err
}

// readDataSpecifierVolatile reads up to and including the next '>' and returns
// the bytes between the already-consumed '<' and '>'.
//
// IMPORTANT: The returned slice is VOLATILE and will be invalidated by the next
// read from the underlying bufio.Reader. Callers must not retain it.
//
// Per the ADIF spec, data specifiers have the form:
//
//	<F:L:T>
//
// where F is the field name, L is the data length, and T is an optional
// single-character data type indicator.
func (s *Scanner) readDataSpecifierVolatile() ([]byte, error) {
	// In the common case ReadSlice returns immediately with no ErrBufferFull,
	// so accumulator stays nil and no extra allocation occurs.
	var accumulator []byte
	for {
		volatile, err := s.r.ReadSlice('>')
		if err == nil {
			if accumulator != nil {
				volatile = append(accumulator, volatile...)
			}
			return volatile[:len(volatile)-1], nil // strip trailing '>'
		}
		if err == bufio.ErrBufferFull {
			accumulator = append(accumulator, volatile...)
			continue
		}
		if err == io.EOF {
			return volatile, ErrMalformedADI
		}
		return volatile, err
	}
}

// discardUntilLessThan reads and discards bytes until '<' is found.
func (s *Scanner) discardUntilLessThan() error {
	_, err := s.r.ReadSlice('<')
	for err == bufio.ErrBufferFull {
		_, err = s.r.ReadSlice('<')
	}
	return err
}

// parseDataLength converts an ASCII decimal byte slice to an int.
// It is an optimized, allocation-free replacement for strconv.Atoi.
func parseDataLength(data []byte) (int, error) {
	// Limit to 9 digits (max 999,999,999) to prevent overflow.
	if count := len(data); count == 0 || count > 9 {
		return 0, ErrMalformedADI
	}
	value := 0
	hasValidDigits := true
	for _, b := range data {
		hasValidDigits = hasValidDigits && b >= '0' && b <= '9'
		value = value*10 + int(b-'0')
	}
	if !hasValidDigits {
		return 0, ErrMalformedADI
	}
	return value, nil
}
