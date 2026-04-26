package cmd

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
)

var checkCmd = &cobra.Command{
	Use:   "check <port>",
	Short: "Report if a port is free or in use",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		port, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid port: %s", args[0])
		}
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Checking port %d (not yet implemented)\n", port)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(checkCmd)
}
