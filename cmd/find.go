package cmd

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/firasmosbehi/portman/internal/reporter"
	"github.com/firasmosbehi/portman/internal/scanner"
	"github.com/firasmosbehi/portman/pkg/models"
	"github.com/spf13/cobra"
)

var findPIDFlag int
var findFormatFlag string

var findCmd = &cobra.Command{
	Use:   "find <process>",
	Short: "Find listening ports by process name or PID",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if findPIDFlag == 0 && len(args) == 0 {
			return fmt.Errorf("process name or --pid is required")
		}
		if findPIDFlag != 0 && len(args) > 0 {
			return fmt.Errorf("cannot specify both process name and --pid")
		}

		s := scanner.NewScanner()
		ports, err := s.ListPorts()
		if err != nil {
			return err
		}

		filtered := make([]models.PortProcess, 0)
		if findPIDFlag != 0 {
			for _, p := range ports {
				if p.PID == findPIDFlag {
					filtered = append(filtered, p)
				}
			}
		} else {
			query := strings.ToLower(args[0])
			for _, p := range ports {
				if strings.Contains(strings.ToLower(p.Process), query) {
					filtered = append(filtered, p)
				}
			}
		}

		sort.Slice(filtered, func(i, j int) bool { return filtered[i].Port < filtered[j].Port })

		if findFormatFlag == "json" {
			b, _ := json.MarshalIndent(filtered, "", "  ")
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), string(b))
			return nil
		}

		if len(filtered) == 0 {
			if findPIDFlag != 0 {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "No process found with PID %d\n", findPIDFlag)
			} else {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "No process found matching %q\n", args[0])
			}
			return nil
		}

		rep := reporter.NewReporterWithWriter(cmd.OutOrStdout())
		rep.PrintPortTable(filtered)
		return nil
	},
}

func init() {
	findCmd.Flags().IntVar(&findPIDFlag, "pid", 0, "Filter by exact PID")
	findCmd.Flags().StringVar(&findFormatFlag, "format", "table", "Output format (table|json)")
	rootCmd.AddCommand(findCmd)
}
