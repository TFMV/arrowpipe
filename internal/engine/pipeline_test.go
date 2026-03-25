package engine

import (
	"context"
	"testing"

	"github.com/apache/arrow-go/v18/arrow"
	"github.com/apache/arrow-go/v18/arrow/array"
	"github.com/apache/arrow-go/v18/arrow/memory"
	"github.com/stretchr/testify/assert"
)

func TestPipeline(t *testing.T) {
	pool := memory.NewGoAllocator()

	// Create a sample record
	schema := arrow.NewSchema(
		[]arrow.Field{{Name: "f1", Type: arrow.PrimitiveTypes.Int64}}, nil)
	b := array.NewInt64Builder(pool)
	defer b.Release()
	b.AppendValues([]int64{1, 2, 3}, nil)
	arr := b.NewArray()
	defer arr.Release()
	rec := array.NewRecord(schema, []arrow.Array{arr}, -1)

	// 1. Create the pipeline
	p := NewPipeline()

	// 2. Define a simple test stage (e.g., a pass-through stage)
	passThroughStage := func(ctx context.Context, in <-chan arrow.Record) <-chan arrow.Record {
		out := make(chan arrow.Record)
		go func() {
			defer close(out)
			for rec := range in {
				out <- rec
			}
		}()
		return out
	}

	// 3. Add the stage to the pipeline
	p.AddStage(passThroughStage)

	// 4. Create a source channel and send the test record
	source := make(chan arrow.Record, 1)
	source <- rec
	close(source)

	// 5. Execute the pipeline
	output := p.Execute(context.Background(), source)

	// 6. Verify the output
	resultRec, ok := <-output
	assert.True(t, ok, "should receive a record from the output channel")
	assert.True(t, array.RecordEqual(rec, resultRec), "input and output records should be the same")

	_, ok = <-output
	assert.False(t, ok, "output channel should be closed after one record")

	resultRec.Release()
}
