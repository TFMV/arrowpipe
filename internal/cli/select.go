package cli

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/TFMV/arrowpipe/internal/engine"
	"github.com/TFMV/arrowpipe/internal/storage"
	"github.com/spf13/cobra"
)

var selectCmd = &cobra.Command{
	Use:   "select",
	Short: "Select specific columns",
	RunE: func(cmd *cobra.Command, args []string) error {
		columns, _ := cmd.Flags().GetString("columns")

		var inReader io.Reader = cmd.InOrStdin()
		var err error

		if len(args) > 0 {
			inputPath := args[0]
			var inFile *os.File
			inFile, err = os.Open(inputPath)
			if err != nil {
				return fmt.Errorf("error opening input file: %w", err)
			}
			defer inFile.Close()
			inReader = inFile
		}

		reader, err := storage.NewColumnarDatasetReaderFromReader(inReader)
		if err != nil {
			return fmt.Errorf("error creating reader: %w", err)
		}
		defer reader.Close()

		rec, err := reader.Read()
		if err != nil {
			return fmt.Errorf("error reading record: %w", err)
		}
		defer rec.Release()

		selected := engine.Select(rec, strings.Split(columns, ",")...)
		defer selected.Release()

		writer, err := storage.NewWriter(cmd.OutOrStdout(), selected.Schema())
		if err != nil {
			return fmt.Errorf("error creating writer: %w", err)
		}

		if err := writer.Write(selected); err != nil {
			return fmt.Errorf("error writing output: %w", err)
		}

		return writer.Close()
	},
}

func init() {
	rootCmd.AddCommand(selectCmd)
	selectCmd.Flags().String("columns", "", "Comma-separated list of columns to select")
	selectCmd.MarkFlagRequired("columns")
}
