package main

import (
	"log"
	"os"

	"github.com/TFMV/arrowpipe/internal/storage"
	"github.com/apache/arrow-go/v18/arrow"
	"github.com/apache/arrow-go/v18/arrow/array"
	"github.com/apache/arrow-go/v18/arrow/memory"
)

func main() {
	pool := memory.NewGoAllocator()
	schema := arrow.NewSchema(
		[]arrow.Field{
			{Name: "f1-i64", Type: arrow.PrimitiveTypes.Int64},
			{Name: "f2-f64", Type: arrow.PrimitiveTypes.Float64},
			{Name: "f3-str", Type: arrow.BinaryTypes.String},
		},
		nil,
	)

	f, err := os.Create("input.arrow")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	writer, err := storage.NewWriter(f, schema)
	if err != nil {
		log.Fatal(err)
	}
	defer writer.Close()

	b := array.NewRecordBuilder(pool, schema)
	defer b.Release()

	b.Field(0).(*array.Int64Builder).AppendValues([]int64{1, 2, 3, 4, 5}, nil)
	b.Field(1).(*array.Float64Builder).AppendValues([]float64{1.1, 2.2, 3.3, 4.4, 5.5}, nil)
	b.Field(2).(*array.StringBuilder).AppendValues([]string{"a", "b", "c", "d", "e"}, nil)

	rec := b.NewRecord()
	defer rec.Release()

	if err := writer.Write(rec); err != nil {
		log.Fatal(err)
	}
}
