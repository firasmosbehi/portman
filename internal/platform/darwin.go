//go:build darwin

package platform

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"strconv"
	"strings"

	"github.com/firasmosbehi/portman/pkg/models"
)

// lsofRunner abstracts the lsof command for testability.
var lsofRunner = func() ([]byte, error) {
	return exec.Command("lsof", "-i", "-P", "-n", "-F").Output()
}

// Resolver implements port and process resolution for macOS.
type Resolver struct{}

// NewResolver creates a new macOS platform resolver.
func NewResolver() *Resolver {
	return &Resolver{}
}

// GetListeningPorts returns all listening ports with process info.
func (r *Resolver) GetListeningPorts() ([]models.PortProcess, error) {
	out, err := lsofRunner()
	if err != nil {
		// lsof often exits 1 with partial output; if we have data, parse it.
		if len(out) == 0 {
			return nil, fmt.Errorf("lsof failed: %w", err)
		}
	}
	return parseLsofOutput(bytes.NewReader(out))
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

// fileEntry holds the parsed fields for a single open file within a process.
type fileEntry struct {
	protocol  string
	address   string
	tcpListen bool
}

// flushPending evaluates the current file entry and appends it to results if it
// represents a listening port.
func flushPending(results *[]models.PortProcess, pending *fileEntry, pid int, process, user string) {
	if pending.address == "" {
		return
	}
	port := extractPortFromLsofAddr(pending.address)
	if port == 0 {
		return
	}
	// Include UDP (no state) or TCP with LISTEN state.
	if pending.protocol == "udp" || pending.tcpListen {
		*results = append(*results, models.PortProcess{
			Port:     port,
			Protocol: pending.protocol,
			Process:  process,
			PID:      pid,
			User:     user,
		})
	}
}

// parseLsofOutput parses the machine-readable output of lsof -i -P -n -F.
func parseLsofOutput(r io.Reader) ([]models.PortProcess, error) {
	scanner := bufio.NewScanner(r)
	var results []models.PortProcess

	var pid int
	var process, user string
	var pending fileEntry

	for scanner.Scan() {
		line := scanner.Text()
		if len(line) < 2 {
			continue
		}

		field := line[0]
		value := line[1:]

		switch field {
		case 'p':
			flushPending(&results, &pending, pid, process, user)
			pending = fileEntry{}
			pid, _ = strconv.Atoi(value)
		case 'c':
			process = value
		case 'L':
			user = value
		case 'f':
			flushPending(&results, &pending, pid, process, user)
			pending = fileEntry{}
		case 'P':
			pending.protocol = strings.ToLower(value)
		case 'T':
			if strings.Contains(value, "ST=LISTEN") {
				pending.tcpListen = true
			}
		case 'n':
			if !strings.Contains(value, "->") {
				pending.address = value
			}
		}
	}

	flushPending(&results, &pending, pid, process, user)
	return results, scanner.Err()
}

// extractPortFromLsofAddr extracts the port number from an lsof address.
// Handles formats like *:3000, 0.0.0.0:3000, [::]:3000, 127.0.0.1:3000.
func extractPortFromLsofAddr(addr string) int {
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
