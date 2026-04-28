//go:build windows

package platform

import (
	"encoding/csv"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/firasmosbehi/portman/pkg/models"
)

// netstatRunner abstracts the netstat command for testability.
var netstatRunner = func() ([]byte, error) {
	return exec.Command("netstat", "-ano").Output()
}

// tasklistRunner abstracts the tasklist command for testability.
var tasklistRunner = func() ([]byte, error) {
	return exec.Command("tasklist", "/FO", "CSV").Output()
}

// wmicRunner abstracts the wmic command for testability.
var wmicRunner = func(pid int) ([]byte, error) {
	return exec.Command("wmic", "process", "where", fmt.Sprintf("ProcessId=%d", pid), "get", "CreationDate").Output()
}

func getProcessAge(pid int) (time.Duration, error) {
	out, err := wmicRunner(pid)
	if err != nil {
		return 0, err
	}
	return parseWMICreationDate(string(out))
}

// Resolver implements port and process resolution for Windows.
type Resolver struct{}

// NewResolver creates a new Windows platform resolver.
func NewResolver() *Resolver {
	return &Resolver{}
}

// GetListeningPorts returns all listening ports with process info.
func (r *Resolver) GetListeningPorts() ([]models.PortProcess, error) {
	netstatOut, err := netstatRunner()
	if err != nil {
		return nil, fmt.Errorf("netstat failed: %w", err)
	}

	ports, err := parseNetstatOutput(string(netstatOut))
	if err != nil {
		return nil, err
	}

	if len(ports) > 0 {
		tasklistOut, err := tasklistRunner()
		if err == nil {
			pidToName := parseTasklistOutput(string(tasklistOut))
			for i := range ports {
				if name, ok := pidToName[ports[i].PID]; ok {
					ports[i].Process = name
				}
			}
		}
	}

	// Populate age for each unique PID.
	for i := range ports {
		if ports[i].PID > 0 {
			if age, err := getProcessAge(ports[i].PID); err == nil {
				ports[i].Age = age
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
	return nil, fmt.Errorf("%w %d", ErrProcessNotFound, port)
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
