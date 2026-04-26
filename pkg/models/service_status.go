package models

// ServiceStatus represents the runtime status of a declared service.
type ServiceStatus struct {
	Name     string
	Expected int
	Actual   string
	Status   string
	Healthy  bool
}
