//go:build windows

package platform

import (
	"encoding/csv"
	"fmt"
	"os/exec"
	"strconv"
	"strings"

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
	netstatOut, err := exec.Command("netstat", "-ano").Output()
	if err != nil {
		return nil, fmt.Errorf("netstat failed: %w", err)
	}

	ports, err := parseNetstatOutput(string(netstatOut))
	if err != nil {
		return nil, err
	}

	if len(ports) > 0 {
		tasklistOut, err := exec.Command("tasklist", "/FO", "CSV").Output()
		if err == nil {
			pidToName := parseTasklistOutput(string(tasklistOut))
			for i := range ports {
				if name, ok := pidToName[ports[i].PID]; ok {
					ports[i].Process = name
				}
			}
		}
	}

	return ports, nil
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
	return nil, fmt.Errorf("no process found on port %d", port)
}

// parseNetstatOutput parses the output of netstat -ano.
func parseNetstatOutput(output string) ([]models.PortProcess, error) {
	var results []models.PortProcess
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "Proto") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}

		proto := strings.ToLower(fields[0])
		localAddr := fields[1]
		var state, pidStr string

		switch proto {
		case "tcp":
			if len(fields) < 5 {
				continue
			}
			state = fields[3]
			pidStr = fields[4]
		case "udp":
			pidStr = fields[3]
		default:
			continue
		}

		if proto == "tcp" && state != "LISTENING" {
			continue
		}

		port := extractPortFromNetstatAddr(localAddr)
		if port == 0 {
			continue
		}

		pid, _ := strconv.Atoi(pidStr)

		results = append(results, models.PortProcess{
			Port:     port,
			Protocol: proto,
			PID:      pid,
		})
	}

	return results, nil
}

// extractPortFromNetstatAddr extracts the port from a netstat address.
func extractPortFromNetstatAddr(addr string) int {
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

// parseTasklistOutput parses CSV output from tasklist /FO CSV.
func parseTasklistOutput(output string) map[int]string {
	result := make(map[int]string)
	reader := csv.NewReader(strings.NewReader(output))
	records, err := reader.ReadAll()
	if err != nil {
		return result
	}

	for i, record := range records {
		if i == 0 {
			continue // skip header
		}
		if len(record) < 2 {
			continue
		}
		name := strings.Trim(record[0], `"`)
		pidStr := strings.Trim(record[1], `"`)
		pid, err := strconv.Atoi(pidStr)
		if err != nil {
			continue
		}
		result[pid] = name
	}

	return result
}
