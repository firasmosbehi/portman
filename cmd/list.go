package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sort"

	"github.com/firasmosbehi/portman/internal/platform"
	"github.com/firasmosbehi/portman/internal/reporter"
	"github.com/firasmosbehi/portman/internal/scanner"
	"github.com/firasmosbehi/portman/pkg/models"
	"github.com/spf13/cobra"
)

var listPortFlag int
var listFormatFlag string

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all listening ports with process info",
	RunE: func(cmd *cobra.Command, args []string) error {
		s := scanner.NewScanner()

		if listPortFlag != 0 {
			p, err := s.FindProcessByPort(listPortFlag)
			if err != nil {
				if errors.Is(err, platform.ErrProcessNotFound) {
					if listFormatFlag == "json" {
						_, _ = fmt.Fprintf(cmd.OutOrStdout(), "[]\n")
					} else {
						_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Port %d is not in use\n", listPortFlag)
					}
					return nil
				}
				return err
			}
			if listFormatFlag == "json" {
				return printJSON(cmd.OutOrStdout(), []models.PortProcess{*p})
			}
			rep := reporter.NewReporterWithWriter(cmd.OutOrStdout())
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
			if listFormatFlag == "json" {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "[]\n")
			} else {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "No listening ports found.")
			}
			return nil
		}

		if listFormatFlag == "json" {
			return printJSON(cmd.OutOrStdout(), ports)
		}

		rep := reporter.NewReporterWithWriter(cmd.OutOrStdout())
		rep.PrintPortTable(ports)
		return nil
	},
}

func printJSON(w io.Writer, ports []models.PortProcess) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(ports)
}

func init() {
	listCmd.Flags().IntVarP(&listPortFlag, "port", "p", 0, "Show only a specific port")
	listCmd.Flags().StringVarP(&listFormatFlag, "format", "f", "table", "Output format (table or json)")
	rootCmd.AddCommand(listCmd)
}
