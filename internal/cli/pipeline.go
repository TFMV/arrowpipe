package cli

import (
	"fmt"
	"strings"

	"github.com/TFMV/arrowpipe/internal/engine"
	"github.com/TFMV/arrowpipe/internal/storage"
	"github.com/spf13/cobra"
)

var pipelineCmd = &cobra.Command{
	Use:   "pipeline",
	Short: "Chain transformations in a single command",
	RunE: func(cmd *cobra.Command, args []string) error {
		filterExpr, _ := cmd.Flags().GetString("filter")
		selectCols, _ := cmd.Flags().GetString("select")
		aggMetrics, _ := cmd.Flags().GetString("aggregate")

		reader, err := storage.NewColumnarDatasetReaderFromReader(cmd.InOrStdin())
		if err != nil {
			return fmt.Errorf("error creating reader: %w", err)
		}
		defer reader.Close()

		rec, err := reader.Read()
		if err != nil {
			return fmt.Errorf("error reading record: %w", err)
		}
		defer rec.Release()

		result := rec

		if filterExpr != "" {
			predicate, err := engine.NewPredicate(filterExpr)
			if err != nil {
				return fmt.Errorf("error parsing filter expression: %w", err)
			}

			filter, err := predicate.Eval(result)
			if err != nil {
				return fmt.Errorf("error evaluating filter: %w", err)
			}
			defer filter.Release()

			result, err = engine.Filter(result, filter)
			if err != nil {
				return fmt.Errorf("error filtering: %w", err)
			}
			defer result.Release()

			if result.NumRows() == 0 {
				return nil
			}
		}

		if selectCols != "" {
			result = engine.Select(result, splitCommaList(selectCols)...)
			defer result.Release()
		}

		if aggMetrics != "" {
			metrics, err := parseMetrics(aggMetrics)
			if err != nil {
				return fmt.Errorf("error parsing metrics: %w", err)
			}

			groupBy, _ := cmd.Flags().GetString("group-by")
			result, err = computeAggregation(result, metrics, groupBy)
			if err != nil {
				return fmt.Errorf("error computing aggregation: %w", err)
			}
			defer result.Release()
		}

		writer, err := storage.NewWriter(cmd.OutOrStdout(), result.Schema())
		if err != nil {
			return fmt.Errorf("error creating writer: %w", err)
		}

		if err := writer.Write(result); err != nil {
			return fmt.Errorf("error writing output: %w", err)
		}

		return writer.Close()
	},
}

func splitCommaList(s string) []string {
	if s == "" {
		return nil
	}
	var result []string
	for _, part := range strings.Split(s, ",") {
		result = append(result, strings.TrimSpace(part))
	}
	return result
}

func init() {
	rootCmd.AddCommand(pipelineCmd)
	pipelineCmd.Flags().String("filter", "", "Filter expression (e.g., 'region == \"north\"')")
	pipelineCmd.Flags().String("select", "", "Comma-separated list of columns to select")
	pipelineCmd.Flags().String("aggregate", "", "Aggregation expression (e.g., 'sum(sales),avg(price)')")
	pipelineCmd.Flags().String("group-by", "", "Column to group by")
}
