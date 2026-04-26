package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var nextRangeFlag string

var nextCmd = &cobra.Command{
	Use:   "next",
	Short: "Suggest the next available port in a range",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Fprintf(cmd.OutOrStdout(), "Finding next available port in range %s (not yet implemented)\n", nextRangeFlag)
		return nil
	},
}

func init() {
	nextCmd.Flags().StringVarP(&nextRangeFlag, "range", "r", "3000-3100", "Port range to scan (e.g., 3000-3010)")
	rootCmd.AddCommand(nextCmd)
}
