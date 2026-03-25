package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var pipelineCmd = &cobra.Command{
	Use:   "pipeline [input]",
	Short: "Chain transformations (pipe style)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("not implemented")
	},
}

func init() {
	rootCmd.AddCommand(pipelineCmd)
	pipelineCmd.Flags().String("filter", "", "Filter expression")
	pipelineCmd.Flags().String("select", "", "Comma-separated list of columns to select")
	pipelineCmd.Flags().String("aggregate", "", "Aggregation expression")
}
