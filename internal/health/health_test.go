package health

import (
	"net"
	"testing"
	"time"
)

func TestTCPCheckSuccess(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()

	port := ln.Addr().(*net.TCPAddr).Port
	c := NewChecker()
	if !c.TCPCheck(port) {
		t.Errorf("expected TCPCheck to succeed for port %d", port)
	}
}

func TestTCPCheckFailure(t *testing.T) {
	c := NewChecker()
	if c.TCPCheck(1) {
		t.Error("expected TCPCheck to fail for port 1")
	}
}

func TestCommandCheckSuccess(t *testing.T) {
	c := NewChecker()
	if !c.CommandCheck("echo hello") {
		t.Error("expected CommandCheck to succeed for echo")
	}
}

func TestCommandCheckFailure(t *testing.T) {
	c := NewChecker()
	if c.CommandCheck("exit 1") {
		t.Error("expected CommandCheck to fail for exit 1")
	}
}

func TestCommandCheckTimeout(t *testing.T) {
	c := NewChecker()
	start := time.Now()
	result := c.CommandCheck("sleep 10")
	elapsed := time.Since(start)
	if result {
		t.Error("expected CommandCheck to fail on timeout")
	}
	if elapsed > 6*time.Second {
		t.Errorf("expected timeout ~5s, took %v", elapsed)
	}
}
