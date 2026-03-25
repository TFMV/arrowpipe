package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var aggregateCmd = &cobra.Command{
	Use:   "aggregate [input]",
	Short: "Compute aggregations",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("not implemented")
	},
}

func init() {
	rootCmd.AddCommand(aggregateCmd)
	aggregateCmd.Flags().String("group-by", "", "Column to group by")
	aggregateCmd.Flags().String("metrics", "", "Comma-separated list of metrics to compute (e.g., 'avg(salary),max(age)')")
	aggregateCmd.MarkFlagRequired("metrics")
}
