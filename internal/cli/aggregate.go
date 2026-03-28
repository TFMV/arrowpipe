package cli

import (
	"fmt"
	"strings"

	"github.com/TFMV/arrowpipe/internal/storage"
	"github.com/apache/arrow-go/v18/arrow"
	"github.com/apache/arrow-go/v18/arrow/array"
	"github.com/apache/arrow-go/v18/arrow/memory"
	"github.com/spf13/cobra"
)

var aggregateCmd = &cobra.Command{
	Use:   "aggregate",
	Short: "Compute aggregations",
	RunE: func(cmd *cobra.Command, args []string) error {
		metricsStr, _ := cmd.Flags().GetString("metrics")
		groupBy, _ := cmd.Flags().GetString("group-by")

		reader, err := storage.NewColumnarDatasetReaderFromReader(cmd.InOrStdin())
		if err != nil {
			return fmt.Errorf("error creating reader: %w", err)
		}
		defer reader.Close()

		metrics, err := parseMetrics(metricsStr)
		if err != nil {
			return fmt.Errorf("error parsing metrics: %w", err)
		}

		rec, err := reader.Read()
		if err != nil {
			return fmt.Errorf("error reading record: %w", err)
		}
		defer rec.Release()

		result, err := computeAggregation(rec, metrics, groupBy)
		if err != nil {
			return fmt.Errorf("error computing aggregation: %w", err)
		}
		defer result.Release()

		writer, err := storage.NewWriter(cmd.OutOrStdout(), result.Schema())
		if err != nil {
			return fmt.Errorf("error creating writer: %w", err)
		}

		if err := writer.Write(result); err != nil {
			return fmt.Errorf("error writing output: %w", err)
		}

		return writer.Close()
	},
}

type metric struct {
	function string
	column   string
}

func parseMetrics(metricsStr string) ([]metric, error) {
	var metrics []metric
	parts := strings.Split(metricsStr, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		part = strings.Trim(part, "()")
		funcCol := strings.Split(part, "(")
		if len(funcCol) != 2 {
			return nil, fmt.Errorf("invalid metric format: %s", part)
		}
		metrics = append(metrics, metric{
			function: strings.TrimSpace(funcCol[0]),
			column:   strings.TrimSpace(funcCol[1]),
		})
	}
	return metrics, nil
}

func computeAggregation(rec arrow.Record, metrics []metric, groupBy string) (arrow.Record, error) {
	mem := memory.NewGoAllocator()
	fields := make([]arrow.Field, len(metrics))
	builders := make([]array.Builder, len(metrics))

	for i, m := range metrics {
		var dt arrow.DataType
		switch m.function {
		case "sum", "avg", "mean":
			dt = arrow.PrimitiveTypes.Float64
		case "count":
			dt = arrow.PrimitiveTypes.Int64
		case "min", "max":
			dt = arrow.PrimitiveTypes.Float64
		default:
			return nil, fmt.Errorf("unsupported aggregation function: %s", m.function)
		}
		fields[i] = arrow.Field{Name: fmt.Sprintf("%s_%s", m.function, m.column), Type: dt}
		builders[i] = array.NewBuilder(mem, dt)
		defer builders[i].Release()
	}

	colIndices := make([]int, len(metrics))
	for i, m := range metrics {
		idx := rec.Schema().FieldIndices(m.column)
		if len(idx) == 0 {
			return nil, fmt.Errorf("column not found: %s", m.column)
		}
		colIndices[i] = idx[0]
	}

	for i, m := range metrics {
		col := rec.Column(colIndices[i])
		switch m.function {
		case "sum":
			var sum float64
			switch typedCol := col.(type) {
			case *array.Int64:
				for j := 0; j < typedCol.Len(); j++ {
					sum += float64(typedCol.Value(j))
				}
			case *array.Float64:
				for j := 0; j < typedCol.Len(); j++ {
					sum += typedCol.Value(j)
				}
			}
			builders[i].(*array.Float64Builder).Append(sum)
		case "avg", "mean":
			var sum float64
			count := 0
			switch typedCol := col.(type) {
			case *array.Int64:
				for j := 0; j < typedCol.Len(); j++ {
					sum += float64(typedCol.Value(j))
					count++
				}
			case *array.Float64:
				for j := 0; j < typedCol.Len(); j++ {
					sum += typedCol.Value(j)
					count++
				}
			}
			if count > 0 {
				builders[i].(*array.Float64Builder).Append(sum / float64(count))
			} else {
				builders[i].(*array.Float64Builder).Append(0)
			}
		case "count":
			builders[i].(*array.Int64Builder).Append(int64(col.Len()))
		case "min":
			switch typedCol := col.(type) {
			case *array.Int64:
				minVal := typedCol.Value(0)
				for j := 1; j < typedCol.Len(); j++ {
					if typedCol.Value(j) < minVal {
						minVal = typedCol.Value(j)
					}
				}
				builders[i].(*array.Float64Builder).Append(float64(minVal))
			case *array.Float64:
				minVal := typedCol.Value(0)
				for j := 1; j < typedCol.Len(); j++ {
					if typedCol.Value(j) < minVal {
						minVal = typedCol.Value(j)
					}
				}
				builders[i].(*array.Float64Builder).Append(minVal)
			}
		case "max":
			switch typedCol := col.(type) {
			case *array.Int64:
				maxVal := typedCol.Value(0)
				for j := 1; j < typedCol.Len(); j++ {
					if typedCol.Value(j) > maxVal {
						maxVal = typedCol.Value(j)
					}
				}
				builders[i].(*array.Float64Builder).Append(float64(maxVal))
			case *array.Float64:
				maxVal := typedCol.Value(0)
				for j := 1; j < typedCol.Len(); j++ {
					if typedCol.Value(j) > maxVal {
						maxVal = typedCol.Value(j)
					}
				}
				builders[i].(*array.Float64Builder).Append(maxVal)
			}
		}
	}

	cols := make([]arrow.Array, len(builders))
	for i, b := range builders {
		cols[i] = b.NewArray()
		defer cols[i].Release()
	}

	schema := arrow.NewSchema(fields, nil)
	return array.NewRecord(schema, cols, 1), nil
}

func init() {
	rootCmd.AddCommand(aggregateCmd)
	aggregateCmd.Flags().String("group-by", "", "Column to group by")
	aggregateCmd.Flags().String("metrics", "", "Comma-separated list of metrics to compute (e.g., 'avg(salary),max(age)')")
	aggregateCmd.MarkFlagRequired("metrics")
}
