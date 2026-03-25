package cli

import (
	"fmt"
	"os"

	"github.com/TFMV/arrowpipe/internal/storage"
	"github.com/apache/arrow-go/v18/arrow"
	"github.com/apache/arrow-go/v18/arrow/array"
	"github.com/apache/arrow-go/v18/arrow/memory"
	"github.com/spf13/cobra"
)

var createDummyDataCmd = &cobra.Command{
	Use:   "create-dummy-data",
	Short: "Create a dummy dataset for testing",
	RunE: func(cmd *cobra.Command, args []string) error {
		out, _ := cmd.Flags().GetString("out")
		rows, _ := cmd.Flags().GetInt("rows")

		schema := arrow.NewSchema(
			[]arrow.Field{
				{Name: "col1", Type: arrow.PrimitiveTypes.Int64},
				{Name: "col2", Type: arrow.PrimitiveTypes.Float64},
				{Name: "col3", Type: arrow.BinaryTypes.String},
			},
			nil,
		)

		pool := memory.NewGoAllocator()
		builder := array.NewRecordBuilder(pool, schema)
		defer builder.Release()

		for i := 0; i < rows; i++ {
			builder.Field(0).(*array.Int64Builder).Append(int64(i))
			builder.Field(1).(*array.Float64Builder).Append(float64(i))
			builder.Field(2).(*array.StringBuilder).Append(fmt.Sprintf("dummy-%d", i))
		}

		rec := builder.NewRecord()
		defer rec.Release()

		outFile, err := os.Create(out)
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
	rootCmd.AddCommand(createDummyDataCmd)
	createDummyDataCmd.Flags().String("out", "", "Output dataset path")
	createDummyDataCmd.MarkFlagRequired("out")
	createDummyDataCmd.Flags().Int("rows", 100, "Number of rows to generate")
}
