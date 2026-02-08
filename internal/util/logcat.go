package util

import (
	"os/exec"
	"strings"
)

// Process holds a PID and its associated package name from `adb shell ps`.
type Process struct {
	PID  string
	Name string
}

// GetProcessList runs `adb shell ps` and returns all processes with their PIDs and names.
func GetProcessList(deviceId string) ([]Process, error) {
	cmd := exec.Command("adb", "-s", deviceId, "shell", "ps", "-A", "-o", "PID,NAME")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return ParseProcessList(string(output)), nil
}

// ParseProcessList parses the output of `adb shell ps -A -o PID,NAME` into a slice of Process.
// The first line is assumed to be a header and is skipped.
func ParseProcessList(output string) []Process {
	lines := strings.Split(strings.TrimRight(output, "\n\r "), "\n")
	if len(lines) <= 1 {
		return nil
	}

	var processes []Process
	for _, line := range lines[1:] {
		fields := strings.Fields(strings.TrimRight(line, "\r"))
		if len(fields) < 2 {
			continue
		}
		processes = append(processes, Process{
			PID:  fields[0],
			Name: fields[1],
		})
	}
	return processes
}

// ResolvePIDs returns the set of PIDs whose package name contains the filter
// string (case-insensitive). Returns nil if filter is empty.
func ResolvePIDs(processes []Process, filter string) map[string]struct{} {
	if filter == "" {
		return nil
	}

	filterLower := strings.ToLower(filter)
	pidSet := make(map[string]struct{})
	for _, p := range processes {
		if strings.Contains(strings.ToLower(p.Name), filterLower) {
			pidSet[p.PID] = struct{}{}
		}
	}
	return pidSet
}
