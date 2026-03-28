package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/TFMV/arrowpipe/internal/storage"
	"github.com/apache/arrow-go/v18/arrow"
	"github.com/apache/arrow-go/v18/arrow/array"
	"github.com/apache/arrow-go/v18/arrow/memory"
	"github.com/spf13/cobra"
)

var columnsCmd = &cobra.Command{
	Use:   "columns",
	Short: "List, rename, cast, or drop columns",
	RunE: func(cmd *cobra.Command, args []string) error {
		list, _ := cmd.Flags().GetBool("list")
		ops, _ := cmd.Flags().GetString("ops")

		var inReader io.Reader = cmd.InOrStdin()
		var err error

		hasArgs := len(args) > 0
		if hasArgs {
			path := args[0]
			info, err := os.Stat(path)
			if err != nil {
				return fmt.Errorf("failed to stat path: %w", err)
			}

			var file *os.File
			if info.IsDir() {
				dataFilePath := filepath.Join(path, "data.arrow")
				file, err = os.Open(dataFilePath)
				if err != nil {
					return fmt.Errorf("failed to open data.arrow file: %w", err)
				}
			} else {
				file, err = os.Open(path)
				if err != nil {
					return fmt.Errorf("failed to open file: %w", err)
				}
			}
			defer file.Close()
			inReader = file
		}

		reader, err := storage.NewColumnarDatasetReaderFromReader(inReader)
		if err != nil {
			return fmt.Errorf("error creating reader: %w", err)
		}
		defer reader.Close()

		schema := reader.Schema()

		if list || (ops == "" && hasArgs) {
			for i, field := range schema.Fields() {
				fmt.Fprintf(cmd.OutOrStdout(), "%d. %s: %s\n", i+1, field.Name, field.Type)
			}
			return nil
		}

		if ops == "" && !hasArgs {
			return fmt.Errorf("no operations specified. Use --list to list columns or --ops to specify operations")
		}

		rec, err := reader.Read()
		if err != nil {
			return fmt.Errorf("error reading record: %w", err)
		}
		defer rec.Release()

		result, err := transformColumns(rec, ops)
		if err != nil {
			return fmt.Errorf("error transforming columns: %w", err)
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

type columnOp struct {
	operation string
	target    string
	value     string
}

func transformColumns(rec arrow.Record, opsStr string) (arrow.Record, error) {
	ops, err := parseColumnOps(opsStr)
	if err != nil {
		return nil, err
	}

	mem := memory.NewGoAllocator()
	newFields := make([]arrow.Field, 0)
	newCols := make([]arrow.Array, 0)

	dropCols := make(map[string]bool)
	renameMap := make(map[string]string)
	castMap := make(map[string]arrow.DataType)

	for _, op := range ops {
		switch op.operation {
		case "drop":
			dropCols[op.target] = true
		case "rename":
			renameMap[op.target] = op.value
		case "cast":
			dt, err := parseDataType(op.value)
			if err != nil {
				return nil, fmt.Errorf("invalid data type: %s", op.value)
			}
			castMap[op.target] = dt
		}
	}

	for i, field := range rec.Schema().Fields() {
		if dropCols[field.Name] {
			continue
		}

		newName := field.Name
		if rn, ok := renameMap[field.Name]; ok {
			newName = rn
		}

		newType := field.Type
		if ct, ok := castMap[field.Name]; ok {
			newType = ct
		}

		newFields = append(newFields, arrow.Field{Name: newName, Type: newType})

		col := rec.Column(i)
		if newType != field.Type {
			bldr := array.NewBuilder(mem, newType)
			defer bldr.Release()

			switch oldCol := col.(type) {
			case *array.String:
				switch newType {
				case arrow.PrimitiveTypes.Int64:
					for j := 0; j < oldCol.Len(); j++ {
						var val int64
						fmt.Sscanf(oldCol.Value(j), "%d", &val)
						bldr.(*array.Int64Builder).Append(val)
					}
				case arrow.PrimitiveTypes.Float64:
					for j := 0; j < oldCol.Len(); j++ {
						var val float64
						fmt.Sscanf(oldCol.Value(j), "%f", &val)
						bldr.(*array.Float64Builder).Append(val)
					}
				}
			case *array.Int64:
				switch newType {
				case arrow.PrimitiveTypes.Float64:
					for j := 0; j < oldCol.Len(); j++ {
						bldr.(*array.Float64Builder).Append(float64(oldCol.Value(j)))
					}
				case arrow.BinaryTypes.String:
					for j := 0; j < oldCol.Len(); j++ {
						bldr.(*array.StringBuilder).Append(fmt.Sprintf("%d", oldCol.Value(j)))
					}
				}
			case *array.Float64:
				switch newType {
				case arrow.BinaryTypes.String:
					for j := 0; j < oldCol.Len(); j++ {
						bldr.(*array.StringBuilder).Append(fmt.Sprintf("%f", oldCol.Value(j)))
					}
				case arrow.PrimitiveTypes.Int64:
					for j := 0; j < oldCol.Len(); j++ {
						bldr.(*array.Int64Builder).Append(int64(oldCol.Value(j)))
					}
				}
			}
			newCols = append(newCols, bldr.NewArray())
			defer newCols[len(newCols)-1].Release()
		} else {
			newCols = append(newCols, col)
		}
	}

	newSchema := arrow.NewSchema(newFields, nil)
	return array.NewRecord(newSchema, newCols, rec.NumRows()), nil
}

func parseColumnOps(opsStr string) ([]columnOp, error) {
	var ops []columnOp
	parts := strings.Split(opsStr, "),")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if !strings.HasSuffix(part, ")") {
			part = part + ")"
		}
		if !strings.HasPrefix(part, "(") {
			part = "(" + part
		}
		part = strings.Trim(part, "()")
		subparts := strings.SplitN(part, "(", 2)
		if len(subparts) != 2 {
			return nil, fmt.Errorf("invalid operation: %s", part)
		}
		op := columnOp{
			operation: strings.TrimSpace(subparts[0]),
		}
		args := subparts[1]
		argParts := strings.Split(args, ",")
		if len(argParts) < 1 {
			return nil, fmt.Errorf("invalid operation arguments: %s", part)
		}
		op.target = strings.TrimSpace(argParts[0])
		if len(argParts) > 1 {
			op.value = strings.TrimSpace(argParts[1])
		}
		ops = append(ops, op)
	}
	return ops, nil
}

func parseDataType(typeStr string) (arrow.DataType, error) {
	switch strings.TrimSpace(typeStr) {
	case "int64":
		return arrow.PrimitiveTypes.Int64, nil
	case "float64":
		return arrow.PrimitiveTypes.Float64, nil
	case "string":
		return arrow.BinaryTypes.String, nil
	case "bool":
		return arrow.FixedWidthTypes.Boolean, nil
	default:
		return nil, fmt.Errorf("unsupported data type: %s", typeStr)
	}
}

func init() {
	rootCmd.AddCommand(columnsCmd)
	columnsCmd.Flags().BoolP("list", "l", false, "List columns and their types")
	columnsCmd.Flags().String("ops", "", "Operations to perform: rename(a,b), drop(c), cast(d,int64)")
}
