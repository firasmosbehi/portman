package cmd

import (
	"errors"
	"fmt"
	"sort"

	"github.com/firasmosbehi/portman/internal/platform"
	"github.com/firasmosbehi/portman/internal/reporter"
	"github.com/firasmosbehi/portman/internal/scanner"
	"github.com/firasmosbehi/portman/pkg/models"
	"github.com/spf13/cobra"
)

var listPortFlag int

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all listening ports with process info",
	RunE: func(cmd *cobra.Command, args []string) error {
		s := scanner.NewScanner()
		rep := reporter.NewReporterWithWriter(cmd.OutOrStdout())

		if listPortFlag != 0 {
			p, err := s.FindProcessByPort(listPortFlag)
			if err != nil {
				if errors.Is(err, platform.ErrProcessNotFound) {
					_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Port %d is not in use\n", listPortFlag)
					return nil
				}
				return err
			}
			rep.PrintPortTable([]models.PortProcess{*p})
			return nil
		}

		ports, err := s.ListPorts()
		if err != nil {
			return err
		}

		sort.Slice(ports, func(i, j int) bool {
			return ports[i].Port < ports[j].Port
		})

		if len(ports) == 0 {
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "No listening ports found.")
			return nil
		}

		rep.PrintPortTable(ports)
		return nil
	},
}

func init() {
	listCmd.Flags().IntVarP(&listPortFlag, "port", "p", 0, "Show only a specific port")
	rootCmd.AddCommand(listCmd)
}
