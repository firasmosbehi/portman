package cmd

import (
	"bufio"
	"bytes"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

func init() {
	SetVersionInfo("test-version", "test-commit", "test-date")
}

func executeCommand(args ...string) (string, error) {
	return executeCommandStdin(nil, args...)
}

func executeCommandStdin(stdin io.Reader, args ...string) (string, error) {
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	if stdin != nil {
		rootCmd.SetIn(stdin)
	}
	rootCmd.SetArgs(args)
	err := rootCmd.Execute()
	rootCmd.SetArgs([]string{})
	rootCmd.SetIn(nil)
	// Reset persistent flags so they don't leak across tests.
	killForceFlag = false
	listPortFlag = 0
	listFormatFlag = "table"
	nextRangeFlag = "3000-3100"
	watchIntervalFlag = 0
	return buf.String(), err
}

// startTestListener starts a child Go process that listens on a random TCP port.
func startTestListener(t *testing.T) (int, *exec.Cmd) {
	code := `package main
import (
	"fmt"
	"net"
	"net/http"
	"os"
)
func main() {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		os.Exit(1)
	}
	fmt.Println(ln.Addr().(*net.TCPAddr).Port)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {})
	http.Serve(ln, nil)
}`
	dir := t.TempDir()
	src := filepath.Join(dir, "main.go")
	if err := os.WriteFile(src, []byte(code), 0644); err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command("go", "run", src)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}

	reader := bufio.NewReader(stdout)
	line, err := reader.ReadString('\n')
	if err != nil {
		t.Fatal(err)
	}
	port, err := strconv.Atoi(strings.TrimSpace(line))
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(50 * time.Millisecond)
	return port, cmd
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

func TestListCmdJSON(t *testing.T) {
	out, err := executeCommand("list", "--format", "json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasPrefix(strings.TrimSpace(out), "[") {
		t.Errorf("expected JSON array output, got: %q", out)
	}
}

func TestListCmdJSONPortNotFound(t *testing.T) {
	out, err := executeCommand("list", "--port", "49999", "--format", "json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.TrimSpace(out) != "[]" {
		t.Errorf("expected empty JSON array, got: %q", out)
	}
}

func TestCheckCmd(t *testing.T) {
	// Port 3000 may be free or in use depending on the system.
	out, err := executeCommand("check", "3000")
	// Output should contain either "free" or "in use".
	if !strings.Contains(out, "Port 3000 is free") && !strings.Contains(out, "Port 3000 is in use") {
		t.Errorf("unexpected output: %q", out)
	}
	// We don't assert err here because it depends on whether port 3000 is in use.
	_ = err
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

func TestKillCmdPortNotFound(t *testing.T) {
	// Use a high ephemeral port that is almost certainly free.
	_, err := executeCommand("kill", "49999")
	if err == nil {
		t.Fatal("expected error when port is not in use")
	}
}

func TestKillCmdForceNotFound(t *testing.T) {
	_, err := executeCommand("kill", "49999", "--force")
	if err == nil {
		t.Fatal("expected error when port is not in use")
	}
}

func TestKillCmdAbort(t *testing.T) {
	port, listener := startTestListener(t)
	defer func() { _ = listener.Process.Kill() }()

	out, err := executeCommandStdin(strings.NewReader("n\n"), "kill", strconv.Itoa(port))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "Aborted") {
		t.Errorf("expected 'Aborted' in output, got: %q", out)
	}
}

func TestKillCmdForce(t *testing.T) {
	port, listener := startTestListener(t)
	defer func() { _ = listener.Process.Kill() }()

	out, err := executeCommand("kill", strconv.Itoa(port), "--force")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "Killed") {
		t.Errorf("expected 'Killed' in output, got: %q", out)
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
	// Output should be a port number in the default range.
	port, err := strconv.Atoi(strings.TrimSpace(out))
	if err != nil {
		t.Fatalf("expected numeric port, got: %q", out)
	}
	if port < 3000 || port > 3100 {
		t.Errorf("expected port in range 3000-3100, got %d", port)
	}
}

func TestNextCmdWithRange(t *testing.T) {
	out, err := executeCommand("next", "--range", "4000-4010")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	port, err := strconv.Atoi(strings.TrimSpace(out))
	if err != nil {
		t.Fatalf("expected numeric port, got: %q", out)
	}
	if port < 4000 || port > 4010 {
		t.Errorf("expected port in range 4000-4010, got %d", port)
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

func TestStatusCmdNoConfig(t *testing.T) {
	out, err := executeCommand("status")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "No portman.yml found") {
		t.Errorf("unexpected output: %q", out)
	}
}

func TestStatusCmdWithConfig(t *testing.T) {
	content := `services:
  - name: test-svc
    port: 49998
`
	if err := os.WriteFile("portman.yml", []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Remove("portman.yml") }()

	out, err := executeCommand("status")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "test-svc") {
		t.Errorf("expected 'test-svc' in output, got: %q", out)
	}
	if !strings.Contains(out, "not running") {
		t.Errorf("expected 'not running' in output, got: %q", out)
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
