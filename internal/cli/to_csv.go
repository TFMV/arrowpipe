package cli

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"

	"github.com/apache/arrow-go/v18/arrow/ipc"
	"github.com/spf13/cobra"
)

func newToCSVCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "to-csv",
		Short: "Converts Arrow data from stdin to CSV on stdout.",
		Long:  `This command reads Arrow IPC data from stdin and converts it to a CSV representation on stdout.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runToCSV(cmd.InOrStdin(), cmd.OutOrStdout())
		},
	}

	return cmd
}

func runToCSV(in io.Reader, out io.Writer) error {
	data, err := io.ReadAll(in)
	if err != nil {
		return fmt.Errorf("error reading input: %w", err)
	}

	if len(data) == 0 {
		return nil
	}

	r, err := ipc.NewFileReader(bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("error creating Arrow file reader: %w", err)
	}
	defer r.Close()

	w := csv.NewWriter(out)
	defer w.Flush()

	// Write header
	fields := r.Schema().Fields()
	header := make([]string, len(fields))
	for i, field := range fields {
		header[i] = field.Name
	}
	if err := w.Write(header); err != nil {
		return fmt.Errorf("error writing CSV header: %w", err)
	}

	// Write records
	for i := 0; i < r.NumRecords(); i++ {
		rec, err := r.Record(i)
		if err != nil {
			return fmt.Errorf("error reading record %d: %w", i, err)
		}
		defer rec.Release()

		for row := 0; row < int(rec.NumRows()); row++ {
			rowStr := make([]string, len(fields))
			for col := 0; col < len(fields); col++ {
				rowStr[col] = rec.Column(col).ValueStr(row)
			}
			if err := w.Write(rowStr); err != nil {
				return fmt.Errorf("error writing CSV row: %w", err)
			}
		}
	}

	return nil
}

func init() {
	rootCmd.AddCommand(newToCSVCmd())
}
