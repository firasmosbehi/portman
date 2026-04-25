//go:build linux

package platform

import (
	"fmt"

	"github.com/firasmosbehi/portman/pkg/models"
)

// Resolver implements port and process resolution for Linux.
type Resolver struct{}

// NewResolver creates a new Linux platform resolver.
func NewResolver() *Resolver {
	return &Resolver{}
}

// GetListeningPorts returns all listening ports with process info.
func (r *Resolver) GetListeningPorts() ([]models.PortProcess, error) {
	return nil, fmt.Errorf("linux resolver not yet implemented")
}

// GetProcessByPort returns the process using the given port.
func (r *Resolver) GetProcessByPort(port int) (*models.PortProcess, error) {
	return nil, fmt.Errorf("linux resolver not yet implemented")
}
