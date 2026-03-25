package storage

import (
	"io"
	"sync"

	"github.com/apache/arrow-go/v18/arrow"
	"github.com/apache/arrow-go/v18/arrow/ipc"
)

// ColumnarDataset represents a columnar dataset stored in the Arrow IPC file format.
type ColumnarDataset struct {
	rw     io.ReadWriteSeeker
	schema *arrow.Schema

	// For writing
	writer *ipc.FileWriter

	// For reading
	reader     *ipc.FileReader
	nextRecord int

	mutex sync.RWMutex
}

// NewColumnarDataset creates a new ColumnarDataset for writing.
func NewColumnarDataset(rw io.ReadWriteSeeker, schema *arrow.Schema) (*ColumnarDataset, error) {
	w, err := ipc.NewFileWriter(rw, ipc.WithSchema(schema))
	if err != nil {
		return nil, err
	}
	return &ColumnarDataset{
		rw:     rw,
		schema: schema,
		writer: w,
	}, nil
}

// NewColumnarDatasetReader creates a new ColumnarDataset for reading.
func NewColumnarDatasetReader(rs ipc.ReadAtSeeker) (*ColumnarDataset, error) {
	r, err := ipc.NewFileReader(rs)
	if err != nil {
		return nil, err
	}
	return &ColumnarDataset{
		rw:         nil, // This is a reader, so we don't need a writer.
		schema:     r.Schema(),
		reader:     r,
		nextRecord: 0,
	}, nil
}

func (cd *ColumnarDataset) Write(rec arrow.Record) error {
	cd.mutex.Lock()
	defer cd.mutex.Unlock()
	return cd.writer.Write(rec)
}

func (cd *ColumnarDataset) Close() error {
	cd.mutex.Lock()
	defer cd.mutex.Unlock()
	if cd.writer != nil {
		return cd.writer.Close()
	}
	if cd.reader != nil {
		return cd.reader.Close()
	}
	return nil
}
