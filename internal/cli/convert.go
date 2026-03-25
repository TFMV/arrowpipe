package cli

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/TFMV/arrowpipe/internal/storage"
	"github.com/apache/arrow-go/v18/arrow"
	"github.com/apache/arrow-go/v18/arrow/array"
	"github.com/apache/arrow-go/v18/arrow/ipc"
	"github.com/apache/arrow-go/v18/arrow/memory"
	"github.com/spf13/cobra"
)

var convertCmd = &cobra.Command{
	Use:   "convert",
	Short: "Convert between CSV and ArrowPipe formats",
}

var toCsvCmd = &cobra.Command{
	Use:   "tocsv [input] [output]",
	Short: "Convert an ArrowPipe file to a CSV file",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		inFile, err := os.Open(args[0])
		if err != nil {
			return fmt.Errorf("error opening input file: %w", err)
		}
		defer inFile.Close()

		outFile, err := os.Create(args[1])
		if err != nil {
			return fmt.Errorf("error creating output file: %w", err)
		}
		defer outFile.Close()

		ipcReader, err := ipc.NewReader(inFile)
		if err != nil {
			return fmt.Errorf("error creating ipc reader: %w", err)
		}
		defer ipcReader.Release()

		csvWriter := csv.NewWriter(outFile)
		defer csvWriter.Flush()

		// Write header
		header := make([]string, len(ipcReader.Schema().Fields()))
		for i, field := range ipcReader.Schema().Fields() {
			header[i] = field.Name
		}
		if err := csvWriter.Write(header); err != nil {
			return fmt.Errorf("error writing header: %w", err)
		}

		for {
			rec, err := ipcReader.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				return fmt.Errorf("error reading record: %w", err)
			}

			for i := 0; i < int(rec.NumRows()); i++ {
				row := make([]string, rec.NumCols())
				for j := 0; j < int(rec.NumCols()); j++ {
					col := rec.Column(j)
					switch typedCol := col.(type) {
					case *array.Int64:
						row[j] = strconv.FormatInt(typedCol.Value(i), 10)
					case *array.Float64:
						row[j] = strconv.FormatFloat(typedCol.Value(i), 'f', -1, 64)
					case *array.String:
						row[j] = typedCol.Value(i)
					}
				}
				if err := csvWriter.Write(row); err != nil {
					return fmt.Errorf("error writing row: %w", err)
				}
			}
			rec.Release()
		}

		return nil
	},
}

var fromCsvCmd = &cobra.Command{
	Use:   "fromcsv [input] [output]",
	Short: "Convert a CSV file to an ArrowPipe file",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		inFile, err := os.Open(args[0])
		if err != nil {
			return fmt.Errorf("error opening input file: %w", err)
		}
		defer inFile.Close()

		outFile, err := os.Create(args[1])
		if err != nil {
			return fmt.Errorf("error creating output file: %w", err)
		}
		defer outFile.Close()

		csvReader := csv.NewReader(inFile)
		header, err := csvReader.Read()
		if err != nil {
			return fmt.Errorf("error reading header: %w", err)
		}

		fields := make([]arrow.Field, len(header))
		for i, name := range header {
			// For simplicity, we're assuming all columns are strings.
			// A more robust implementation would inspect the data to determine the type.
			fields[i] = arrow.Field{Name: name, Type: arrow.BinaryTypes.String}
		}
		schema := arrow.NewSchema(fields, nil)

		writer, err := storage.NewColumnarDataset(outFile, schema)
		if err != nil {
			return fmt.Errorf("error creating writer: %w", err)
		}
		defer writer.Close()

		mem := memory.NewGoAllocator()
		bld := array.NewRecordBuilder(mem, schema)
		defer bld.Release()

		for {
			row, err := csvReader.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				return fmt.Errorf("error reading row: %w", err)
			}

			for i, val := range row {
				bld.Field(i).(*array.StringBuilder).Append(val)
			}

			if bld.Field(0).Len() == 1000 { // Write in chunks of 1000
				rec := bld.NewRecord()
				if err := writer.Write(rec); err != nil {
					return fmt.Errorf("error writing record: %w", err)
				}
				rec.Release()
			}
		}

		if bld.Field(0).Len() > 0 {
			rec := bld.NewRecord()
			if err := writer.Write(rec); err != nil {
				return fmt.Errorf("error writing record: %w", err)
			}
			rec.Release()
		}

		return nil
	},
}

func init() {
	convertCmd.AddCommand(toCsvCmd)
	convertCmd.AddCommand(fromCsvCmd)
	rootCmd.AddCommand(convertCmd)
}
