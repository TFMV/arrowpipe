package engine

import (
	"context"
	"fmt"

	"github.com/apache/arrow-go/v18/arrow"
	"github.com/apache/arrow-go/v18/arrow/array"
	"github.com/apache/arrow-go/v18/arrow/memory"
)

// StreamingPipeline represents a streaming pipeline.
type StreamingPipeline struct {
	pipeline *Pipeline
	in       <-chan arrow.Record
}

// Aggregation is an enumeration of aggregation functions.
type Aggregation string

const (
	// Sum is the sum aggregation.
	Sum Aggregation = "SUM"
	// Mean is the mean aggregation.
	Mean Aggregation = "MEAN"
	// Count is the count aggregation.
	Count Aggregation = "COUNT"
)

// NewStreamingPipeline creates a new streaming pipeline.
func NewStreamingPipeline(in <-chan arrow.Record) *StreamingPipeline {
	return &StreamingPipeline{
		pipeline: NewPipeline(),
		in:       in,
	}
}

// Filter adds a filter stage to the pipeline.
func (sp *StreamingPipeline) Filter(expression string) *StreamingPipeline {
	sp.pipeline.AddStage(func(ctx context.Context, in <-chan arrow.Record) <-chan arrow.Record {
		out := make(chan arrow.Record)
		go func() {
			defer close(out)
			for rec := range in {
				select {
				case <-ctx.Done():
					return
				default:
					predicate, err := NewPredicate(expression)
					if err != nil {
						// Handle error appropriately.
						continue
					}

					filterArr, err := predicate.Eval(rec)
					if err != nil {
						// Handle error appropriately.
						continue
					}
					defer filterArr.Release()

					boolFilter := filterArr.(*array.Boolean)
					mem := memory.NewGoAllocator()

					numRows := 0
					for i := 0; i < boolFilter.Len(); i++ {
						if !boolFilter.IsNull(i) && boolFilter.Value(i) {
							numRows++
						}
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
						}
						cols[i] = bldr.NewArray()
						defer cols[i].Release()
					}

					newSchema := arrow.NewSchema(fields, nil)
					newRec := array.NewRecord(newSchema, cols, int64(numRows))
					out <- newRec
				}
			}
		}()
		return out
	})
	return sp
}

// Project adds a projection stage to the pipeline.
func (sp *StreamingPipeline) Project(columns []string) *StreamingPipeline {
	sp.pipeline.AddStage(func(ctx context.Context, in <-chan arrow.Record) <-chan arrow.Record {
		out := make(chan arrow.Record)
		go func() {
			defer close(out)
			for rec := range in {
				select {
				case <-ctx.Done():
					return
				default:
					indices := make([]int, 0, len(columns))
					for _, name := range columns {
						idx := rec.Schema().FieldIndices(name)
						if len(idx) > 0 {
							indices = append(indices, idx[0])
						}
					}

					fields := make([]arrow.Field, len(indices))
					arrays := make([]arrow.Array, len(indices))
					for i, idx := range indices {
						fields[i] = rec.Schema().Field(idx)
						arrays[i] = rec.Column(idx)
					}

					newSchema := arrow.NewSchema(fields, nil)
					newRec := array.NewRecord(newSchema, arrays, rec.NumRows())
					out <- newRec
				}
			}
		}()
		return out
	})
	return sp
}

// Aggregate adds an aggregation stage to the pipeline.
func (sp *StreamingPipeline) Aggregate(agg Aggregation, column string) *StreamingPipeline {
	sp.pipeline.AddStage(func(ctx context.Context, in <-chan arrow.Record) <-chan arrow.Record {
		out := make(chan arrow.Record)
		go func() {
			defer close(out)

			var sum float64
			var count int64

			for rec := range in {
				idx := rec.Schema().FieldIndices(column)
				if len(idx) == 0 {
					// Handle error: column not found
					continue
				}

				col := rec.Column(idx[0])
				count += int64(col.Len())

				switch typedCol := col.(type) {
				case *array.Int64:
					for i := 0; i < typedCol.Len(); i++ {
						sum += float64(typedCol.Value(i))
					}
				case *array.Float64:
					for i := 0; i < typedCol.Len(); i++ {
						sum += typedCol.Value(i)
					}
				default:
					// Handle error: unsupported type
				}
			}

			mem := memory.NewGoAllocator()
			var schema *arrow.Schema
			var newRec arrow.Record

			switch agg {
			case Sum:
				fields := []arrow.Field{{Name: fmt.Sprintf("sum(%s)", column), Type: arrow.PrimitiveTypes.Float64}}
				schema = arrow.NewSchema(fields, nil)
				bldr := array.NewFloat64Builder(mem)
				defer bldr.Release()
				bldr.Append(sum)
				arr := bldr.NewArray()
				defer arr.Release()
				newRec = array.NewRecord(schema, []arrow.Array{arr}, 1)
			case Mean:
				fields := []arrow.Field{{Name: fmt.Sprintf("mean(%s)", column), Type: arrow.PrimitiveTypes.Float64}}
				schema = arrow.NewSchema(fields, nil)
				mean := sum / float64(count)
				bldr := array.NewFloat64Builder(mem)
				defer bldr.Release()
				bldr.Append(mean)
				arr := bldr.NewArray()
				defer arr.Release()
				newRec = array.NewRecord(schema, []arrow.Array{arr}, 1)
			case Count:
				fields := []arrow.Field{{Name: fmt.Sprintf("count(%s)", column), Type: arrow.PrimitiveTypes.Int64}}
				schema = arrow.NewSchema(fields, nil)
				bldr := array.NewInt64Builder(mem)
				defer bldr.Release()
				bldr.Append(count)
				arr := bldr.NewArray()
				defer arr.Release()
				newRec = array.NewRecord(schema, []arrow.Array{arr}, 1)
			}

			out <- newRec
		}()
		return out
	})
	return sp
}

// Execute runs the streaming pipeline.
func (sp *StreamingPipeline) Execute(ctx context.Context) <-chan arrow.Record {
	return sp.pipeline.Execute(ctx, sp.in)
}
