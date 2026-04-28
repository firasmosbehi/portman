package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/firasmosbehi/portman/internal/scanner"
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

		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
		defer stop()

		s := scanner.NewScanner()

		// Check immediately first
		free, err := s.IsPortFree(port)
		if err != nil {
			return err
		}
		if free {
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Port %d is already available\n", port)
			return nil
		}

		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Watching port %d... (press Ctrl+C to cancel)\n", port)

		ticker := time.NewTicker(watchIntervalFlag)
		defer ticker.Stop()

		start := time.Now()
		first := true

		for {
			select {
			case <-ctx.Done():
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "\nCancelled.")
				return nil
			case <-ticker.C:
				free, err := s.IsPortFree(port)
				if err != nil {
					return err
				}
				if free {
					_, _ = fmt.Fprintf(cmd.OutOrStdout(), "\r✓ Port %d is now available\n", port)
					return nil
				}

				elapsed := time.Since(start).Round(time.Second)
				if first {
					_, _ = fmt.Fprintf(cmd.OutOrStdout(), "⏳ Port %d is still in use (waiting %s)...", port, elapsed)
					first = false
				} else {
					_, _ = fmt.Fprintf(cmd.OutOrStdout(), "\r⏳ Port %d is still in use (waiting %s)...", port, elapsed)
				}
			}
		}
	},
}

func init() {
	watchCmd.Flags().DurationVarP(&watchIntervalFlag, "interval", "i", 1*time.Second, "Polling interval")
	rootCmd.AddCommand(watchCmd)
}
