//go:build linux

package platform

import (
	"testing"
)

func TestParseSsOutput(t *testing.T) {
	input := `Netid  State   Recv-Q  Send-Q  Local Address:Port   Peer Address:Port  Process
tcp    LISTEN  0       128     0.0.0.0:22           0.0.0.0:*          users:(("sshd",pid=1234,fd=3))
tcp    LISTEN  0       128     [::]:22              [::]:*             users:(("sshd",pid=1234,fd=4))
udp    UNCONN  0       0       0.0.0.0:68           0.0.0.0:*          users:(("dhclient",pid=567,fd=6))
tcp    LISTEN  0       128     127.0.0.1:5432       0.0.0.0:*          users:(("postgres",pid=999,fd=5))
tcp    LISTEN  0       128     127.0.0.1:8080       0.0.0.0:*
`

	results, err := parseSsOutput(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(results) != 5 {
		t.Fatalf("expected 5 results, got %d", len(results))
	}

	tests := []struct {
		idx      int
		port     int
		protocol string
		process  string
		pid      int
	}{
		{0, 22, "tcp", "sshd", 1234},
		{1, 22, "tcp", "sshd", 1234},
		{2, 68, "udp", "dhclient", 567},
		{3, 5432, "tcp", "postgres", 999},
		{4, 8080, "tcp", "", 0},
	}

	for _, tt := range tests {
		r := results[tt.idx]
		if r.Port != tt.port {
			t.Errorf("result[%d].Port = %d, want %d", tt.idx, r.Port, tt.port)
		}
		if r.Protocol != tt.protocol {
			t.Errorf("result[%d].Protocol = %s, want %s", tt.idx, r.Protocol, tt.protocol)
		}
		if r.Process != tt.process {
			t.Errorf("result[%d].Process = %s, want %s", tt.idx, r.Process, tt.process)
		}
		if r.PID != tt.pid {
			t.Errorf("result[%d].PID = %d, want %d", tt.idx, r.PID, tt.pid)
		}
	}
}

func TestExtractPortFromSsAddr(t *testing.T) {
	tests := []struct {
		addr string
		want int
	}{
		{"0.0.0.0:22", 22},
		{"[::]:22", 22},
		{"127.0.0.1:5432", 5432},
		{"[::1]:3000", 3000},
		{"no-port", 0},
		{"", 0},
	}

	for _, tt := range tests {
		got := extractPortFromSsAddr(tt.addr)
		if got != tt.want {
			t.Errorf("extractPortFromSsAddr(%q) = %d, want %d", tt.addr, got, tt.want)
		}
	}
}
