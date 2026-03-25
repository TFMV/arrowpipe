package engine

import "github.com/apache/arrow-go/v18/arrow"

// Record represents a chunk of columnar data.
// A stream of Records forms a chunked, streaming data layout.
type Record struct {
	arrow.Record
}

// NewRecord creates a new Record from an arrow.Record.
func NewRecord(rec arrow.Record) *Record {
	return &Record{rec}
}

// Release releases the memory held by the Record.
func (r *Record) Release() {
	r.Record.Release()
}
