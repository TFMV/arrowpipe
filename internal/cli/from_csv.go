package cli

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"

	"github.com/TFMV/arrowpipe/internal/storage"
	"github.com/apache/arrow-go/v18/arrow"
	"github.com/apache/arrow-go/v18/arrow/array"
	"github.com/apache/arrow-go/v18/arrow/memory"
	"github.com/spf13/cobra"
)

func newFromCSVCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "from-csv",
		Short: "Converts CSV data from stdin to Arrow format on stdout.",
		Long:  `This command reads CSV data from stdin, infers the schema from the header, and writes the data in Arrow IPC format to stdout.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runFromCSV(cmd.InOrStdin(), cmd.OutOrStdout())
		},
	}

	return cmd
}

func runFromCSV(in io.Reader, out io.Writer) error {
	r := csv.NewReader(in)

	// Read header
	header, err := r.Read()
	if err != nil {
		return fmt.Errorf("error reading CSV header: %w", err)
	}

	// Read all records
	records, err := r.ReadAll()
	if err != nil {
		return fmt.Errorf("error reading CSV records: %w", err)
	}

	if len(records) == 0 {
		return nil // Nothing to do
	}

	// Infer schema and build Arrow data
	fields := make([]arrow.Field, len(header))
	columns := make([]array.Builder, len(header))
	mem := memory.NewGoAllocator()

	for i, name := range header {
		// For simplicity, we'll infer all columns as strings first
		fields[i] = arrow.Field{Name: name, Type: arrow.BinaryTypes.String}
		columns[i] = array.NewStringBuilder(mem)
	}
	schema := arrow.NewSchema(fields, nil)

	// Attempt to infer more specific types from the first data row
	if len(records) > 0 {
		for i, value := range records[0] {
			if _, err := strconv.ParseInt(value, 10, 64); err == nil {
				fields[i].Type = arrow.PrimitiveTypes.Int64
				columns[i] = array.NewInt64Builder(mem)
			} else if _, err := strconv.ParseFloat(value, 64); err == nil {
				fields[i].Type = arrow.PrimitiveTypes.Float64
				columns[i] = array.NewFloat64Builder(mem)
			} else if _, err := strconv.ParseBool(value); err == nil {
				fields[i].Type = arrow.FixedWidthTypes.Boolean
				columns[i] = array.NewBooleanBuilder(mem)
			}
		}
		schema = arrow.NewSchema(fields, nil) // Recreate schema with updated types
	}

	// Populate columns
	for _, record := range records {
		for i, value := range record {
			switch b := columns[i].(type) {
			case *array.StringBuilder:
				b.Append(value)
			case *array.Int64Builder:
				if v, err := strconv.ParseInt(value, 10, 64); err == nil {
					b.Append(v)
				} else {
					b.AppendNull()
				}
			case *array.Float64Builder:
				if v, err := strconv.ParseFloat(value, 64); err == nil {
					b.Append(v)
				} else {
					b.AppendNull()
				}
			case *array.BooleanBuilder:
				if v, err := strconv.ParseBool(value); err == nil {
					b.Append(v)
				} else {
					b.AppendNull()
				}
			}
		}
	}

	// Build the record
	cols := make([]arrow.Array, len(columns))
	for i, b := range columns {
		cols[i] = b.NewArray()
		defer cols[i].Release()
		defer b.Release()
	}

	rec := array.NewRecord(schema, cols, int64(len(records)))
	defer rec.Release()

	// Write to output
	w, err := storage.NewWriter(out, schema)
	if err != nil {
		return err
	}

	if err := w.Write(rec); err != nil {
		return err
	}

	return w.Close()
}

func init() {
	rootCmd.AddCommand(newFromCSVCmd())
}
