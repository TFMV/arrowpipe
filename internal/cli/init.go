package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init [path]",
	Short: "Initialize a new ArrowPipe dataset",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		path := args[0]
		if err := os.MkdirAll(path, 0755); err != nil {
			return fmt.Errorf("failed to create dataset directory: %w", err)
		}
		fmt.Printf("Initialized ArrowPipe dataset at %s\n", path)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
