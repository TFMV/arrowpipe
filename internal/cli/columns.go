package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/TFMV/arrowpipe/internal/storage"
	"github.com/spf13/cobra"
)

var columnsCmd = &cobra.Command{
	Use:   "columns [path]",
	Short: "List columns and types of an ArrowPipe dataset",
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

		schema := reader.Schema()
		for i, field := range schema.Fields() {
			fmt.Fprintf(cmd.OutOrStdout(), "%d. %s: %s\n", i+1, field.Name, field.Type)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(columnsCmd)
}
