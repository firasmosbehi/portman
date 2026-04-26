package platform

import "errors"

// ErrProcessNotFound is returned when no process is using the requested port.
var ErrProcessNotFound = errors.New("no process found on port")
