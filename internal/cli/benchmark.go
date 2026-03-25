package cli

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/TFMV/arrowpipe/internal/storage"
	"github.com/apache/arrow-go/v18/arrow"
	"github.com/apache/arrow-go/v18/arrow/array"
	"github.com/apache/arrow-go/v18/arrow/memory"
	"github.com/spf13/cobra"
)

var benchmarkCmd = &cobra.Command{
	Use:   "benchmark",
	Short: "Benchmark read and write performance",
}

var benchmarkReadCmd = &cobra.Command{
	Use:   "read [file]",
	Short: "Benchmark read performance",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		start := time.Now()

		file, err := os.Open(args[0])
		if err != nil {
			return fmt.Errorf("error opening file: %w", err)
		}
		defer file.Close()

		reader, err := storage.NewColumnarDatasetReader(file)
		if err != nil {
			return fmt.Errorf("error creating reader: %w", err)
		}

		for {
			rec, err := reader.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				return fmt.Errorf("error reading record: %w", err)
			}
			rec.Release()
		}

		duration := time.Since(start)
		fmt.Printf("Read benchmark completed in %s\n", duration)

		return nil
	},
}

var benchmarkWriteCmd = &cobra.Command{
	Use:   "write [file]",
	Short: "Benchmark write performance",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		start := time.Now()

		file, err := os.Create(args[0])
		if err != nil {
			return fmt.Errorf("error creating file: %w", err)
		}
		defer file.Close()

		mem := memory.NewGoAllocator()
		schema := arrow.NewSchema(
			[]arrow.Field{
				{Name: "f1-i64", Type: arrow.PrimitiveTypes.Int64},
			},
			nil,
		)

		writer, err := storage.NewColumnarDataset(file, schema)
		if err != nil {
			return fmt.Errorf("error creating writer: %w", err)
		}
		defer writer.Close()

		bldr := array.NewInt64Builder(mem)
		defer bldr.Release()

		for i := 0; i < 1_000_000; i++ {
			bldr.Append(int64(i))
		}

		arr := bldr.NewArray()
		defer arr.Release()

		rec := array.NewRecord(schema, []arrow.Array{arr}, int64(arr.Len()))
		defer rec.Release()

		if err := writer.Write(rec); err != nil {
			return fmt.Errorf("error writing record: %w", err)
		}

		duration := time.Since(start)
		fmt.Printf("Write benchmark completed in %s\n", duration)

		return nil
	},
}

func init() {
	benchmarkCmd.AddCommand(benchmarkReadCmd)
	benchmarkCmd.AddCommand(benchmarkWriteCmd)
	rootCmd.AddCommand(benchmarkCmd)
}
