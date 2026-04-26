package scanner

import (
	"errors"
	"fmt"

	"github.com/firasmosbehi/portman/internal/platform"
	"github.com/firasmosbehi/portman/pkg/models"
)

// PortResolver defines the interface for platform-specific port resolution.
type PortResolver interface {
	GetListeningPorts() ([]models.PortProcess, error)
	GetProcessByPort(port int) (*models.PortProcess, error)
}

// Scanner provides port scanning and process lookup.
type Scanner struct {
	resolver PortResolver
}

// NewScanner creates a new scanner using the platform resolver.
func NewScanner() *Scanner {
	return NewScannerWithResolver(platform.NewResolver())
}

// NewScannerWithResolver creates a new scanner with the given resolver.
func NewScannerWithResolver(r PortResolver) *Scanner {
	return &Scanner{resolver: r}
}

// ListPorts returns all listening ports with process info.
func (s *Scanner) ListPorts() ([]models.PortProcess, error) {
	return s.resolver.GetListeningPorts()
}

// FindProcessByPort returns the process using the given port.
func (s *Scanner) FindProcessByPort(port int) (*models.PortProcess, error) {
	return s.resolver.GetProcessByPort(port)
}

// IsPortFree reports whether the given port is not in use.
func (s *Scanner) IsPortFree(port int) (bool, error) {
	_, err := s.resolver.GetProcessByPort(port)
	if err == nil {
		return false, nil
	}
	if errors.Is(err, platform.ErrProcessNotFound) {
		return true, nil
	}
	return false, err
}

// FindNextAvailablePort scans the given range and returns the first free port.
func (s *Scanner) FindNextAvailablePort(start, end int) (int, error) {
	for port := start; port <= end; port++ {
		free, err := s.IsPortFree(port)
		if err != nil {
			return 0, err
		}
		if free {
			return port, nil
		}
	}
	return 0, fmt.Errorf("no available ports in range %d-%d", start, end)
}
