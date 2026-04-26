package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/firasmosbehi/portman/internal/scanner"
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

		s := scanner.NewScanner()
		p, err := s.FindProcessByPort(port)
		if err != nil {
			return err
		}

		if !killForceFlag {
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "? Kill process %s (pid %d) on port %d? (y/N) ", p.Process, p.PID, port)
			reader := bufio.NewReader(cmd.InOrStdin())
			input, _ := reader.ReadString('\n')
			input = strings.TrimSpace(strings.ToLower(input))
			if input != "y" && input != "yes" {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Aborted.")
				return nil
			}
		}

		process, err := os.FindProcess(p.PID)
		if err != nil {
			return fmt.Errorf("failed to find process %d: %w", p.PID, err)
		}
		if err := process.Kill(); err != nil {
			return fmt.Errorf("failed to kill process %d: %w", p.PID, err)
		}

		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Killed %s (pid %d) on port %d\n", p.Process, p.PID, port)
		return nil
	},
}

func init() {
	killCmd.Flags().BoolVarP(&killForceFlag, "force", "f", false, "Skip confirmation prompt")
	rootCmd.AddCommand(killCmd)
}
