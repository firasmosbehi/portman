package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/firasmosbehi/portman/internal/scanner"
	"github.com/spf13/cobra"
)

var nextRangeFlag string

var nextCmd = &cobra.Command{
	Use:   "next",
	Short: "Suggest the next available port in a range",
	RunE: func(cmd *cobra.Command, args []string) error {
		start, end, err := parseRange(nextRangeFlag)
		if err != nil {
			return err
		}

		s := scanner.NewScanner()
		port, err := s.FindNextAvailablePort(start, end)
		if err != nil {
			return err
		}

		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "%d\n", port)
		return nil
	},
}

func init() {
	nextCmd.Flags().StringVarP(&nextRangeFlag, "range", "r", "3000-3100", "Port range to scan (e.g., 3000-3010)")
	rootCmd.AddCommand(nextCmd)
}

// parseRange parses a range string like "3000-3010" into start and end integers.
func parseRange(r string) (int, int, error) {
	parts := strings.SplitN(r, "-", 2)
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid range format: %s (expected start-end)", r)
	}
	start, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil {
		return 0, 0, fmt.Errorf("invalid range start: %s", parts[0])
	}
	end, err := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil {
		return 0, 0, fmt.Errorf("invalid range end: %s", parts[1])
	}
	if start > end {
		return 0, 0, fmt.Errorf("range start (%d) must be <= end (%d)", start, end)
	}
	return start, end, nil
}
