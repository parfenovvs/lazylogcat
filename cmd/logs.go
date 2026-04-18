package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/parfenovvs/lazylogcat/internal/app"
	"github.com/parfenovvs/lazylogcat/internal/config"
	"github.com/parfenovvs/lazylogcat/internal/model"
	"github.com/parfenovvs/lazylogcat/internal/util"
	"github.com/spf13/cobra"
)

var (
	logsDevice string
	logsLines  int
	logsFollow bool
	logsFormat string
)

var logsCmd = &cobra.Command{
	Use:   "logs",
	Short: "Stream or parse Android logcat output",
}

var logsDumpCmd = &cobra.Command{
	Use:   "dump",
	Short: "Stream logcat from a device (filtered)",
	Long: `Runs adb logcat for the selected device and prints matching lines to stdout.
Use --follow to stream until interrupted, or omit it to collect up to --lines lines then exit.`,
	Example: `  lazylogcat logs dump --lines 200 --format text
  lazylogcat logs dump --device emulator-5554 --pkg com.example.app --follow`,
	SilenceUsage: true,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return app.SetupLogging(debugFlag)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return runLogsDump(cmd, args)
	},
}

var logsParseCmd = &cobra.Command{
	Use:   "parse",
	Short: "Parse logcat lines from stdin",
	Long: `Reads raw logcat lines from stdin, applies the same filters as dump, and prints
matching lines. Use for pipelines: adb logcat -d | lazylogcat logs parse --text error

Package filtering requires adb and --device when the package filter is non-empty.`,
	Example: `  adb logcat -d -v threadtime | lazylogcat logs parse --format jsonl
  lazylogcat logs parse --tag System --format text < saved.log`,
	SilenceUsage: true,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return app.SetupLogging(debugFlag)
	},
	RunE: runLogsParse,
}

func runLogsDump(cmd *cobra.Command, args []string) error {
	if err := app.PreLaunchChecks(); err != nil {
		return err
	}
	if err := validateLogsFormat(logsFormat); err != nil {
		return err
	}
	c, err := resolveAppConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: %v\n", err)
	}
	mergeLogFilters(cmd, &c)
	f := util.FilterFromConfig(&c)
	f.PackageName.Compile()
	f.Tag.Compile()
	f.Text.Compile()

	deviceID, err := resolveLogsDeviceID(logsDevice)
	if err != nil {
		return err
	}

	var followMax int
	if logsFollow {
		if cmd.Flags().Changed("lines") {
			followMax = logsLines
		}
	} else if logsLines <= 0 {
		return fmt.Errorf("--lines must be positive when --follow is false")
	}

	r := util.NewLogcatReader()
	if err := r.Connect(deviceID, f); err != nil {
		return err
	}
	util.UpdatePIDSetForPackage(r, deviceID, f.PackageName)

	if logsFollow {
		return runLogsDumpFollow(r, logsFormat, followMax)
	}
	return runLogsDumpBatch(r, logsFormat, logsLines)
}

func runLogsDumpFollow(r *util.LogcatReader, format string, maxLines int) error {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	printed := 0
	for {
		select {
		case <-sigCh:
			r.Disconnect()
			return nil
		case <-ticker.C:
			for _, line := range r.Drain() {
				if err := printLogLine(line, format); err != nil {
					r.Disconnect()
					return err
				}
				printed++
				if maxLines > 0 && printed >= maxLines {
					r.Disconnect()
					return nil
				}
			}
			if !r.IsConnected() {
				if err := r.Err(); err != nil {
					return err
				}
				return nil
			}
		}
	}
}

func runLogsDumpBatch(r *util.LogcatReader, format string, wantLines int) error {
	defer r.Disconnect()

	printed := 0
	for printed < wantLines {
		for _, line := range r.Drain() {
			if err := printLogLine(line, format); err != nil {
				return err
			}
			printed++
			if printed >= wantLines {
				return nil
			}
		}
		if !r.IsConnected() {
			return r.Err()
		}
		time.Sleep(50 * time.Millisecond)
	}
	return nil
}

func runLogsParse(cmd *cobra.Command, args []string) error {
	if err := validateLogsFormat(logsFormat); err != nil {
		return err
	}
	c, err := resolveAppConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: %v\n", err)
	}
	mergeLogFilters(cmd, &c)
	f := util.FilterFromConfig(&c)
	f.PackageName.Compile()
	f.Tag.Compile()
	f.Text.Compile()

	var pidSet map[string]struct{}
	if !f.PackageName.IsEmpty() {
		if err := app.PreLaunchChecks(); err != nil {
			return err
		}
		deviceID, err := resolveLogsDeviceID(logsDevice)
		if err != nil {
			return err
		}
		processes, err := util.GetProcessList(deviceID)
		if err != nil {
			return fmt.Errorf("process list for package filter: %w", err)
		}
		pidSet = util.ResolvePIDs(processes, &f.PackageName)
	}

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Buffer(make([]byte, 0, 256*1024), 1024*1024)
	for scanner.Scan() {
		raw := scanner.Text()
		line := model.ParseLogLine(raw)
		if !util.MatchesFilter(raw, line, &f, pidSet) {
			continue
		}
		if err := printLogLine(line, logsFormat); err != nil {
			return err
		}
	}
	return scanner.Err()
}

func mergeLogFilters(cmd *cobra.Command, c *config.Config) {
	if cmd.Flags().Changed("pkg") {
		c.Filter.Pkg = config.TextFilter{Value: pkgFlag}
	}
	if cmd.Flags().Changed("tag") {
		c.Filter.Tag = config.TextFilter{Value: tagFlag}
	}
	if cmd.Flags().Changed("text") {
		c.Filter.Txt = config.TextFilter{Value: textFlag}
	}
}

func resolveLogsDeviceID(flag string) (string, error) {
	devices, err := util.GetConnectedDevices()
	if err != nil {
		return "", err
	}
	if flag != "" {
		for _, d := range devices {
			if d.Id == flag {
				return d.Id, nil
			}
		}
		return "", fmt.Errorf("device %q is not among connected devices", flag)
	}
	switch len(devices) {
	case 0:
		return "", fmt.Errorf("no devices connected; connect one or pass --device")
	case 1:
		return devices[0].Id, nil
	default:
		return "", fmt.Errorf("multiple devices connected; specify --device (example: %s)", devices[0].Id)
	}
}

func validateLogsFormat(s string) error {
	switch s {
	case "jsonl", "text":
		return nil
	default:
		return fmt.Errorf("invalid --format %q (use jsonl or text)", s)
	}
}

func printLogLine(line model.LogLine, format string) error {
	switch format {
	case "jsonl":
		b, err := json.Marshal(line)
		if err != nil {
			return err
		}
		fmt.Println(string(b))
	case "text":
		fmt.Println(line.Raw)
	}
	return nil
}

func init() {
	logsDumpCmd.Flags().StringVar(&logsDevice, "device", "", "ADB device id (default: only connected device, or required when several)")
	logsDumpCmd.Flags().StringVar(&pkgFlag, "pkg", "", "Filter by package name (contains match, overrides config)")
	logsDumpCmd.Flags().StringVar(&tagFlag, "tag", "", "Filter by log tag (contains match, overrides config)")
	logsDumpCmd.Flags().StringVar(&textFlag, "text", "", "Filter by log text (contains match, overrides config)")
	logsDumpCmd.Flags().IntVar(&logsLines, "lines", 500, "With --follow: optional max lines (default: unlimited). Without --follow: stop after this many lines")
	logsDumpCmd.Flags().BoolVar(&logsFollow, "follow", false, "Stream until SIGINT; use --lines to cap output")
	logsDumpCmd.Flags().StringVar(&logsFormat, "format", "jsonl", "Output format: jsonl or text")

	logsParseCmd.Flags().StringVar(&logsDevice, "device", "", "ADB device id (required for --pkg when resolving process PIDs)")
	logsParseCmd.Flags().StringVar(&pkgFlag, "pkg", "", "Filter by package name (contains match, overrides config)")
	logsParseCmd.Flags().StringVar(&tagFlag, "tag", "", "Filter by log tag (contains match, overrides config)")
	logsParseCmd.Flags().StringVar(&textFlag, "text", "", "Filter by log text (contains match, overrides config)")
	logsParseCmd.Flags().StringVar(&logsFormat, "format", "jsonl", "Output format: jsonl or text")

	logsCmd.AddCommand(logsDumpCmd, logsParseCmd)
	rootCmd.AddCommand(logsCmd)
}
