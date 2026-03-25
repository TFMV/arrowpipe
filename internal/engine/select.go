package engine

import (
	"context"
	"github.com/apache/arrow-go/v18/arrow"
	"github.com/apache/arrow-go/v18/arrow/array"
	"github.com/apache/arrow-go/v18/arrow/compute"
	"github.com/apache/arrow-go/v18/arrow/memory"
)

// Select returns a new record with a subset of columns.
func Select(rec arrow.Record, cols ...string) arrow.Record {
	if len(cols) == 0 {
		rec.Retain()
		return rec
	}

	indices := make([]int, len(cols))
	fields := make([]arrow.Field, len(cols))
	columns := make([]arrow.Array, len(cols))

	for i, col := range cols {
		idx := rec.Schema().FieldIndices(col)[0]
		indices[i] = idx
		fields[i] = rec.Schema().Field(idx)
		columns[i] = rec.Column(idx)
	}

	schema := arrow.NewSchema(fields, nil)
	return array.NewRecord(schema, columns, rec.NumRows())
}

func Filter(rec arrow.Record, filter arrow.Array) (arrow.Record, error) {
	ctx := compute.WithAllocator(context.Background(), memory.DefaultAllocator)
	result, err := compute.Filter(ctx,
		compute.NewDatum(rec),
		compute.NewDatum(filter),
		compute.FilterOptions{})

	if err != nil {
		return nil, err
	}
	defer result.Release()

	return result.(*compute.RecordDatum).Value, nil
}
