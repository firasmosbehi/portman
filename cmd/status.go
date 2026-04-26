package cmd

import (
	"fmt"
	"os"

	"github.com/firasmosbehi/portman/internal/health"
	"github.com/firasmosbehi/portman/internal/registry"
	"github.com/firasmosbehi/portman/internal/reporter"
	"github.com/firasmosbehi/portman/internal/scanner"
	"github.com/firasmosbehi/portman/pkg/models"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check project services against portman.yml registry",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Look for portman.yml in current directory.
		const fileName = "portman.yml"
		if _, err := os.Stat(fileName); os.IsNotExist(err) {
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "No portman.yml found in current directory.")
			return nil
		}

		reg, err := registry.Load(fileName)
		if err != nil {
			return err
		}

		s := scanner.NewScanner()
		h := health.NewChecker()
		var statuses []models.ServiceStatus

		for _, svc := range reg.Services {
			st := models.ServiceStatus{
				Name:     svc.Name,
				Expected: svc.Port,
			}

			free, err := s.IsPortFree(svc.Port)
			if err != nil {
				return err
			}

			if free {
				st.Actual = "-"
				st.Status = "not running"
				st.Healthy = false
			} else {
				st.Actual = fmt.Sprintf("%d", svc.Port)
				if svc.HealthCheck != "" {
					if h.CommandCheck(svc.HealthCheck) {
						st.Status = "healthy"
						st.Healthy = true
					} else {
						st.Status = "unhealthy"
						st.Healthy = false
					}
				} else {
					st.Status = "running"
					st.Healthy = true
				}
			}

			statuses = append(statuses, st)
		}

		r := reporter.NewReporterWithWriter(cmd.OutOrStdout())
		r.PrintServiceStatusTable(statuses)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
