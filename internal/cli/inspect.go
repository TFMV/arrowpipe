package cli

import (
	"fmt"
	"io"
	"os"

	"github.com/apache/arrow-go/v18/arrow/ipc"
	"github.com/spf13/cobra"
)

var inspectCmd = &cobra.Command{
	Use:   "inspect [file]",
	Short: "Inspect individual chunks in an ArrowPipe file",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		file, err := os.Open(args[0])
		if err != nil {
			return fmt.Errorf("error opening file: %w", err)
		}
		defer file.Close()

		ipcReader, err := ipc.NewReader(file)
		if err != nil {
			return fmt.Errorf("error creating ipc reader: %w", err)
		}
		defer ipcReader.Release()

		chunkIndex, _ := cmd.Flags().GetInt("chunk")

		for i := 0; ; i++ {
			rec, err := ipcReader.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				return fmt.Errorf("error reading record: %w", err)
			}

			if chunkIndex == -1 || chunkIndex == i {
				fmt.Printf("Chunk %d:\n", i)
				fmt.Printf("  Rows: %d\n", rec.NumRows())
				fmt.Printf("  Schema:\n%s\n", rec.Schema())
			}

			if chunkIndex != -1 && chunkIndex == i {
				rec.Release()
				break
			}
			rec.Release()
		}

		return nil
	},
}

func init() {
	inspectCmd.Flags().IntP("chunk", "c", -1, "The index of the chunk to inspect")
	rootCmd.AddCommand(inspectCmd)
}
