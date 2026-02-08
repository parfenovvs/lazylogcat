package util

import (
	"bufio"
	"fmt"
	"io"
	"log/slog"
	"os/exec"
	"strconv"
	"strings"

	"github.com/parfenovvs/lazylogcat/internal/model"
)

const initialLogLinesCount = 1000

var (
	ErrFailedToGetDevices     = fmt.Errorf("failed to get connected devices")
	ErrFailedToStartLogcat    = fmt.Errorf("failed to start logcat process")
	ErrFailedToGetStdoutPipe  = fmt.Errorf("failed to get stdout pipe")
	ErrLogcatConnectionClosed = fmt.Errorf("logcat connection is closed")
	ErrReadingLogcat          = fmt.Errorf("error reading logcat stream")
)

var (
	logcatCmd     *exec.Cmd
	logcatScanner *bufio.Scanner
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

func ConnectLogcat(deviceId string, filter model.Filter, format model.Format) error {
	args := []string{"-s", deviceId, "logcat", "-T", strconv.Itoa(initialLogLinesCount), "-v", "threadtime"}

	slog.Debug("Executing adb logcat command", "args", args)

	cmd := exec.Command("adb", args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("%w: %w", ErrFailedToGetStdoutPipe, err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("%w: %w", ErrFailedToStartLogcat, err)
	}

	logcatCmd = cmd
	logcatScanner = bufio.NewScanner(stdout)

	return nil
}

func ReadNextLogLine() (string, error) {
	if logcatScanner == nil {
		return "", ErrLogcatConnectionClosed
	}

	if logcatScanner.Scan() {
		return logcatScanner.Text(), nil
	}

	if err := logcatScanner.Err(); err != nil {
		return "", fmt.Errorf("%w: %w", ErrReadingLogcat, err)
	}

	return "", io.EOF
}

func CloseLogcat() error {
	if logcatCmd != nil && logcatCmd.Process != nil {
		slog.Debug("Killing adb logcat process")
		logcatCmd.Process.Kill()
		logcatCmd.Wait()
		logcatCmd = nil
	}
	logcatScanner = nil
	return nil
}
