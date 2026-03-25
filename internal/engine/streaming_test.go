package engine

import (
	"context"
	"testing"

	"github.com/apache/arrow-go/v18/arrow"
	"github.com/apache/arrow-go/v18/arrow/array"
	"github.com/apache/arrow-go/v18/arrow/memory"
)

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
	ib.AppendValues([]int64{1, 2, 3}, nil)
	i1 := ib.NewArray()
	defer i1.Release()

	fb := array.NewFloat64Builder(mem)
	fb.AppendValues([]float64{1.1, 2.2, 3.3}, nil)
	f1 := fb.NewArray()
	defer f1.Release()

	sb := array.NewStringBuilder(mem)
	sb.AppendValues([]string{"a", "b", "c"}, nil)
	s1 := sb.NewArray()
	defer s1.Release()

	rec := array.NewRecord(schema, []arrow.Array{i1, f1, s1}, 3)
	defer rec.Release()

	t.Run("Project", func(t *testing.T) {
		in := make(chan arrow.Record, 1)
		in <- rec
		close(in)

		sp := NewStreamingPipeline(in)
		out := sp.Project([]string{"f1-i64", "f3-str"}).Execute(context.Background())

		for res := range out {
			if res.NumCols() != 2 {
				t.Errorf("expected 2 columns, got %d", res.NumCols())
			}
			if res.Schema().Field(0).Name != "f1-i64" {
				t.Errorf("expected column f1-i64, got %s", res.Schema().Field(0).Name)
			}
			if res.Schema().Field(1).Name != "f3-str" {
				t.Errorf("expected column f3-str, got %s", res.Schema().Field(1).Name)
			}
		}
	})
}
