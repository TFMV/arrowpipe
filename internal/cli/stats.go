package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/TFMV/arrowpipe/internal/storage"
	"github.com/spf13/cobra"
)

var statsCmd = &cobra.Command{
	Use:   "stats [path]",
	Short: "Calculate and display file statistics",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
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

		reader, err := storage.NewColumnarDatasetReader(file)
		if err != nil {
			return fmt.Errorf("error creating reader: %w", err)
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Schema:\n%s\n", reader.Schema())

		numChunks := 0
		var totalRows int64
		for {
			rec, err := reader.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				return fmt.Errorf("error reading record: %w", err)
			}
			totalRows += rec.NumRows()
			numChunks++
			rec.Release()
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Number of chunks: %d\n", numChunks)
		fmt.Fprintf(cmd.OutOrStdout(), "Total number of rows: %d\n", totalRows)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(statsCmd)
}
