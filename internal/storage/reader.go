package storage

import (
	"bytes"
	"fmt"
	"io"

	"github.com/apache/arrow-go/v18/arrow"
	"github.com/apache/arrow-go/v18/arrow/ipc"
)

// OpenForReading initializes the dataset for reading from an Arrow IPC file.
func OpenForReading(r ipc.ReadAtSeeker) (*ColumnarDataset, error) {
	reader, err := ipc.NewFileReader(r)
	if err != nil {
		return nil, err
	}

	return &ColumnarDataset{reader: reader}, nil
}

// OpenForReadingFromPipe initializes the dataset for reading from an Arrow IPC stream.
func OpenForReadingFromPipe(r io.Reader) (*ColumnarDataset, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	buf := bytes.NewReader(data)
	return OpenForReading(buf)
}

// Schema returns the schema of the dataset.
func (cd *ColumnarDataset) Schema() *arrow.Schema {
	if cd.reader == nil {
		return nil
	}
	return cd.reader.Schema()
}

// NumChunks returns the number of chunks in the dataset.
func (cd *ColumnarDataset) NumChunks() int {
	if cd.reader == nil {
		return 0
	}
	return cd.reader.NumRecords()
}

// Chunk returns a specific chunk by its index.
func (cd *ColumnarDataset) Chunk(i int) (arrow.Record, error) {
	cd.mutex.RLock()
	defer cd.mutex.RUnlock()

	if cd.reader == nil {
		return nil, fmt.Errorf("dataset not open for reading")
	}

	if i < 0 || i >= cd.reader.NumRecords() {
		return nil, fmt.Errorf("chunk index out of bounds")
	}

	// Retain the record to manage its lifecycle outside the reader
	rec, err := cd.reader.RecordAt(i)
	if err != nil {
		return nil, err
	}
	rec.Retain()
	return rec, nil
}

func (cd *ColumnarDataset) Read() (arrow.Record, error) {
	cd.mutex.RLock()
	defer cd.mutex.RUnlock()

	if cd.reader == nil {
		return nil, fmt.Errorf("dataset not open for reading")
	}

	if cd.nextRecord >= cd.reader.NumRecords() {
		return nil, io.EOF
	}

	rec, err := cd.reader.RecordAt(cd.nextRecord)
	if err != nil {
		return nil, err
	}
	rec.Retain()
	cd.nextRecord++
	return rec, nil
}
