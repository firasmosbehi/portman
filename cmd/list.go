package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var listPortFlag int

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all listening ports with process info",
	RunE: func(cmd *cobra.Command, args []string) error {
		if listPortFlag != 0 {
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Listing port %d (not yet implemented)\n", listPortFlag)
			return nil
		}
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Listing all listening ports (not yet implemented)")
		return nil
	},
}

func init() {
	listCmd.Flags().IntVarP(&listPortFlag, "port", "p", 0, "Show only a specific port")
	rootCmd.AddCommand(listCmd)
}
