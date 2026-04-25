package reporter

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/fatih/color"
	"github.com/firasmosbehi/portman/pkg/models"
)

// Reporter handles terminal output formatting.
type Reporter struct {
	tw *tabwriter.Writer
}

// NewReporter creates a new reporter.
func NewReporter() *Reporter {
	return &Reporter{
		tw: tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0),
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
	fmt.Println(msg)
}

// PrintError prints an error message.
func (r *Reporter) PrintError(err error) {
	color.Red("Error: %v\n", err)
}

// PrintSuccess prints a success message.
func (r *Reporter) PrintSuccess(msg string) {
	color.Green("✓ %s\n", msg)
}
