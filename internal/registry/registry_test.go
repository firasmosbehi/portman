package registry

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadValid(t *testing.T) {
	content := `
services:
  - name: web
    port: 3000
    command: npm run dev

  - name: api
    port: 3001
    command: npm run api

  - name: db
    port: 5432
    health_check: pg_isready

  - name: cache
    port: 6379
`
	dir := t.TempDir()
	path := filepath.Join(dir, "portman.yml")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	reg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(reg.Services) != 4 {
		t.Fatalf("expected 4 services, got %d", len(reg.Services))
	}

	expected := []Service{
		{Name: "web", Port: 3000, Command: "npm run dev"},
		{Name: "api", Port: 3001, Command: "npm run api"},
		{Name: "db", Port: 5432, HealthCheck: "pg_isready"},
		{Name: "cache", Port: 6379},
	}

	for i, exp := range expected {
		if reg.Services[i] != exp {
			t.Errorf("service[%d] = %+v, want %+v", i, reg.Services[i], exp)
		}
	}
}

func TestLoadFileNotFound(t *testing.T) {
	_, err := Load("/nonexistent/portman.yml")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestLoadInvalidYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "portman.yml")
	if err := os.WriteFile(path, []byte("not: valid: yaml: ["), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestLoadMissingName(t *testing.T) {
	content := `services:
  - port: 3000
`
	dir := t.TempDir()
	path := filepath.Join(dir, "portman.yml")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for missing name")
	}
}

func TestLoadMissingPort(t *testing.T) {
	content := `services:
  - name: web
`
	dir := t.TempDir()
	path := filepath.Join(dir, "portman.yml")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for missing port")
	}
}

func TestFindInCurrentDir(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "portman.yml")
	if err := os.WriteFile(path, []byte("services:\n  - name: web\n    port: 3000\n"), 0644); err != nil {
		t.Fatal(err)
	}

	origDir, _ := os.Getwd()
	_ = os.Chdir(dir)
	defer func() { _ = os.Chdir(origDir) }()

	found, err := Find()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found == "" {
		t.Fatal("expected to find portman.yml")
	}
}

func TestFindInParentDir(t *testing.T) {
	parent := t.TempDir()
	child := filepath.Join(parent, "subdir")
	if err := os.MkdirAll(child, 0755); err != nil {
		t.Fatal(err)
	}

	path := filepath.Join(parent, "portman.yml")
	if err := os.WriteFile(path, []byte("services:\n  - name: web\n    port: 3000\n"), 0644); err != nil {
		t.Fatal(err)
	}

	origDir, _ := os.Getwd()
	_ = os.Chdir(child)
	defer func() { _ = os.Chdir(origDir) }()

	found, err := Find()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found == "" {
		t.Fatal("expected to find portman.yml in parent dir")
	}
}

func TestFindNotFound(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	_ = os.Chdir(dir)
	defer func() { _ = os.Chdir(origDir) }()

	found, err := Find()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found != "" {
		t.Fatalf("expected empty path, got %q", found)
	}
}
