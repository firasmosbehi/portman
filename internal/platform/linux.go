//go:build linux

package platform

import (
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/firasmosbehi/portman/pkg/models"
)

// ssRunner abstracts the ss command for testability.
var ssRunner = func() ([]byte, error) {
	return exec.Command("ss", "-tulnp").Output()
}

// Resolver implements port and process resolution for Linux.
type Resolver struct{}

// NewResolver creates a new Linux platform resolver.
func NewResolver() *Resolver {
	return &Resolver{}
}

// GetListeningPorts returns all listening ports with process info.
func (r *Resolver) GetListeningPorts() ([]models.PortProcess, error) {
	out, err := ssRunner()
	if err != nil {
		if len(out) == 0 {
			return nil, fmt.Errorf("ss failed: %w", err)
		}
	}
	return parseSsOutput(string(out))
}

// GetProcessByPort returns the process using the given port.
func (r *Resolver) GetProcessByPort(port int) (*models.PortProcess, error) {
	ports, err := r.GetListeningPorts()
	if err != nil {
		return nil, err
	}
	for _, p := range ports {
		if p.Port == port {
			return &p, nil
		}
	}
	return nil, fmt.Errorf("%w %d", ErrProcessNotFound, port)
}

// processRe extracts the first process name and PID from ss -p output.
// Example: users:(("sshd",pid=1234,fd=3))
var processRe = regexp.MustCompile(`users:\(\("([^"]+)",pid=(\d+)`)

// parseSsOutput parses the output of ss -tulnp.
func parseSsOutput(output string) ([]models.PortProcess, error) {
	var results []models.PortProcess
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "Netid") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 6 {
			continue
		}

		netid := fields[0]
		state := fields[1]
		localAddr := fields[4]

		// Filter: TCP must be LISTEN, UDP must be UNCONN.
		if netid == "tcp" && state != "LISTEN" {
			continue
		}
		if netid == "udp" && state != "UNCONN" {
			continue
		}

		port := extractPortFromSsAddr(localAddr)
		if port == 0 {
			continue
		}

		var process string
		var pid int
		if len(fields) > 6 {
			matches := processRe.FindStringSubmatch(fields[6])
			if len(matches) >= 3 {
				process = matches[1]
				pid, _ = strconv.Atoi(matches[2])
			}
		}

		results = append(results, models.PortProcess{
			Port:     port,
			Protocol: strings.ToLower(netid),
			Process:  process,
			PID:      pid,
		})
	}

	return results, nil
}

// extractPortFromSsAddr extracts the port from an ss address like 0.0.0.0:22 or [::]:22.
func extractPortFromSsAddr(addr string) int {
	idx := strings.LastIndex(addr, ":")
	if idx == -1 {
		return 0
	}
	portStr := addr[idx+1:]
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return 0
	}
	return port
}
