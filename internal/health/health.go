package health

import (
	"context"
	"fmt"
	"net"
	"os/exec"
	"time"
)

// Checker executes health checks for services.
type Checker struct{}

// NewChecker creates a new health checker.
func NewChecker() *Checker {
	return &Checker{}
}

// TCPCheck returns true if the given port is open locally.
func (c *Checker) TCPCheck(port int) bool {
	addr := fmt.Sprintf("127.0.0.1:%d", port)
	conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}

// CommandCheck runs the given command and returns true if it exits 0.
func (c *Checker) CommandCheck(command string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "sh", "-c", command)
	if err := cmd.Run(); err != nil {
		return false
	}
	return true
}
