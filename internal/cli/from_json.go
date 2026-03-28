package cli

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/TFMV/arrowpipe/internal/storage"
	"github.com/apache/arrow-go/v18/arrow"
	"github.com/apache/arrow-go/v18/arrow/array"
	"github.com/apache/arrow-go/v18/arrow/memory"
	"github.com/spf13/cobra"
)

func newFromJSONCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "from-json",
		Short: "Converts JSON data from stdin to Arrow format on stdout.",
		Long:  `This command reads a JSON array from stdin, infers the schema, and writes the data in Arrow IPC format to stdout.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runFromJSON(cmd.InOrStdin(), cmd.OutOrStdout())
		},
	}

	return cmd
}

func runFromJSON(in io.Reader, out io.Writer) error {
	dec := json.NewDecoder(in)

	// Read the opening bracket of the JSON array
	_, err := dec.Token()
	if err != nil {
		return fmt.Errorf("expected JSON array: %w", err)
	}

	// First pass: read all records and collect field order from first record
	var records []map[string]interface{}
	var fieldOrder []string
	fieldTypes := make(map[string]arrow.DataType)

	for dec.More() {
		var r map[string]interface{}
		if err := dec.Decode(&r); err != nil {
			return fmt.Errorf("error decoding JSON record: %w", err)
		}
		records = append(records, r)

		// Collect field order from first record only
		if len(records) == 1 {
			for k, v := range r {
				fieldOrder = append(fieldOrder, k)
				switch v.(type) {
				case float64:
					fieldTypes[k] = arrow.PrimitiveTypes.Float64
				case string:
					fieldTypes[k] = arrow.BinaryTypes.String
				case bool:
					fieldTypes[k] = arrow.FixedWidthTypes.Boolean
				default:
					return fmt.Errorf("unsupported data type in JSON: %T for key %s", v, k)
				}
			}
		}
	}

	if len(records) == 0 {
		return nil // Nothing to do
	}

	// Build fields in consistent order
	fields := make([]arrow.Field, len(fieldOrder))
	for i, name := range fieldOrder {
		fields[i] = arrow.Field{Name: name, Type: fieldTypes[name]}
	}
	schema := arrow.NewSchema(fields, nil)

	mem := memory.NewGoAllocator()
	builder := array.NewRecordBuilder(mem, schema)
	defer builder.Release()

	for _, rec := range records {
		for i, f := range schema.Fields() {
			val, ok := rec[f.Name]
			if !ok || val == nil {
				builder.Field(i).AppendNull()
				continue
			}
			switch f.Type.ID() {
			case arrow.FLOAT64:
				builder.Field(i).(*array.Float64Builder).Append(val.(float64))
			case arrow.STRING:
				builder.Field(i).(*array.StringBuilder).Append(val.(string))
			case arrow.BOOL:
				builder.Field(i).(*array.BooleanBuilder).Append(val.(bool))
			}
		}
	}

	rec := builder.NewRecord()
	defer rec.Release()

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
	rootCmd.AddCommand(newFromJSONCmd())
}
