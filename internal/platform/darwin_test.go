//go:build darwin

package platform

import (
	"errors"
	"strings"
	"testing"

	"github.com/firasmosbehi/portman/pkg/models"
)

func TestParseLsofOutput(t *testing.T) {
	input := `p1234
g1234
R1
cnode
u501
Lalice
f11
au
l 
tIPv4
G0x0;0x0
d0x0
o0t0
PTCP
n*:3000
TST=LISTEN
TQR=0
TQS=0
f12
au
l 
tIPv4
PTCP
n127.0.0.1:3001
TST=LISTEN
p5678
g5678
R1
cpostgres
u70
Lpostgres
f22
PTCP
n*:5432
TST=LISTEN
f23
PTCP
n192.168.1.5:54321->142.250.80.46:443
TST=ESTABLISHED
p9999
g9999
R1
credis-server
uredis
Lredis
f33
PUDP
n*:6379
`

	results, err := parseLsofOutput(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(results) != 4 {
		t.Fatalf("expected 4 results, got %d", len(results))
	}

	expected := []models.PortProcess{
		{Port: 3000, Protocol: "tcp", Process: "node", PID: 1234, User: "alice"},
		{Port: 3001, Protocol: "tcp", Process: "node", PID: 1234, User: "alice"},
		{Port: 5432, Protocol: "tcp", Process: "postgres", PID: 5678, User: "postgres"},
		{Port: 6379, Protocol: "udp", Process: "redis-server", PID: 9999, User: "redis"},
	}

	for i, exp := range expected {
		if results[i].Port != exp.Port {
			t.Errorf("result[%d].Port = %d, want %d", i, results[i].Port, exp.Port)
		}
		if results[i].Protocol != exp.Protocol {
			t.Errorf("result[%d].Protocol = %s, want %s", i, results[i].Protocol, exp.Protocol)
		}
		if results[i].Process != exp.Process {
			t.Errorf("result[%d].Process = %s, want %s", i, results[i].Process, exp.Process)
		}
		if results[i].PID != exp.PID {
			t.Errorf("result[%d].PID = %d, want %d", i, results[i].PID, exp.PID)
		}
		if results[i].User != exp.User {
			t.Errorf("result[%d].User = %s, want %s", i, results[i].User, exp.User)
		}
	}
}

func TestParseLsofOutputEmpty(t *testing.T) {
	results, err := parseLsofOutput(strings.NewReader(""))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Fatalf("expected 0 results, got %d", len(results))
	}
}

func TestParseLsofOutputTCPNoListen(t *testing.T) {
	input := `p1234
cnode
Lalice
f11
PTCP
n*:3000
TST=ESTABLISHED
`
	results, err := parseLsofOutput(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Fatalf("expected 0 results for non-listening TCP, got %d", len(results))
	}
}

func TestParseLsofOutputFlushPendingEmptyAddress(t *testing.T) {
	input := `p1234
cnode
Lalice
f11
PTCP
TST=LISTEN
`
	results, err := parseLsofOutput(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Fatalf("expected 0 results when address is empty, got %d", len(results))
	}
}

func TestParseLsofOutputFlushPendingInvalidPort(t *testing.T) {
	input := `p1234
cnode
Lalice
f11
PTCP
n*:abc
TST=LISTEN
`
	results, err := parseLsofOutput(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Fatalf("expected 0 results for invalid port, got %d", len(results))
	}
}

func TestExtractPortFromLsofAddr(t *testing.T) {
	tests := []struct {
		addr string
		want int
	}{
		{"*:3000", 3000},
		{"0.0.0.0:5432", 5432},
		{"[::]:6379", 6379},
		{"127.0.0.1:8080", 8080},
		{"[fe80::1]:3000", 3000},
		{"no-port", 0},
		{"", 0},
		{":abc", 0},
	}

	for _, tt := range tests {
		got := extractPortFromLsofAddr(tt.addr)
		if got != tt.want {
			t.Errorf("extractPortFromLsofAddr(%q) = %d, want %d", tt.addr, got, tt.want)
		}
	}
}

func TestDarwinResolverGetListeningPorts(t *testing.T) {
	old := lsofRunner
	defer func() { lsofRunner = old }()

	lsofRunner = func() ([]byte, error) {
		return []byte("p1234\ncnode\nLalice\nf11\nPTCP\nn*:3000\nTST=LISTEN\n"), nil
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

func TestDarwinResolverGetListeningPortsError(t *testing.T) {
	old := lsofRunner
	defer func() { lsofRunner = old }()

	lsofRunner = func() ([]byte, error) {
		return nil, errors.New("lsof not found")
	}

	r := NewResolver()
	_, err := r.GetListeningPorts()
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestDarwinResolverGetListeningPortsPartialOutput(t *testing.T) {
	old := lsofRunner
	defer func() { lsofRunner = old }()

	lsofRunner = func() ([]byte, error) {
		return []byte("p1234\ncnode\nLalice\nf11\nPTCP\nn*:3000\nTST=LISTEN\n"), errors.New("exit 1")
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

func TestDarwinResolverGetProcessByPortFound(t *testing.T) {
	old := lsofRunner
	defer func() { lsofRunner = old }()

	lsofRunner = func() ([]byte, error) {
		return []byte("p1234\ncnode\nLalice\nf11\nPTCP\nn*:3000\nTST=LISTEN\n"), nil
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

func TestDarwinResolverGetProcessByPortNotFound(t *testing.T) {
	old := lsofRunner
	defer func() { lsofRunner = old }()

	lsofRunner = func() ([]byte, error) {
		return []byte("p1234\ncnode\nLalice\nf11\nPTCP\nn*:3000\nTST=LISTEN\n"), nil
	}

	r := NewResolver()
	_, err := r.GetProcessByPort(9999)
	if !errors.Is(err, ErrProcessNotFound) {
		t.Fatalf("expected ErrProcessNotFound, got %v", err)
	}
}

func TestDarwinResolverGetProcessByPortError(t *testing.T) {
	old := lsofRunner
	defer func() { lsofRunner = old }()

	lsofRunner = func() ([]byte, error) {
		return nil, errors.New("lsof failed")
	}

	r := NewResolver()
	_, err := r.GetProcessByPort(3000)
	if err == nil {
		t.Fatal("expected error")
	}
}

type failingReader struct{}

func (f failingReader) Read(p []byte) (int, error) {
	return 0, errors.New("read error")
}

func TestParseLsofOutputScannerError(t *testing.T) {
	_, err := parseLsofOutput(&failingReader{})
	if err == nil {
		t.Fatal("expected scanner error")
	}
}

func TestParseLsofOutputShortLines(t *testing.T) {
	input := "p1234\ncnode\nLalice\n\nn\nPTCP\nn*:3000\nTST=LISTEN\n"
	results, err := parseLsofOutput(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
}
