package reporter

import (
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/fatih/color"
	"github.com/firasmosbehi/portman/pkg/models"
)

// Reporter handles terminal output formatting.
type Reporter struct {
	out io.Writer
	tw  *tabwriter.Writer
}

// NewReporter creates a new reporter.
func NewReporter() *Reporter {
	return NewReporterWithWriter(os.Stdout)
}

// NewReporterWithWriter creates a new reporter that writes to w.
func NewReporterWithWriter(w io.Writer) *Reporter {
	return &Reporter{
		out: w,
		tw:  tabwriter.NewWriter(w, 0, 0, 2, ' ', 0),
	}
}

// PrintPortTable prints a table of port/process info.
func (r *Reporter) PrintPortTable(ports []models.PortProcess) {
	if os.Getenv("NO_COLOR") == "" {
		color.NoColor = false
	}

	header := color.New(color.Bold).SprintFunc()
	_, _ = fmt.Fprintln(r.tw, header("PORT\tPROTOCOL\tPROCESS\tPID\tUSER\tAGE"))
	_, _ = fmt.Fprintln(r.tw, strings.Repeat("─", 60))

	for _, p := range ports {
		_, _ = fmt.Fprintf(r.tw, "%d\t%s\t%s\t%d\t%s\t%s\n",
			p.Port, p.Protocol, p.Process, p.PID, p.User, p.Age)
	}

	_ = r.tw.Flush()
}

// PrintStatus prints a status message.
func (r *Reporter) PrintStatus(msg string) {
	_, _ = fmt.Fprintln(r.out, msg)
}

// PrintError prints an error message.
func (r *Reporter) PrintError(err error) {
	_, _ = fmt.Fprintf(r.out, "Error: %v\n", err)
}

// PrintSuccess prints a success message.
func (r *Reporter) PrintSuccess(msg string) {
	_, _ = fmt.Fprintf(r.out, "✓ %s\n", msg)
}

// PrintServiceStatusTable prints a table of service statuses.
func (r *Reporter) PrintServiceStatusTable(statuses []models.ServiceStatus) {
	if os.Getenv("NO_COLOR") == "" {
		color.NoColor = false
	}

	header := color.New(color.Bold).SprintFunc()
	_, _ = fmt.Fprintln(r.tw, header("SERVICE\tEXPECTED\tACTUAL\tSTATUS"))
	_, _ = fmt.Fprintln(r.tw, strings.Repeat("─", 55))

	green := color.New(color.FgGreen).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()

	allHealthy := true
	for _, s := range statuses {
		var statusStr string
		if s.Healthy {
			statusStr = green("✓ " + s.Status)
		} else {
			statusStr = red("✗ " + s.Status)
			allHealthy = false
		}
		_, _ = fmt.Fprintf(r.tw, "%s\t%d\t%s\t%s\n", s.Name, s.Expected, s.Actual, statusStr)
	}

	_ = r.tw.Flush()

	if allHealthy {
		_, _ = fmt.Fprintln(r.out, green("\nAll services healthy."))
	} else {
		_, _ = fmt.Fprintln(r.out, red("\nSome services are not healthy."))
	}
}
