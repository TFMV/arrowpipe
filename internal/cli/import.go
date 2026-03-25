package cli

import (
	"encoding/csv"
	"io"
	"os"

	"github.com/TFMV/arrowpipe/internal/storage"
	"github.com/apache/arrow-go/v18/arrow"
	"github.com/apache/arrow-go/v18/arrow/array"
	"github.com/apache/arrow-go/v18/arrow/memory"
	"github.com/spf13/cobra"
)

var importCmd = &cobra.Command{
	Use:   "import",
	Short: "Import data into an ArrowPipe dataset",
}

var importCsvCmd = &cobra.Command{
	Use:   "csv [input]",
	Short: "Import a CSV file into an ArrowPipe dataset",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		inputPath := args[0]
		outPath, _ := cmd.Flags().GetString("out")

		inFile, err := os.Open(inputPath)
		if err != nil {
			return err
		}
		defer inFile.Close()

		reader := csv.NewReader(inFile)
		headers, err := reader.Read()
		if err != nil {
			return err
		}

		fields := make([]arrow.Field, len(headers))
		for i, header := range headers {
			fields[i] = arrow.Field{Name: header, Type: arrow.BinaryTypes.String}
		}
		schema := arrow.NewSchema(fields, nil)

		pool := memory.NewGoAllocator()
		builder := array.NewRecordBuilder(pool, schema)
		defer builder.Release()

		for {
			row, err := reader.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				return err
			}

			for i, v := range row {
				builder.Field(i).(*array.StringBuilder).AppendString(v)
			}
		}

		rec := builder.NewRecord()
		defer rec.Release()

		outFile, err := os.Create(outPath)
		if err != nil {
			return err
		}
		defer outFile.Close()

		ds, err := storage.NewColumnarDataset(outFile, schema)
		if err != nil {
			return err
		}

		if err := ds.Write(rec); err != nil {
			return err
		}
		return ds.Close()
	},
}

func init() {
	importCmd.AddCommand(importCsvCmd)
	rootCmd.AddCommand(importCmd)

	importCsvCmd.Flags().String("out", "", "Output dataset path")
	importCsvCmd.MarkFlagRequired("out")
}
