package storage

import (
	"io"

	"github.com/apache/arrow-go/v18/arrow"
	"github.com/apache/arrow-go/v18/arrow/ipc"
)

// Writer is an interface for writing Arrow record batches.
type Writer interface {
	Write(arrow.Record) error
	Close() error
}

// NewWriter creates a new Arrow writer.
func NewWriter(w io.Writer, schema *arrow.Schema) (Writer, error) {
	fw, err := ipc.NewFileWriter(w, ipc.WithSchema(schema))
	if err != nil {
		return nil, err
	}
	return &featherWriter{fw}, nil
}

type featherWriter struct {
	w *ipc.FileWriter
}

func (fw *featherWriter) Write(rec arrow.Record) error {
	return fw.w.Write(rec)
}

func (fw *featherWriter) Close() error {
	return fw.w.Close()
}
