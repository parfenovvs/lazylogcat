package util

import (
	"fmt"
	"os/exec"
	"strings"
)

func GetPidByPackageName(deviceId string, pkg string) (string, error) {
	pidCmd := exec.Command("adb", "-s", deviceId, "shell", "pidof", pkg)
	pid, err := pidCmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get pid: %w", err)
	}
	return strings.Trim(string(pid), "\n\r "), nil
}
