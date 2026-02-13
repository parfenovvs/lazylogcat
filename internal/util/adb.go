package util

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/parfenovvs/lazylogcat/internal/model"
)

var (
	ErrFailedToGetDevices    = fmt.Errorf("failed to get connected devices")
	ErrFailedToStartLogcat   = fmt.Errorf("failed to start logcat process")
	ErrFailedToGetStdoutPipe = fmt.Errorf("failed to get stdout pipe")
)

func GetConnectedDevices() ([]model.Device, error) {
	cmd := exec.Command("adb", "devices", "-l")

	output, err := cmd.Output()
	if err != nil {
		return nil, ErrFailedToGetDevices
	}

	devices := make([]model.Device, 0)

	strOutput := string(output)
	lines := strings.Split(strOutput, "\n")
	for _, l := range lines[1:] {
		if strings.Contains(l, "device") {
			parts := strings.Fields(l)
			if len(parts) == 0 {
				continue
			}
			id := parts[0]
			name := "Undefined"
			for _, p := range parts {
				if strings.HasPrefix(p, "model:") {
					product := strings.Split(p, ":")
					name = product[len(product)-1]
					break
				}
			}
			devices = append(devices, model.Device{Id: id, Name: name})
		}
	}

	return devices, nil
}
