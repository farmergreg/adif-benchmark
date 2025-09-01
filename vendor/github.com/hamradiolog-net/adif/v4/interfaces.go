package adif

// ADIFReader processes Amateur Data Interchange Format (ADIF) records sequentially.
type ADIFReader interface {

	// Next reads and returns the next Record in the input.
	// It returns io.EOF when no more records are available.
	Next() (record Record, isHeader bool, bytesRead int64, err error)
}
