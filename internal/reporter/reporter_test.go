package reporter

import (
	"bytes"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/firasmosbehi/portman/pkg/models"
)

func TestNewReporter(t *testing.T) {
	r := NewReporter()
	if r == nil {
		t.Fatal("expected reporter, got nil")
	}
}

func TestPrintPortTableNoColor(t *testing.T) {
	_ = os.Setenv("NO_COLOR", "1")
	defer func() { _ = os.Unsetenv("NO_COLOR") }()

	buf := new(bytes.Buffer)
	r := NewReporterWithWriter(buf)
	ports := []models.PortProcess{
		{Port: 3000, Protocol: "tcp", Process: "node", PID: 1234, User: "alice", Age: 2 * time.Hour},
		{Port: 5432, Protocol: "tcp", Process: "postgres", PID: 5678, User: "postgres", Age: 72 * time.Hour},
	}

	r.PrintPortTable(ports)
	out := buf.String()

	if !strings.Contains(out, "PORT") {
		t.Error("expected output to contain PORT header")
	}
	if !strings.Contains(out, "3000") {
		t.Error("expected output to contain port 3000")
	}
	if !strings.Contains(out, "5432") {
		t.Error("expected output to contain port 5432")
	}
}

func TestPrintPortTableWithColor(t *testing.T) {
	_ = os.Unsetenv("NO_COLOR")

	buf := new(bytes.Buffer)
	r := NewReporterWithWriter(buf)
	ports := []models.PortProcess{
		{Port: 3000, Protocol: "tcp", Process: "node", PID: 1234, User: "alice", Age: 2 * time.Hour},
	}

	r.PrintPortTable(ports)
	out := buf.String()

	// When colors are enabled, output should contain ANSI escape sequences
	if !strings.Contains(out, "\x1b[") {
		t.Error("expected output to contain ANSI escape codes when NO_COLOR is unset")
	}
}

func TestPrintPortTableEmpty(t *testing.T) {
	_ = os.Setenv("NO_COLOR", "1")
	defer func() { _ = os.Unsetenv("NO_COLOR") }()

	buf := new(bytes.Buffer)
	r := NewReporterWithWriter(buf)
	r.PrintPortTable([]models.PortProcess{})

	out := buf.String()
	if !strings.Contains(out, "PORT") {
		t.Error("expected output to contain PORT header")
	}
}

func TestPrintStatus(t *testing.T) {
	buf := new(bytes.Buffer)
	r := NewReporterWithWriter(buf)
	r.PrintStatus("all good")
	if buf.String() != "all good\n" {
		t.Errorf("unexpected output: %q", buf.String())
	}
}

func TestPrintError(t *testing.T) {
	buf := new(bytes.Buffer)
	r := NewReporterWithWriter(buf)
	r.PrintError(os.ErrNotExist)
	expected := "Error: file does not exist\n"
	if buf.String() != expected {
		t.Errorf("unexpected output: %q", buf.String())
	}
}

func TestPrintSuccess(t *testing.T) {
	buf := new(bytes.Buffer)
	r := NewReporterWithWriter(buf)
	r.PrintSuccess("done")
	expected := "✓ done\n"
	if buf.String() != expected {
		t.Errorf("unexpected output: %q", buf.String())
	}
}

func TestPrintServiceStatusTable(t *testing.T) {
	_ = os.Setenv("NO_COLOR", "1")
	defer func() { _ = os.Unsetenv("NO_COLOR") }()

	buf := new(bytes.Buffer)
	r := NewReporterWithWriter(buf)
	statuses := []models.ServiceStatus{
		{Name: "web", Expected: 3000, Actual: "3000", Status: "running", Healthy: true},
		{Name: "db", Expected: 5432, Actual: "-", Status: "not running", Healthy: false},
	}
	r.PrintServiceStatusTable(statuses)
	out := buf.String()

	if !strings.Contains(out, "SERVICE") {
		t.Error("expected output to contain SERVICE header")
	}
	if !strings.Contains(out, "web") {
		t.Error("expected output to contain 'web'")
	}
	if !strings.Contains(out, "db") {
		t.Error("expected output to contain 'db'")
	}
	if !strings.Contains(out, "Some services are not healthy") {
		t.Errorf("expected unhealthy summary, got: %q", out)
	}
}

func TestPrintServiceStatusTableAllHealthy(t *testing.T) {
	_ = os.Setenv("NO_COLOR", "1")
	defer func() { _ = os.Unsetenv("NO_COLOR") }()

	buf := new(bytes.Buffer)
	r := NewReporterWithWriter(buf)
	statuses := []models.ServiceStatus{
		{Name: "api", Expected: 3001, Actual: "3001", Status: "running", Healthy: true},
	}
	r.PrintServiceStatusTable(statuses)
	out := buf.String()

	if !strings.Contains(out, "All services healthy") {
		t.Errorf("expected healthy summary, got: %q", out)
	}
}
