//go:build windows

package platform

import (
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
	}

	for _, tt := range tests {
		got := extractPortFromNetstatAddr(tt.addr)
		if got != tt.want {
			t.Errorf("extractPortFromNetstatAddr(%q) = %d, want %d", tt.addr, got, tt.want)
		}
	}
}
