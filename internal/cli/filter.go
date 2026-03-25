package cli

import (
	"fmt"
	"os"

	"github.com/TFMV/arrowpipe/internal/engine"
	"github.com/TFMV/arrowpipe/internal/storage"
	"github.com/apache/arrow-go/v18/arrow"
	"github.com/apache/arrow-go/v18/arrow/array"
	"github.com/spf13/cobra"
)

var filterCmd = &cobra.Command{
	Use:   "filter [input]",
	Short: "Filter rows by condition",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		inputPath := args[0]
		expr, _ := cmd.Flags().GetString("expr")

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

		rec, err := reader.Read()
		if err != nil {
			return err
		}
		defer rec.Release()

		predicate, err := engine.NewPredicate(expr)
		if err != nil {
			return err
		}

		filter, err := predicate.Eval(rec)
		if err != nil {
			return err
		}
		defer filter.Release()

		filtered, err := engine.Filter(rec, filter)
		if err != nil {
			return err
		}
		defer filtered.Release()

		writer, err := array.NewRecordReader(filtered.Schema(), []arrow.Record{filtered})
		if err != nil {
			return err
		}
		defer writer.Release()

		fmt.Println(writer)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(filterCmd)
	filterCmd.Flags().String("expr", "", "Filter expression (e.g., 'age > 30')")
	filterCmd.MarkFlagRequired("expr")
}
