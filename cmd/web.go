package cmd

import (
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/pkg/browser"
	"github.com/spf13/cobra"

	"github.com/parfenovvs/lazylogcat/internal/app"
	"github.com/parfenovvs/lazylogcat/internal/config"
	"github.com/parfenovvs/lazylogcat/internal/web"
)

var (
	portFlag int
	demoFlag bool
)

var webCmd = &cobra.Command{
	Use:          "web",
	Short:        "Start browser-based logcat viewer",
	Long:         `Start an HTTP server that serves a browser-based logcat viewer with real-time streaming over WebSocket.`,
	SilenceUsage: true,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return app.SetupLogging(debugFlag)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Fprintln(os.Stderr, "WARNING: The web experience is experimental and may change or break without notice.")

		// Skip adb check in demo mode
		if !demoFlag {
			if err := app.PreLaunchChecks(); err != nil {
				return err
			}
		}

		c, err := config.Resolve(nil)
		if err != nil {
			slog.Warn("Config resolution had errors", "error", err)
		}

		if cmd.Flags().Changed("pkg") {
			c.Filter.Pkg = config.TextFilter{Value: pkgFlag}
		}
		if cmd.Flags().Changed("tag") {
			c.Filter.Tag = config.TextFilter{Value: tagFlag}
		}
		if cmd.Flags().Changed("text") {
			c.Filter.Txt = config.TextFilter{Value: textFlag}
		}

		slog.Debug("Configuration loaded", "config", c.String(), "demo", demoFlag)

		srv, err := web.NewServer(portFlag, c, demoFlag)
		if err != nil {
			return fmt.Errorf("failed to create server: %w", err)
		}

		// Open browser (non-blocking, best-effort)
		url := fmt.Sprintf("http://localhost:%d", portFlag)
		if err := browser.OpenURL(url); err != nil {
			slog.Warn("Failed to open browser", "error", err)
			fmt.Fprintf(os.Stderr, "Open %s in your browser\n", url)
		}

		// Start server in a goroutine so we can wait for shutdown signal
		errCh := make(chan error, 1)
		go func() {
			errCh <- srv.Start()
		}()

		// Wait for SIGINT/SIGTERM or server error
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

		select {
		case sig := <-sigCh:
			slog.Info("Received signal", "signal", sig)
			if err := srv.Shutdown(5 * time.Second); err != nil {
				return fmt.Errorf("shutdown error: %w", err)
			}
		case err := <-errCh:
			if err != nil {
				return err
			}
		}

		return nil
	},
}

func init() {
	webCmd.Flags().IntVar(&portFlag, "port", 8321, "Port to listen on")
	webCmd.Flags().BoolVar(&demoFlag, "demo", false, "Run with a fake device and synthetic log lines (no adb required)")
	webCmd.Flags().StringVar(&pkgFlag, "pkg", "", "Filter by package name (contains match, overrides config)")
	webCmd.Flags().StringVar(&tagFlag, "tag", "", "Filter by log tag (contains match, overrides config)")
	webCmd.Flags().StringVar(&textFlag, "text", "", "Filter by log text (contains match, overrides config)")
	rootCmd.AddCommand(webCmd)
}
