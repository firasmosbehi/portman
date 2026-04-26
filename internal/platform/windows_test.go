//go:build windows

package platform

import (
	"errors"
	"testing"
)

func TestParseNetstatOutput(t *testing.T) {
	input := `
  Proto  Local Address          Foreign Address        State           PID
  TCP    0.0.0.0:135            0.0.0.0:0              LISTENING       1234
  TCP    127.0.0.1:3000         0.0.0.0:0              LISTENING       5678
  TCP    192.168.1.5:54321      142.250.80.46:443      ESTABLISHED     5678
  UDP    0.0.0.0:68             *:*                                    9012
  UDP    [::]:53                *:*                                    4
`

	results, err := parseNetstatOutput(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(results) != 4 {
		t.Fatalf("expected 4 results, got %d", len(results))
	}

	tests := []struct {
		idx      int
		port     int
		protocol string
		pid      int
	}{
		{0, 135, "tcp", 1234},
		{1, 3000, "tcp", 5678},
		{2, 68, "udp", 9012},
		{3, 53, "udp", 4},
	}

	for _, tt := range tests {
		r := results[tt.idx]
		if r.Port != tt.port {
			t.Errorf("result[%d].Port = %d, want %d", tt.idx, r.Port, tt.port)
		}
		if r.Protocol != tt.protocol {
			t.Errorf("result[%d].Protocol = %s, want %s", tt.idx, r.Protocol, tt.protocol)
		}
		if r.PID != tt.pid {
			t.Errorf("result[%d].PID = %d, want %d", tt.idx, r.PID, tt.pid)
		}
	}
}

func TestParseNetstatOutputEmpty(t *testing.T) {
	results, err := parseNetstatOutput("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Fatalf("expected 0 results, got %d", len(results))
	}
}

func TestParseTasklistOutput(t *testing.T) {
	input := `"Image Name","PID","Session Name","Session#","Mem Usage"
"System Idle Process","0","Services","0","4 K"
"System","4","Services","0","1,200 K"
"node.exe","5678","Console","1","45,678 K"
"postgres.exe","999","Console","1","12,345 K"
`

	pidToName := parseTasklistOutput(input)

	if len(pidToName) != 4 {
		t.Fatalf("expected 4 entries, got %d", len(pidToName))
	}

	tests := map[int]string{
		0:    "System Idle Process",
		4:    "System",
		5678: "node.exe",
		999:  "postgres.exe",
	}

	for pid, want := range tests {
		if got := pidToName[pid]; got != want {
			t.Errorf("pidToName[%d] = %q, want %q", pid, got, want)
		}
	}
}

func TestExtractPortFromNetstatAddr(t *testing.T) {
	tests := []struct {
		addr string
		want int
	}{
		{"0.0.0.0:135", 135},
		{"127.0.0.1:3000", 3000},
		{"[::]:53", 53},
		{"[::1]:8080", 8080},
		{"no-port", 0},
		{"", 0},
		{":abc", 0},
	}

	for _, tt := range tests {
		got := extractPortFromNetstatAddr(tt.addr)
		if got != tt.want {
			t.Errorf("extractPortFromNetstatAddr(%q) = %d, want %d", tt.addr, got, tt.want)
		}
	}
}

func TestWindowsResolverGetListeningPorts(t *testing.T) {
	oldNetstat := netstatRunner
	oldTasklist := tasklistRunner
	defer func() {
		netstatRunner = oldNetstat
		tasklistRunner = oldTasklist
	}()

	netstatRunner = func() ([]byte, error) {
		return []byte("Proto Local Address Foreign Address State PID\nTCP 0.0.0.0:3000 0.0.0.0:0 LISTENING 1234\n"), nil
	}
	tasklistRunner = func() ([]byte, error) {
		return []byte("\"Image Name\",\"PID\"\n\"node.exe\",\"1234\"\n"), nil
	}

	r := NewResolver()
	ports, err := r.GetListeningPorts()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ports) != 1 {
		t.Fatalf("expected 1 port, got %d", len(ports))
	}
	if ports[0].Port != 3000 {
		t.Errorf("expected port 3000, got %d", ports[0].Port)
	}
	if ports[0].Process != "node.exe" {
		t.Errorf("expected process node.exe, got %s", ports[0].Process)
	}
}

func TestWindowsResolverGetListeningPortsNetstatError(t *testing.T) {
	old := netstatRunner
	defer func() { netstatRunner = old }()

	netstatRunner = func() ([]byte, error) {
		return nil, errors.New("netstat not found")
	}

	r := NewResolver()
	_, err := r.GetListeningPorts()
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestWindowsResolverGetListeningPortsTasklistError(t *testing.T) {
	oldNetstat := netstatRunner
	oldTasklist := tasklistRunner
	defer func() {
		netstatRunner = oldNetstat
		tasklistRunner = oldTasklist
	}()

	netstatRunner = func() ([]byte, error) {
		return []byte("Proto Local Address Foreign Address State PID\nTCP 0.0.0.0:3000 0.0.0.0:0 LISTENING 1234\n"), nil
	}
	tasklistRunner = func() ([]byte, error) {
		return nil, errors.New("tasklist failed")
	}

	r := NewResolver()
	ports, err := r.GetListeningPorts()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ports) != 1 {
		t.Fatalf("expected 1 port, got %d", len(ports))
	}
	// Process name should be empty when tasklist fails
	if ports[0].Process != "" {
		t.Errorf("expected empty process name, got %s", ports[0].Process)
	}
}

func TestWindowsResolverGetProcessByPortFound(t *testing.T) {
	oldNetstat := netstatRunner
	oldTasklist := tasklistRunner
	defer func() {
		netstatRunner = oldNetstat
		tasklistRunner = oldTasklist
	}()

	netstatRunner = func() ([]byte, error) {
		return []byte("Proto Local Address Foreign Address State PID\nTCP 0.0.0.0:3000 0.0.0.0:0 LISTENING 1234\n"), nil
	}
	tasklistRunner = func() ([]byte, error) {
		return []byte("\"Image Name\",\"PID\"\n\"node.exe\",\"1234\"\n"), nil
	}

	r := NewResolver()
	p, err := r.GetProcessByPort(3000)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Port != 3000 {
		t.Errorf("expected port 3000, got %d", p.Port)
	}
}

func TestWindowsResolverGetProcessByPortNotFound(t *testing.T) {
	oldNetstat := netstatRunner
	oldTasklist := tasklistRunner
	defer func() {
		netstatRunner = oldNetstat
		tasklistRunner = oldTasklist
	}()

	netstatRunner = func() ([]byte, error) {
		return []byte("Proto Local Address Foreign Address State PID\nTCP 0.0.0.0:3000 0.0.0.0:0 LISTENING 1234\n"), nil
	}
	tasklistRunner = func() ([]byte, error) {
		return []byte("\"Image Name\",\"PID\"\n\"node.exe\",\"1234\"\n"), nil
	}

	r := NewResolver()
	_, err := r.GetProcessByPort(9999)
	if !errors.Is(err, ErrProcessNotFound) {
		t.Fatalf("expected ErrProcessNotFound, got %v", err)
	}
}
