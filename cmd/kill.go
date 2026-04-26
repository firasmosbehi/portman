package cmd

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
)

var killForceFlag bool

var killCmd = &cobra.Command{
	Use:   "kill <port>",
	Short: "Find and kill the process using a port",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		port, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid port: %s", args[0])
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Killing process on port %d (force=%v) (not yet implemented)\n", port, killForceFlag)
		return nil
	},
}

func init() {
	killCmd.Flags().BoolVarP(&killForceFlag, "force", "f", false, "Skip confirmation prompt")
	rootCmd.AddCommand(killCmd)
}
