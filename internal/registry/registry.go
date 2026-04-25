package registry

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Service represents a declared service in portman.yml.
type Service struct {
	Name        string `yaml:"name"`
	Port        int    `yaml:"port"`
	Command     string `yaml:"command,omitempty"`
	HealthCheck string `yaml:"health_check,omitempty"`
}

// Registry holds the parsed portman.yml configuration.
type Registry struct {
	Services []Service `yaml:"services"`
}

// Load reads and parses a portman.yml file from the given path.
func Load(path string) (*Registry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading portman.yml: %w", err)
	}

	var reg Registry
	if err := yaml.Unmarshal(data, &reg); err != nil {
		return nil, fmt.Errorf("parsing portman.yml: %w", err)
	}

	for i, svc := range reg.Services {
		if svc.Name == "" {
			return nil, fmt.Errorf("service at index %d is missing required field: name", i)
		}
		if svc.Port == 0 {
			return nil, fmt.Errorf("service %q is missing required field: port", svc.Name)
		}
	}

	return &reg, nil
}
