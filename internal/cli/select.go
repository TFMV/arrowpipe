package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/TFMV/arrowpipe/internal/engine"
	"github.com/TFMV/arrowpipe/internal/storage"
	"github.com/apache/arrow-go/v18/arrow"
	"github.com/apache/arrow-go/v18/arrow/array"
	"github.com/spf13/cobra"
)

var selectCmd = &cobra.Command{
	Use:   "select [input]",
	Short: "Select specific columns",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		inputPath := args[0]
		columns, _ := cmd.Flags().GetString("columns")

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

		selected := engine.Select(rec, strings.Split(columns, ",")...)
		defer selected.Release()

		writer, err := array.NewRecordReader(selected.Schema(), []arrow.Record{selected})
		if err != nil {
			return err
		}
		defer writer.Release()

		fmt.Println(writer)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(selectCmd)
	selectCmd.Flags().String("columns", "", "Comma-separated list of columns to select")
	selectCmd.MarkFlagRequired("columns")
}
