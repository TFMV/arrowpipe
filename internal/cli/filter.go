package cli

import (
	"fmt"
	"io"
	"os"

	"github.com/TFMV/arrowpipe/internal/engine"
	"github.com/TFMV/arrowpipe/internal/storage"
	"github.com/spf13/cobra"
)

var filterCmd = &cobra.Command{
	Use:   "filter",
	Short: "Filter rows by condition",
	RunE: func(cmd *cobra.Command, args []string) error {
		expr, _ := cmd.Flags().GetString("expr")

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

		if rec.NumRows() == 0 {
			return nil
		}

		predicate, err := engine.NewPredicate(expr)
		if err != nil {
			return fmt.Errorf("error parsing expression: %w", err)
		}

		filter, err := predicate.Eval(rec)
		if err != nil {
			return fmt.Errorf("error evaluating filter: %w", err)
		}
		defer filter.Release()

		filtered, err := engine.Filter(rec, filter)
		if err != nil {
			return fmt.Errorf("error filtering: %w", err)
		}
		defer filtered.Release()

		if filtered.NumRows() == 0 {
			return nil
		}

		writer, err := storage.NewWriter(cmd.OutOrStdout(), filtered.Schema())
		if err != nil {
			return fmt.Errorf("error creating writer: %w", err)
		}

		if err := writer.Write(filtered); err != nil {
			return fmt.Errorf("error writing output: %w", err)
		}

		return writer.Close()
	},
}

func init() {
	rootCmd.AddCommand(filterCmd)
	filterCmd.Flags().String("expr", "", "Filter expression (e.g., 'age > 30' or \"name == 'test'\")")
	filterCmd.MarkFlagRequired("expr")
}
