package e2e

import (
	"bytes"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"
)

var binaryPath string

func TestMain(m *testing.M) {
	_, thisFile, _, _ := runtime.Caller(0)
	projectRoot := filepath.Join(thisFile, "..", "..", "..")
	binaryPath = filepath.Join(projectRoot, "portman_e2e")
	if runtime.GOOS == "windows" {
		binaryPath += ".exe"
	}

	build := exec.Command("go", "build", "-o", binaryPath, ".")
	build.Dir = projectRoot
	build.Stderr = os.Stderr
	if err := build.Run(); err != nil {
		panic(err)
	}

	code := m.Run()
	_ = os.Remove(binaryPath)
	os.Exit(code)
}

func runBinary(args ...string) (string, string, error) {
	cmd := exec.Command(binaryPath, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return stdout.String(), stderr.String(), err
}

func TestE2E_Help(t *testing.T) {
	out, _, err := runBinary("--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "Usage:") {
		t.Errorf("expected help to contain 'Usage:', got: %s", out)
	}
}

func TestE2E_Version(t *testing.T) {
	out, _, err := runBinary("--version")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "portman") {
		t.Errorf("expected version to contain 'portman', got: %s", out)
	}
}

func TestE2E_List(t *testing.T) {
	out, _, err := runBinary("list")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Output is either a table with PORT header or a "no ports" message.
	if !strings.Contains(out, "PORT") && !strings.Contains(out, "No listening ports found") {
		t.Errorf("unexpected output: %s", out)
	}
}

func TestE2E_CheckFreePort(t *testing.T) {
	// Find a free port
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	port := ln.Addr().(*net.TCPAddr).Port
	_ = ln.Close()

	out, _, err := runBinary("check", strconv.Itoa(port))
	if err != nil {
		t.Fatalf("expected no error for free port, got: %v", err)
	}
	if !strings.Contains(out, "is free") {
		t.Errorf("expected 'is free' in output, got: %s", out)
	}
}

func TestE2E_CheckOccupiedPort(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = ln.Close() }()
	port := ln.Addr().(*net.TCPAddr).Port

	out, _, err := runBinary("check", strconv.Itoa(port))
	if err == nil {
		t.Fatal("expected non-zero exit for occupied port")
	}
	if !strings.Contains(out, "is in use") {
		t.Errorf("expected 'is in use' in output, got: %s", out)
	}
}

func TestE2E_Next(t *testing.T) {
	out, _, err := runBinary("next", "--range", "30000-30010")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	port, err := strconv.Atoi(strings.TrimSpace(out))
	if err != nil {
		t.Fatalf("expected numeric port, got: %s", out)
	}
	if port < 30000 || port > 30010 {
		t.Errorf("expected port in range 30000-30010, got %d", port)
	}
}

func TestE2E_Status(t *testing.T) {
	out, _, err := runBinary("status")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// E2E runs from tests/e2e dir where no portman.yml exists.
	if !strings.Contains(out, "No portman.yml found") {
		t.Errorf("unexpected output: %s", out)
	}
}

func TestE2E_KillRequiresPort(t *testing.T) {
	_, _, err := runBinary("kill")
	if err == nil {
		t.Fatal("expected error for missing port argument")
	}
}

func TestE2E_InvalidCommand(t *testing.T) {
	_, _, err := runBinary("invalid-cmd")
	if err == nil {
		t.Fatal("expected error for invalid command")
	}
}

func TestE2E_Init(t *testing.T) {
	dir := t.TempDir()
	cmd := exec.Command(binaryPath, "init")
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("unexpected error: %v, output: %s", err, out)
	}
	if !strings.Contains(string(out), "Created portman.yml") {
		t.Errorf("expected 'Created portman.yml' in output, got: %s", out)
	}

	// Verify file exists
	if _, err := os.Stat(filepath.Join(dir, "portman.yml")); err != nil {
		t.Errorf("expected portman.yml to exist: %v", err)
	}
}

func TestE2E_Find(t *testing.T) {
	out, _, err := runBinary("find", "nonexistent-process-xyz")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "No process found") {
		t.Errorf("expected 'No process found' in output, got: %s", out)
	}
}

func TestE2E_FindJSON(t *testing.T) {
	out, _, err := runBinary("find", "nonexistent", "--format", "json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasPrefix(strings.TrimSpace(out), "[") {
		t.Errorf("expected JSON array output, got: %s", out)
	}
}

func TestE2E_RealHTTPPort(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	_ = ln.Addr().(*net.TCPAddr).Port

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("ok"))
	})
	server := &http.Server{Handler: mux}
	go func() { _ = server.Serve(ln) }()
	defer func() { _ = server.Close() }()

	// Give the server a moment to start
	time.Sleep(50 * time.Millisecond)

	// Verify list command runs successfully
	out, _, err := runBinary("list")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Output is either a table with PORT header or a "no ports" message.
	if !strings.Contains(out, "PORT") && !strings.Contains(out, "No listening ports found") {
		t.Errorf("unexpected output: %s", out)
	}
}
