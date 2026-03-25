package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "arrowpipe",
	Short: "ArrowPipe is a tool for working with Arrow data.",
	Run: func(cmd *cobra.Command, args []string) {
		// Default action when no subcommand is provided
		cmd.Help()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolP("version", "v", false, "Print the version and exit")
}
