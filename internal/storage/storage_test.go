package storage

import (
	"context"
	"os"
	"testing"

	"github.com/TFMV/arrowpipe/internal/engine"
	"github.com/apache/arrow-go/v18/arrow"
	"github.com/apache/arrow-go/v18/arrow/array"
	"github.com/apache/arrow-go/v18/arrow/memory"
	"github.com/stretchr/testify/assert"
)

func TestColumnarDataset_WriteAndRead(t *testing.T) {
	pool := memory.NewGoAllocator()
	schema := arrow.NewSchema(
		[]arrow.Field{
			{Name: "f1-i64", Type: arrow.PrimitiveTypes.Int64},
			{Name: "f2-f64", Type: arrow.PrimitiveTypes.Float64},
			{Name: "f3-str", Type: arrow.BinaryTypes.String},
		},
		nil,
	)

	ib := array.NewInt64Builder(pool)
	ib.AppendValues([]int64{1, 2, 3}, nil)
	i1 := ib.NewArray()
	defer i1.Release()

	fb := array.NewFloat64Builder(pool)
	fb.AppendValues([]float64{1.1, 2.2, 3.3}, nil)
	f1 := fb.NewArray()
	defer f1.Release()

	sb := array.NewStringBuilder(pool)
	sb.AppendValues([]string{"a", "b", "c"}, nil)
	s1 := sb.NewArray()
	defer s1.Release()

	rec := array.NewRecord(schema, []arrow.Array{i1, f1, s1}, 3)
	defer rec.Release()

	tmpfile, err := os.CreateTemp("", "test_arrow_ipc")
	assert.NoError(t, err)
	defer os.Remove(tmpfile.Name())

	writer, err := NewColumnarDataset(tmpfile, schema)
	assert.NoError(t, err)
	err = writer.Write(rec)
	assert.NoError(t, err)
	err = writer.Close()
	assert.NoError(t, err)

	// Open for reading
	f, err := os.Open(tmpfile.Name())
	assert.NoError(t, err)
	defer f.Close()

	reader, err := NewColumnarDatasetReader(f)
	assert.NoError(t, err)

	// Verify schema
	assert.True(t, schema.Equal(reader.Schema()), "schemas do not match")
}

func TestStreamingPipeline(t *testing.T) {
	mem := memory.NewGoAllocator()
	schema := arrow.NewSchema(
		[]arrow.Field{
			{Name: "f1-i64", Type: arrow.PrimitiveTypes.Int64},
			{Name: "f2-f64", Type: arrow.PrimitiveTypes.Float64},
			{Name: "f3-str", Type: arrow.BinaryTypes.String},
		},
		nil,
	)

	ib := array.NewInt64Builder(mem)
	ib.AppendValues([]int64{1, 2, 3, 4, 5}, nil)
	i1 := ib.NewArray()
	defer i1.Release()

	fb := array.NewFloat64Builder(mem)
	fb.AppendValues([]float64{1.1, 2.2, 3.3, 4.4, 5.5}, nil)
	f1 := fb.NewArray()
	defer f1.Release()

	sb := array.NewStringBuilder(mem)
	sb.AppendValues([]string{"a", "b", "c", "d", "e"}, nil)
	s1 := sb.NewArray()
	defer s1.Release()

	rec := array.NewRecord(schema, []arrow.Array{i1, f1, s1}, 5)
	defer rec.Release()

	t.Run("Project and Filter", func(t *testing.T) {
		in := make(chan arrow.Record, 1)
		in <- rec
		close(in)

		sp := engine.NewStreamingPipeline(in)
		out := sp.Project([]string{"f1-i64", "f3-str"}).Filter("f1-i64 > 2").Execute(context.Background())

		for res := range out {
			assert.Equal(t, int64(3), res.NumRows())
			assert.Equal(t, int64(2), res.NumCols())
			assert.Equal(t, "f1-i64", res.Schema().Field(0).Name)
			assert.Equal(t, "f3-str", res.Schema().Field(1).Name)
		}
	})
}
