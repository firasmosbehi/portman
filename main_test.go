package main

import "testing"

func TestVersionVarsExist(t *testing.T) {
	// Ensure the ldflags variables are declared and accessible.
	_ = version
	_ = commit
	_ = date
}
