package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func init() {
	SetVersionInfo("test-version", "test-commit", "test-date")
}

func executeCommand(args ...string) (string, error) {
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs(args)
	err := rootCmd.Execute()
	rootCmd.SetArgs([]string{})
	return buf.String(), err
}

func TestRootCmd(t *testing.T) {
	out, err := executeCommand()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "Usage:") {
		t.Error("expected help output to contain 'Usage:'")
	}
}

func TestListCmd(t *testing.T) {
	out, err := executeCommand("list")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Output is either a table with PORT header or a "no ports" message.
	if !strings.Contains(out, "PORT") && !strings.Contains(out, "No listening ports found") {
		t.Errorf("unexpected output: %q", out)
	}
}

func TestListCmdWithPort(t *testing.T) {
	out, err := executeCommand("list", "--port", "3000")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Port 3000 may be in use (table) or free (message); both contain "3000".
	if !strings.Contains(out, "3000") {
		t.Errorf("unexpected output: %q", out)
	}
}

func TestCheckCmd(t *testing.T) {
	out, err := executeCommand("check", "3000")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "Checking port 3000") {
		t.Errorf("unexpected output: %q", out)
	}
}

func TestCheckCmdInvalidPort(t *testing.T) {
	_, err := executeCommand("check", "abc")
	if err == nil {
		t.Fatal("expected error for invalid port")
	}
}

func TestCheckCmdMissingArg(t *testing.T) {
	_, err := executeCommand("check")
	if err == nil {
		t.Fatal("expected error for missing argument")
	}
}

func TestKillCmd(t *testing.T) {
	out, err := executeCommand("kill", "8080")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "8080") {
		t.Errorf("unexpected output: %q", out)
	}
}

func TestKillCmdForce(t *testing.T) {
	out, err := executeCommand("kill", "8080", "--force")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "force=true") {
		t.Errorf("unexpected output: %q", out)
	}
}

func TestKillCmdInvalidPort(t *testing.T) {
	_, err := executeCommand("kill", "abc")
	if err == nil {
		t.Fatal("expected error for invalid port")
	}
}

func TestNextCmd(t *testing.T) {
	out, err := executeCommand("next")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "3000-3100") {
		t.Errorf("unexpected output: %q", out)
	}
}

func TestNextCmdWithRange(t *testing.T) {
	out, err := executeCommand("next", "--range", "4000-4010")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "4000-4010") {
		t.Errorf("unexpected output: %q", out)
	}
}

func TestWatchCmd(t *testing.T) {
	out, err := executeCommand("watch", "3000")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "3000") {
		t.Errorf("unexpected output: %q", out)
	}
}

func TestWatchCmdInvalidPort(t *testing.T) {
	_, err := executeCommand("watch", "abc")
	if err == nil {
		t.Fatal("expected error for invalid port")
	}
}

func TestStatusCmd(t *testing.T) {
	out, err := executeCommand("status")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "Checking project status") {
		t.Errorf("unexpected output: %q", out)
	}
}

func TestVersionFlag(t *testing.T) {
	out, err := executeCommand("--version")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "portman") {
		t.Errorf("unexpected output: %q", out)
	}
}

func TestExecuteC(t *testing.T) {
	rootCmd.SetArgs([]string{"--version"})
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	err := ExecuteC()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "portman") {
		t.Errorf("unexpected output: %q", buf.String())
	}
	rootCmd.SetArgs([]string{})
}
