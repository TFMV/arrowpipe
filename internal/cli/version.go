package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

const version = "0.0.1"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of arrowpipe",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Fprintf(cmd.OutOrStdout(), "arrowpipe version %s", version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
