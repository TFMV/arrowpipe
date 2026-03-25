package storage

import (
	"os"
	"testing"

	"github.com/apache/arrow-go/v18/arrow"
	"github.com/apache/arrow-go/v18/arrow/array"
	"github.com/apache/arrow-go/v18/arrow/memory"
	"github.com/stretchr/testify/assert"
)

func TestColumnarDataset_Write(t *testing.T) {
	pool := memory.NewGoAllocator()
	schema := arrow.NewSchema(
		[]arrow.Field{{Name: "f1", Type: arrow.PrimitiveTypes.Int64}}, nil)

	tmpfile, err := os.CreateTemp("", "test-arrow-")
	assert.NoError(t, err)
	defer os.Remove(tmpfile.Name())

	ds, err := NewColumnarDataset(tmpfile, schema)
	assert.NoError(t, err)

	b := array.NewInt64Builder(pool)
	defer b.Release()
	b.AppendValues([]int64{1, 2, 3}, nil)
	arr := b.NewArray()
	defer arr.Release()
	rec := array.NewRecord(schema, []arrow.Array{arr}, 3)
	defer rec.Release()

	err = ds.Write(rec)
	assert.NoError(t, err)
	err = ds.Close()
	assert.NoError(t, err)

	// Now read it back and check
	f, err := os.Open(tmpfile.Name())
	assert.NoError(t, err)
	defer f.Close()

	reader, err := NewColumnarDatasetReader(f)
	assert.NoError(t, err)
	defer reader.Close()

	readRec, err := reader.Read()
	assert.NoError(t, err)
	defer readRec.Release()

	assert.Equal(t, int64(3), readRec.NumRows())
}

func BenchmarkColumnarDataset_Write(b *testing.B) {
	pool := memory.NewGoAllocator()
	schema := arrow.NewSchema(
		[]arrow.Field{{Name: "f1", Type: arrow.PrimitiveTypes.Int64}}, nil)

	tmpfile, err := os.CreateTemp("", "bench-arrow-")
	assert.NoError(b, err)
	defer os.Remove(tmpfile.Name())

	ds, err := NewColumnarDataset(tmpfile, schema)
	assert.NoError(b, err)

	builder := array.NewInt64Builder(pool)
	defer builder.Release()
	builder.AppendValues(make([]int64, 100), nil)
	arr := builder.NewArray()
	defer arr.Release()
	rec := array.NewRecord(schema, []arrow.Array{arr}, 100)
	defer rec.Release()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := ds.Write(rec)
		assert.NoError(b, err)
	}

	err = ds.Close()
	assert.NoError(b, err)
}
