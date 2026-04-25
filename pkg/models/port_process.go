package models

import "time"

// PortProcess represents a listening port and its owning process.
type PortProcess struct {
	Port     int
	Protocol string
	Process  string
	PID      int
	User     string
	Age      time.Duration
}
