package platform

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// parsePsEtime parses the output of `ps -p <pid> -o etime=`.
// Formats: DD-HH:MM:SS, HH:MM:SS, MM:SS, or SS.
func parsePsEtime(output string) (time.Duration, error) {
	output = strings.TrimSpace(output)
	if output == "" {
		return 0, fmt.Errorf("empty etime output")
	}

	// Handle DD-HH:MM:SS
	parts := strings.SplitN(output, "-", 2)
	var days int
	var timePart string
	if len(parts) == 2 {
		days, _ = strconv.Atoi(parts[0])
		timePart = parts[1]
	} else {
		timePart = parts[0]
	}

	// Parse HH:MM:SS, MM:SS, or SS
	timeFields := strings.Split(timePart, ":")
	var hours, minutes, seconds int
	switch len(timeFields) {
	case 3:
		hours, _ = strconv.Atoi(timeFields[0])
		minutes, _ = strconv.Atoi(timeFields[1])
		seconds, _ = strconv.Atoi(timeFields[2])
	case 2:
		minutes, _ = strconv.Atoi(timeFields[0])
		seconds, _ = strconv.Atoi(timeFields[1])
	case 1:
		seconds, _ = strconv.Atoi(timeFields[0])
	default:
		return 0, fmt.Errorf("unexpected etime format: %s", output)
	}

	d := time.Duration(days)*24*time.Hour +
		time.Duration(hours)*time.Hour +
		time.Duration(minutes)*time.Minute +
		time.Duration(seconds)*time.Second
	return d, nil
}

// parseWMICreationDate parses WMI CreationDate format: YYYYMMDDHHMMSS.mmmmmmsUUU
func parseWMICreationDate(output string) (time.Duration, error) {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.EqualFold(line, "CreationDate") {
			continue
		}
		// Parse YYYYMMDDHHMMSS.mmmmmmsUUU
		if len(line) < 14 {
			continue
		}
		year, _ := strconv.Atoi(line[0:4])
		month, _ := strconv.Atoi(line[4:6])
		day, _ := strconv.Atoi(line[6:8])
		hour, _ := strconv.Atoi(line[8:10])
		minute, _ := strconv.Atoi(line[10:12])
		second, _ := strconv.Atoi(line[12:14])

		created := time.Date(year, time.Month(month), day, hour, minute, second, 0, time.Local)
		return time.Since(created), nil
	}
	return 0, fmt.Errorf("could not parse creation date from wmic output")
}
