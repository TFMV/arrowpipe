package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"github.com/apache/arrow-go/v18/arrow/ipc"
	"github.com/spf13/cobra"
)

func newToJSONCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "to-json",
		Short: "Converts Arrow data from stdin to JSON on stdout.",
		Long:  `This command reads Arrow IPC data from stdin and converts it to a JSON representation on stdout.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runToJSON(cmd.InOrStdin(), cmd.OutOrStdout())
		},
	}

	return cmd
}

func runToJSON(in io.Reader, out io.Writer) error {
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

	encoder := json.NewEncoder(out)

	for i := 0; i < r.NumRecords(); i++ {
		rec, err := r.Record(i)
		if err != nil {
			return fmt.Errorf("error reading record %d: %w", i, err)
		}
		defer rec.Release()

		// This is a simplified conversion. A more robust implementation would handle
		// different data types and nested structures.
		for row := 0; row < int(rec.NumRows()); row++ {
			rowMap := make(map[string]interface{})
			for col, field := range r.Schema().Fields() {
				rowMap[field.Name] = rec.Column(col).ValueStr(row)
			}
			if err := encoder.Encode(rowMap); err != nil {
				return fmt.Errorf("error encoding row to JSON: %w", err)
			}
		}
	}

	return nil
}

func init() {
	rootCmd.AddCommand(newToJSONCmd())
}
