package cli

import (
	"encoding/csv"
	"io"
	"os"

	"github.com/TFMV/arrowpipe/internal/storage"
	"github.com/apache/arrow-go/v18/arrow/array"
	"github.com/spf13/cobra"
)

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export data from an ArrowPipe dataset",
}

var exportCsvCmd = &cobra.Command{
	Use:   "csv [input]",
	Short: "Export an ArrowPipe dataset to a CSV file",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		inputPath := args[0]
		outPath, _ := cmd.Flags().GetString("out")

		inFile, err := os.Open(inputPath)
		if err != nil {
			return err
		}
		defer inFile.Close()

		reader, err := storage.NewColumnarDatasetReader(inFile)
		if err != nil {
			return err
		}
		defer reader.Close()

		outFile, err := os.Create(outPath)
		if err != nil {
			return err
		}
		defer outFile.Close()

		writer := csv.NewWriter(outFile)
		defer writer.Flush()

		headers := []string{}
		for _, field := range reader.Schema().Fields() {
			headers = append(headers, field.Name)
		}
		if err := writer.Write(headers); err != nil {
			return err
		}

		for {
			rec, err := reader.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				return err
			}
			defer rec.Release()

			for i := 0; i < int(rec.NumRows()); i++ {
				row := make([]string, rec.NumCols())
				for j := 0; j < int(rec.NumCols()); j++ {
					row[j] = rec.Column(j).(*array.String).Value(i)
				}
				if err := writer.Write(row); err != nil {
					return err
				}
			}
		}

		return nil
	},
}

func init() {
	exportCmd.AddCommand(exportCsvCmd)
	rootCmd.AddCommand(exportCmd)

	exportCsvCmd.Flags().String("out", "", "Output file path")
	exportCsvCmd.MarkFlagRequired("out")
}
