package cmd

import (
	"fmt"
	"strconv"
	"time"

	"github.com/spf13/cobra"
)

var watchIntervalFlag time.Duration

var watchCmd = &cobra.Command{
	Use:   "watch <port>",
	Short: "Monitor a port and notify when it becomes available",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		port, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid port: %s", args[0])
		}
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Watching port %d every %v (not yet implemented)\n", port, watchIntervalFlag)
		return nil
	},
}

func init() {
	watchCmd.Flags().DurationVarP(&watchIntervalFlag, "interval", "i", time.Second, "Polling interval")
	rootCmd.AddCommand(watchCmd)
}
