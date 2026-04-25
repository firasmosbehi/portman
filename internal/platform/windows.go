//go:build windows

package platform

import (
	"fmt"

	"github.com/firasmosbehi/portman/pkg/models"
)

// Resolver implements port and process resolution for Windows.
type Resolver struct{}

// NewResolver creates a new Windows platform resolver.
func NewResolver() *Resolver {
	return &Resolver{}
}

// GetListeningPorts returns all listening ports with process info.
func (r *Resolver) GetListeningPorts() ([]models.PortProcess, error) {
	return nil, fmt.Errorf("windows resolver not yet implemented")
}

// GetProcessByPort returns the process using the given port.
func (r *Resolver) GetProcessByPort(port int) (*models.PortProcess, error) {
	return nil, fmt.Errorf("windows resolver not yet implemented")
}
