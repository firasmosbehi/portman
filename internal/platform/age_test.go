package platform

import (
	"testing"
	"time"
)

func TestParsePsEtime(t *testing.T) {
	tests := []struct {
		input    string
		expected time.Duration
	}{
		{"45:12", 45*time.Minute + 12*time.Second},
		{"03:45:12", 3*time.Hour + 45*time.Minute + 12*time.Second},
		{"2-03:45:12", 2*24*time.Hour + 3*time.Hour + 45*time.Minute + 12*time.Second},
		{"30", 30 * time.Second},
		{"  15:30  ", 15*time.Minute + 30*time.Second},
	}

	for _, tt := range tests {
		got, err := parsePsEtime(tt.input)
		if err != nil {
			t.Fatalf("parsePsEtime(%q) unexpected error: %v", tt.input, err)
		}
		if got != tt.expected {
			t.Errorf("parsePsEtime(%q) = %v, want %v", tt.input, got, tt.expected)
		}
	}
}

func TestParsePsEtimeEmpty(t *testing.T) {
	_, err := parsePsEtime("")
	if err == nil {
		t.Fatal("expected error for empty input")
	}
}

func TestParsePsEtimeInvalid(t *testing.T) {
	_, err := parsePsEtime("a:b:c:d:e")
	if err == nil {
		t.Fatal("expected error for invalid format")
	}
}

func TestParseWMICreationDate(t *testing.T) {
	// WMI datetime format: YYYYMMDDHHMMSS.mmmmmmsUUU
	output := "CreationDate  \n20240115123045.000000+060  \n"
	got, err := parseWMICreationDate(output)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// We can't assert exact duration since it depends on current time,
	// but we can assert it's positive.
	if got <= 0 {
		t.Errorf("expected positive duration, got %v", got)
	}
}

func TestParseWMICreationDateInvalid(t *testing.T) {
	_, err := parseWMICreationDate("CreationDate\n\n")
	if err == nil {
		t.Fatal("expected error for invalid wmic output")
	}
}
