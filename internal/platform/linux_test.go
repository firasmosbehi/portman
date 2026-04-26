//go:build linux

package platform

import (
	"errors"
	"testing"
)

func TestParseSsOutput(t *testing.T) {
	input := `Netid  State   Recv-Q  Send-Q  Local Address:Port   Peer Address:Port  Process
tcp    LISTEN  0       128     0.0.0.0:22           0.0.0.0:*          users:(("sshd",pid=1234,fd=3))
tcp    LISTEN  0       128     [::]:22              [::]:*             users:(("sshd",pid=1234,fd=4))
udp    UNCONN  0       0       0.0.0.0:68           0.0.0.0:*          users:(("dhclient",pid=567,fd=6))
tcp    LISTEN  0       128     127.0.0.1:5432       0.0.0.0:*          users:(("postgres",pid=999,fd=5))
tcp    LISTEN  0       128     127.0.0.1:8080       0.0.0.0:*
tcp    ESTAB   0       0       192.168.1.5:54321    142.250.80.46:443  users:(("chrome",pid=1111,fd=22))
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

func TestParseSsOutputEmpty(t *testing.T) {
	results, err := parseSsOutput("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Fatalf("expected 0 results, got %d", len(results))
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
		{":abc", 0},
	}

	for _, tt := range tests {
		got := extractPortFromSsAddr(tt.addr)
		if got != tt.want {
			t.Errorf("extractPortFromSsAddr(%q) = %d, want %d", tt.addr, got, tt.want)
		}
	}
}

func TestLinuxResolverGetListeningPorts(t *testing.T) {
	old := ssRunner
	defer func() { ssRunner = old }()

	ssRunner = func() ([]byte, error) {
		return []byte("Netid State Recv-Q Send-Q Local Address:Port Peer Address:Port Process\ntcp LISTEN 0 128 0.0.0.0:3000 0.0.0.0:* users:((\"node\",pid=1234,fd=3))\n"), nil
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
}

func TestLinuxResolverGetListeningPortsError(t *testing.T) {
	old := ssRunner
	defer func() { ssRunner = old }()

	ssRunner = func() ([]byte, error) {
		return nil, errors.New("ss not found")
	}

	r := NewResolver()
	_, err := r.GetListeningPorts()
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestLinuxResolverGetListeningPortsPartialOutput(t *testing.T) {
	old := ssRunner
	defer func() { ssRunner = old }()

	ssRunner = func() ([]byte, error) {
		return []byte("Netid State Recv-Q Send-Q Local Address:Port Peer Address:Port Process\ntcp LISTEN 0 128 0.0.0.0:3000 0.0.0.0:* users:((\"node\",pid=1234,fd=3))\n"), errors.New("exit 1")
	}

	r := NewResolver()
	ports, err := r.GetListeningPorts()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ports) != 1 {
		t.Fatalf("expected 1 port, got %d", len(ports))
	}
}

func TestLinuxResolverGetProcessByPortFound(t *testing.T) {
	old := ssRunner
	defer func() { ssRunner = old }()

	ssRunner = func() ([]byte, error) {
		return []byte("Netid State Recv-Q Send-Q Local Address:Port Peer Address:Port Process\ntcp LISTEN 0 128 0.0.0.0:3000 0.0.0.0:* users:((\"node\",pid=1234,fd=3))\n"), nil
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

func TestLinuxResolverGetProcessByPortNotFound(t *testing.T) {
	old := ssRunner
	defer func() { ssRunner = old }()

	ssRunner = func() ([]byte, error) {
		return []byte("Netid State Recv-Q Send-Q Local Address:Port Peer Address:Port Process\ntcp LISTEN 0 128 0.0.0.0:3000 0.0.0.0:* users:((\"node\",pid=1234,fd=3))\n"), nil
	}

	r := NewResolver()
	_, err := r.GetProcessByPort(9999)
	if !errors.Is(err, ErrProcessNotFound) {
		t.Fatalf("expected ErrProcessNotFound, got %v", err)
	}
}
