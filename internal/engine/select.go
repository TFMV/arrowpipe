package engine

import (
	"fmt"

	"github.com/apache/arrow-go/v18/arrow"
	"github.com/apache/arrow-go/v18/arrow/array"
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
	boolFilter, ok := filter.(*array.Boolean)
	if !ok {
		return nil, fmt.Errorf("filter must be a boolean array")
	}

	mem := memory.NewGoAllocator()
	numRows := 0
	for i := 0; i < boolFilter.Len(); i++ {
		if !boolFilter.IsNull(i) && boolFilter.Value(i) {
			numRows++
		}
	}

	if numRows == 0 {
		return nil, fmt.Errorf("filter resulted in zero rows")
	}

	cols := make([]arrow.Array, rec.NumCols())
	fields := make([]arrow.Field, rec.NumCols())

	for i, col := range rec.Columns() {
		fields[i] = rec.Schema().Field(i)
		bldr := array.NewBuilder(mem, col.DataType())
		defer bldr.Release()

		switch b := bldr.(type) {
		case *array.Int64Builder:
			typedCol := col.(*array.Int64)
			for j := 0; j < typedCol.Len(); j++ {
				if !boolFilter.IsNull(j) && boolFilter.Value(j) {
					b.Append(typedCol.Value(j))
				}
			}
		case *array.Float64Builder:
			typedCol := col.(*array.Float64)
			for j := 0; j < typedCol.Len(); j++ {
				if !boolFilter.IsNull(j) && boolFilter.Value(j) {
					b.Append(typedCol.Value(j))
				}
			}
		case *array.StringBuilder:
			typedCol := col.(*array.String)
			for j := 0; j < typedCol.Len(); j++ {
				if !boolFilter.IsNull(j) && boolFilter.Value(j) {
					b.Append(typedCol.Value(j))
				}
			}
		case *array.BooleanBuilder:
			typedCol := col.(*array.Boolean)
			for j := 0; j < typedCol.Len(); j++ {
				if !boolFilter.IsNull(j) && boolFilter.Value(j) {
					b.Append(typedCol.Value(j))
				}
			}
		default:
			return nil, fmt.Errorf("unsupported column type: %s", col.DataType())
		}
		cols[i] = bldr.NewArray()
		defer cols[i].Release()
	}

	newSchema := arrow.NewSchema(fields, nil)
	return array.NewRecord(newSchema, cols, int64(numRows)), nil
}
