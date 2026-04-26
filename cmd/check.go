package cmd

import (
	"fmt"
	"strconv"

	"github.com/firasmosbehi/portman/internal/scanner"
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

		s := scanner.NewScanner()
		free, err := s.IsPortFree(port)
		if err != nil {
			return err
		}

		if free {
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Port %d is free\n", port)
			return nil
		}

		p, err := s.FindProcessByPort(port)
		if err != nil {
			return err
		}

		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Port %d is in use by %s (pid %d)\n", port, p.Process, p.PID)
		return fmt.Errorf("port %d is in use", port)
	},
}

func init() {
	rootCmd.AddCommand(checkCmd)
}
