package scanner

import (
	"errors"
	"testing"

	"github.com/firasmosbehi/portman/internal/platform"
	"github.com/firasmosbehi/portman/pkg/models"
)

type mockResolver struct {
	ports []models.PortProcess
	err   error
}

func (m *mockResolver) GetListeningPorts() ([]models.PortProcess, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.ports, nil
}

func (m *mockResolver) GetProcessByPort(port int) (*models.PortProcess, error) {
	if m.err != nil {
		return nil, m.err
	}
	for _, p := range m.ports {
		if p.Port == port {
			return &p, nil
		}
	}
	return nil, platform.ErrProcessNotFound
}

func TestNewScanner(t *testing.T) {
	s := NewScanner()
	if s == nil {
		t.Fatal("expected scanner, got nil")
	}
	if s.resolver == nil {
		t.Fatal("expected resolver to be set")
	}
}

func TestListPorts(t *testing.T) {
	expected := []models.PortProcess{
		{Port: 3000, Protocol: "tcp", Process: "node", PID: 1234},
		{Port: 5432, Protocol: "tcp", Process: "postgres", PID: 5678},
	}
	m := &mockResolver{ports: expected}
	s := NewScannerWithResolver(m)

	ports, err := s.ListPorts()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ports) != len(expected) {
		t.Fatalf("expected %d ports, got %d", len(expected), len(ports))
	}
}

func TestListPortsError(t *testing.T) {
	m := &mockResolver{err: errors.New("boom")}
	s := NewScannerWithResolver(m)

	_, err := s.ListPorts()
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestFindProcessByPortFound(t *testing.T) {
	expected := models.PortProcess{Port: 3000, Protocol: "tcp", Process: "node", PID: 1234}
	m := &mockResolver{ports: []models.PortProcess{expected}}
	s := NewScannerWithResolver(m)

	p, err := s.FindProcessByPort(3000)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Port != expected.Port {
		t.Errorf("expected port %d, got %d", expected.Port, p.Port)
	}
}

func TestFindProcessByPortNotFound(t *testing.T) {
	m := &mockResolver{ports: []models.PortProcess{}}
	s := NewScannerWithResolver(m)

	_, err := s.FindProcessByPort(3000)
	if !errors.Is(err, platform.ErrProcessNotFound) {
		t.Fatalf("expected ErrProcessNotFound, got %v", err)
	}
}

func TestFindProcessByPortError(t *testing.T) {
	m := &mockResolver{err: errors.New("boom")}
	s := NewScannerWithResolver(m)

	_, err := s.FindProcessByPort(3000)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestIsPortFree(t *testing.T) {
	m := &mockResolver{ports: []models.PortProcess{{Port: 3000, PID: 1234}}}
	s := NewScannerWithResolver(m)

	free, err := s.IsPortFree(3001)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !free {
		t.Error("expected port 3001 to be free")
	}

	free, err = s.IsPortFree(3000)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if free {
		t.Error("expected port 3000 to be in use")
	}
}

func TestIsPortFreeResolverError(t *testing.T) {
	m := &mockResolver{err: errors.New("boom")}
	s := NewScannerWithResolver(m)

	_, err := s.IsPortFree(3000)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestFindNextAvailablePort(t *testing.T) {
	m := &mockResolver{ports: []models.PortProcess{
		{Port: 3000, PID: 1},
		{Port: 3001, PID: 2},
		{Port: 3003, PID: 4},
	}}
	s := NewScannerWithResolver(m)

	port, err := s.FindNextAvailablePort(3000, 3005)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if port != 3002 {
		t.Errorf("expected port 3002, got %d", port)
	}
}

func TestFindNextAvailablePortNoneFree(t *testing.T) {
	m := &mockResolver{ports: []models.PortProcess{
		{Port: 3000, PID: 1},
		{Port: 3001, PID: 2},
	}}
	s := NewScannerWithResolver(m)

	_, err := s.FindNextAvailablePort(3000, 3001)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestFindNextAvailablePortResolverError(t *testing.T) {
	m := &mockResolver{err: errors.New("boom")}
	s := NewScannerWithResolver(m)

	_, err := s.FindNextAvailablePort(3000, 3001)
	if err == nil {
		t.Fatal("expected error")
	}
}
