package adif

import (
	"bufio"
	"bytes"
	"io"
	"strings"
	"unsafe"

	"github.com/farmergreg/spec/v6/adifield"
)

var _ RecordReader = (*adiReader)(nil)

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
	appFieldMap map[string]adifield.Field

	// bufValue is a reusable buffer used to temporarily store the VALUE of the current field.
	bufValue []byte

	// preAllocateFields is the number of fields to allocate for each record.
	preAllocateFields int

	// skipHeader is true if the header record should be skipped.
	skipHeader bool
}

// NewADIRecordReader returns an ADIFReader that can parse ADIF *.adi formatted records.
// If skipHeader is true, Next() will not return the header record if it exists.
// This is a streaming parser that processes the input as it is read, using minimal memory.
func NewADIRecordReader(r io.Reader, skipHeader bool) *adiReader {
	br, ok := r.(*bufio.Reader)
	if !ok {
		br = bufio.NewReader(r)
	}

	p := &adiReader{
		r:          br,
		skipHeader: skipHeader,
	}
	p.appFieldMap = make(map[string]adifield.Field, 128)
	p.bufValue = make([]byte, 4096)

	return p
}

// Next reads and returns the next Record.
// It returns io.EOF when no more records are available.
func (p *adiReader) Next() (Record, bool, error) {
	result := newRecordWithCapacity(p.preAllocateFields)
	for {
		// Find the start of the next adi field
		err := p.discardUntilLessThan()
		if err != nil {
			return nil, false, err
		}

		field, value, err := p.parseOneField()
		if err != nil {
			return nil, false, err
		}

		switch field {
		case adifield.EOR:
			if count := result.Count(); count > p.preAllocateFields {
				p.preAllocateFields = count
			}
			return result, false, nil
		case adifield.EOH:
			if !p.skipHeader {
				return result, true, nil
			}
			// we are skipping returning the EOH record (if any)
			// reset to prepare to read the next record
			result.reset()
		}

		// n.b. if a duplicate field is found, it will replace the previous value
		result.setInternal(field, value)
	}
}

// parseOneField reads the next field definition and returns the field name and value
//
// It is heavily optimized for speed and memory use.
// Currently, It can tripple the speed of go's stdlib JSON marshaling for similar data.
func (p *adiReader) parseOneField() (field adifield.Field, value string, err error) {
	// Step 1: Finish reading the data specifier "<fieldname:length:...>", removing the trailing '>'
	volatileSpecifier, err := p.readDataSpecifierVolatile()
	if err != nil {
		return "", "", err
	}

	// Step 2: Parse Field Name
	volatileField, volatileLength, foundFirstColon := bytes.Cut(volatileSpecifier, []byte(":"))
	if len(volatileField) == 0 {
		return "", "", ErrAdiReaderMalformedADI // field name is empty
	}

	// Step 2.1: field name string interning
	var ok bool
	fieldStringUnsafe := unsafe.String(&volatileField[0], len(volatileField))
	if field, ok = p.appFieldMap[fieldStringUnsafe]; !ok {
		fieldStringSafe := strings.Clone(fieldStringUnsafe)
		field = adifield.New(fieldStringSafe)
		p.appFieldMap[fieldStringSafe] = field
	}

	if !foundFirstColon {
		// EOH, EOR
		// And, also, LoTW's deviation from the official spec: APP_LOTW_EOF
		return field, "", nil
	}
	// Step 3: Parse Field Length
	if idx := len(volatileLength) - 2; idx > 0 && volatileLength[idx] == ':' {
		// We assume that data type indicators are exactly 1 character long.
		volatileLength = volatileLength[:idx]
	}

	length, err := parseDataLength(volatileLength)
	if err != nil {
		// handle data length parsing errors
		return "", "", err
	}
	if length < 1 {
		return field, "", nil
	}

	// Step 4: Read the field value
	// ParseDataLength ensures that length is a reasonable value for us.
	if cap(p.bufValue) < length {
		p.bufValue = make([]byte, length)
	}
	p.bufValue = p.bufValue[:length]

	var c int
	c, err = io.ReadFull(p.r, p.bufValue) // this will overwrite all of the 'volatile' variables (see above)
	if err == io.EOF {
		return "", "", ErrAdiReaderMalformedADI
	}
	value = string(p.bufValue[:c])
	return field, value, nil
}

// readDataSpecifierVolatile reads and returns the next data specifier as a byte slice, and any error encountered.
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
func (p *adiReader) readDataSpecifierVolatile() (volatileSpecifier []byte, err error) {
	// If ReadSlice returns bufio.ErrBufferFull, accumulator will contain ALL of the bytes read.
	// In most cases, accumulator will be null because we won't hit the bufio.ErrBufferFull condition.
	var accumulator []byte
	for {
		volatileSpecifier, err = p.r.ReadSlice('>')
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
			return volatileSpecifier, ErrAdiReaderMalformedADI
		}

		return volatileSpecifier, err
	}
	volatileSpecifier = volatileSpecifier[:len(volatileSpecifier)-1] // remove the trailing '>'
	return volatileSpecifier, nil
}

// discardUntilLessThan reads until it finds the '<' character
func (p *adiReader) discardUntilLessThan() (err error) {
	for {
		_, err = p.r.ReadSlice('<')

		switch err {
		case nil:
			return nil
		case bufio.ErrBufferFull:
			continue
		default:
			return err
		}
	}
}

// parseDataLength is an optimized replacement for strconv.Atoi.
func parseDataLength(data []byte) (value int, err error) {
	// prevent overflow; int32 max is 2,147,483,647 (10 digits)
	// we limit ourselves to 1,000,000,000 maximum
	if count := len(data); count == 0 || count > 10 {
		return 0, ErrAdiReaderMalformedADI
	}

	for _, b := range data {
		if b < '0' || b > '9' {
			return 0, ErrAdiReaderMalformedADI
		}
		value = value*10 + int(b-'0') // Parse digit, avoiding string allocations
	}

	if value > maxADIReaderDataSize {
		return 0, ErrAdiReaderMalformedADI
	}

	return value, nil
}
