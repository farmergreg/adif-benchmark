package adif

import (
	"bufio"
	"bytes"
	"io"
	"strings"
	"unsafe"

	"github.com/hamradiolog-net/adif-spec/v6/adifield"
)

var _ ADIFReader = (*adiReader)(nil) // Implements ADIFReader

const (
	// 1MB - this is the maximum size of a field value that we will accept.
	// This is intended to be a generous limit for most applications while providing some protection against malformed and/or malicious input.
	//
	// The data is part of the ADIF "Data-Specifier."
	// Per the ADIF spec:
	//   ADI-exporting applications can place as much data in a Data-Specifier as they choose.
	//   ADI-importing applications can import as much data from a Data-Specifier as they choose.
	maxADIReaderDataSize = 1024 * 1024 * 1
)

// adiReader is a high-performance ADIF Reader that can parse ADIF *.adi formatted records.
type adiReader struct {
	r *bufio.Reader

	// appFieldMap is a map of field names used to reduce allocations via string interning.
	appFieldMap map[string]adifield.ADIField

	// bufValue is a reusable buffer used to temporarily store the VALUE of the current field.
	bufValue []byte

	// preAllocateFields is the number of fields to allocate for each record.
	preAllocateFields int

	// skipHeader is true if the header record should be skipped.
	skipHeader bool
}

// NewADIReader returns an ADIFReader that can parse ADIF *.adi formatted records.
// If skipHeader is true, Next() will not return the header record if it exists.
func NewADIReader(r io.Reader, skipHeader bool) ADIFReader {
	br, ok := r.(*bufio.Reader)
	if !ok {
		br = bufio.NewReader(r)
	}

	p := &adiReader{
		r:          br,
		skipHeader: skipHeader,
	}
	p.appFieldMap = make(map[string]adifield.ADIField, 128)

	return p
}

// Next reads and returns the next Record.
// It returns io.EOF when no more records are available.
func (p *adiReader) Next() (Record, bool, int64, error) {
	result := NewRecordWithCapacity(p.preAllocateFields)
	var n int64
	for {
		// Find the start of the next adi field
		c, err := p.discardUntilLessThan()
		n += c
		if err != nil {
			return result, false, n, err
		}

		field, value, c, err := p.parseOneField()
		n += c
		if err != nil {
			return result, false, n, err
		}

		switch field {
		case adifield.EOH:
			if len(result) > 0 {
				if !p.skipHeader {
					return result, true, n, nil
				}

				// we are skipping returning the EOH record (if any)
				// reset to prepare to read the next record
				result.Reset()
			}
			continue
		case adifield.EOR:
			if len(result) > 0 {
				if len(result) > p.preAllocateFields {
					p.preAllocateFields = len(result)
				}
				return result, false, n, nil
			}
			// we know record is empty... no need to reset it
			continue
		}

		// n.b. if a duplicate field is found, it will replace the previous value
		result[field] = value
	}
}

// parseOneField reads the next field definition and returns the field name, value, and the number of bytes read.
//
// It is heavily optimized for speed and memory use.
// Currently, It can tripple the speed of go's stdlib JSON marshaling for similar data.
//
// Future Plans: I would like to take a look at using simd directly.
// However, the current implementation IS attempting to take advantage of the standard library's existing simd capabilities.
func (p *adiReader) parseOneField() (field adifield.ADIField, value string, n int64, err error) {
	// Step 1: Read in the entire data specifier "<fieldname:length:...>" and remove the trailing '>'
	volatileSpecifier, n, err := p.readDataSpecifierVolatile()
	if err != nil {
		return "", "", n, err
	}

	// Step 2: Parse Field Name
	volatileField, volatileLength, foundFirstColon := bytes.Cut(volatileSpecifier, []byte(":"))
	if len(volatileField) == 0 {
		return "", "", n, ErrMalformedADI // field name is empty
	}

	// Step 2.1: field name string interning - reduce memory allocations
	var ok bool
	fieldStringUnsafe := unsafe.String(&volatileField[0], len(volatileField))
	if field, ok = p.appFieldMap[fieldStringUnsafe]; !ok {
		fieldStringSafe := strings.Clone(fieldStringUnsafe)
		field = adifield.ADIField(strings.ToUpper(fieldStringSafe))

		// The key is the original, non-forced-uppercase value because we assume we'll see it repeatedly.
		// It will always need the same transformation (if any) applied.
		p.appFieldMap[fieldStringSafe] = field
	}

	// Step 3: Parse Field Length
	var length int
	if foundFirstColon {
		// look for the second colon which if present, indicates an adi optional type
		endIdx := bytes.IndexByte(volatileLength, ':')
		if endIdx == -1 {
			length, err = parseDataLength(volatileLength)
		} else {
			// we have a data type indicator.
			// this parser doesn't support it; ignore it
			length, err = parseDataLength(volatileLength[:endIdx])
		}
		if err != nil {
			// handle data length parsing errors
			return field, "", n, err
		}

		// Step 4: Read the field value (if any)
		// ParseDataLength ensures that length is a reasonable value for us.
		// inlining v.s. a function call gains a tiny, but measurable amount of performance...
		if length > 0 {
			if cap(p.bufValue) < length {
				p.bufValue = make([]byte, length)
			}
			p.bufValue = p.bufValue[:length]

			var c int
			c, err = io.ReadFull(p.r, p.bufValue) // this will overwrite all of the 'volatile' variables (see above)
			n += int64(c)
			value = string(p.bufValue[:c])
			if err == io.EOF {
				return field, value, n, ErrMalformedADI
			}
			return field, value, n, err
		}
	}

	return field, "", n, nil
}

// readDataSpecifierVolatile reads and returns the next data specifier as a byte slice, the number of bytes read, and any error encountered.
// The trailing '>' is removed from the returned byte slice.
//
// IMPORTANT:
// The returned byte slice is VOLATILE and will be invalidated by the next read from the underlying bufio.Reader.
//
// Per the Spec:
//
//	Data-Specifiers used to convey data in an ADI file are composed of a case-independent
//	field name F, a data length L, and an optional data type indicator T separated by colons and enclosed in angle brackets,
//	followed by data D of length L:
//
//	<F:L:T>
func (p *adiReader) readDataSpecifierVolatile() (volatileSpecifier []byte, n int64, err error) {
	// If ReadSlice returns bufio.ErrBufferFull, accumulator will contain ALL of the bytes read.
	// In most cases, accumulator will be null because we won't hit the bufio.ErrBufferFull condition.
	var accumulator []byte
	for {
		volatileSpecifier, err = p.r.ReadSlice('>')
		n += int64(len(volatileSpecifier))
		if err == nil {
			if accumulator != nil {
				// We've found '>' and have accumulated some bytes.
				// Update volatileSpecifier to point at the accumulated bytes before breaking out of the loop.
				volatileSpecifier = append(accumulator, volatileSpecifier...)
			}
			break
		}

		if err == bufio.ErrBufferFull {
			accumulator = append(accumulator, volatileSpecifier...)
			continue
		}

		if err == io.EOF {
			return volatileSpecifier, n, ErrMalformedADI
		}

		return volatileSpecifier, n, err
	}
	volatileSpecifier = volatileSpecifier[:len(volatileSpecifier)-1] // remove the trailing '>'
	return volatileSpecifier, n, nil
}

// discardUntilLessThan reads until it finds the '<' character, returning the number of bytes read
func (p *adiReader) discardUntilLessThan() (n int64, err error) {
	for {
		var b []byte
		b, err = p.r.ReadSlice('<')
		n += int64(len(b))

		switch err {
		case nil:
			return n, nil
		case bufio.ErrBufferFull:
			continue
		default:
			return n, err
		}
	}
}

// parseDataLength is an optimized replacement for strconv.Atoi.
func parseDataLength(data []byte) (value int, err error) {
	if len(data) == 0 {
		return 0, ErrInvalidFieldLength
	}

	for _, b := range data {
		if b < '0' || b > '9' {
			return 0, ErrInvalidFieldLength
		}

		// Parse digit, avoiding string allocations
		newVal := value*10 + int(b-'0')

		// Check for overflow or too big
		if newVal < value || newVal > maxADIReaderDataSize {
			return 0, ErrInvalidFieldLength
		}

		value = newVal
	}

	return value, nil
}
