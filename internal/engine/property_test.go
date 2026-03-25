package engine

import (
	"testing"

	"github.com/apache/arrow-go/v18/arrow"
	"github.com/apache/arrow-go/v18/arrow/array"
	"github.com/apache/arrow-go/v18/arrow/memory"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

func TestSelect(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("Select should preserve the number of rows", prop.ForAll(
		func(rec arrow.Record) bool {
			rec.Retain()
			defer rec.Release()

			// Select all columns
			var fieldNames []string
			for _, f := range rec.Schema().Fields() {
				fieldNames = append(fieldNames, f.Name)
			}
			selected := Select(rec, fieldNames...)
			defer selected.Release()

			return selected.NumRows() == rec.NumRows()
		},
		genArrowRecord(),
	))

	properties.TestingRun(t)
}

// genArrowRecord creates a generator for arrow.Record objects.
func genArrowRecord() gopter.Gen {
	return gen.Int64Range(1, 100).Map(func(size int64) arrow.Record {
		pool := memory.NewGoAllocator()
		schema := arrow.NewSchema(
			[]arrow.Field{{Name: "f1", Type: arrow.PrimitiveTypes.Int64}}, nil)

		b := array.NewInt64Builder(pool)
		defer b.Release()

		for i := int64(0); i < size; i++ {
			b.Append(i)
		}

		arr := b.NewArray()
		defer arr.Release()
		return array.NewRecord(schema, []arrow.Array{arr}, size)
	})
}
