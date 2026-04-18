package cmd

import (
	"log/slog"

	"github.com/parfenovvs/lazylogcat/internal/app"
	"github.com/parfenovvs/lazylogcat/internal/config"
	"github.com/spf13/cobra"
)

var debugFlag bool

var (
	pkgFlag  string
	tagFlag  string
	textFlag string

	cliProjectConfig string
	cliLocalConfig   string
)

var rootCmd = &cobra.Command{
	Use:          "lazylogcat",
	Short:        "Interactive Android logcat viewer",
	Long:         `lazylogcat is an interactive TUI application for viewing Android device logs.`,
	Example: `  lazylogcat
  lazylogcat --pkg com.example.app --tag MainActivity
  lazylogcat --text error --debug`,
	SilenceUsage: true,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return app.SetupLogging(debugFlag)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := app.PreLaunchChecks(); err != nil {
			return err
		}

		c, err := resolveAppConfig()
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

		slog.Debug("Configuration loaded", "config", c.String())
		return app.LaunchTUI(c)
	},
	PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
		return app.CloseLog()
	},
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&debugFlag, "debug", false, "Enable debug logging to .lazylogcat.log file")
	rootCmd.PersistentFlags().StringVar(&cliProjectConfig, "config", "", "Override path for project config (default: .lazylogcat/config.json)")
	rootCmd.PersistentFlags().StringVar(&cliLocalConfig, "config-local", "", "Override path for local config (default: .lazylogcat/config.local.json)")
	rootCmd.Flags().StringVar(&pkgFlag, "pkg", "", "Filter by package name (contains match, overrides config)")
	rootCmd.Flags().StringVar(&tagFlag, "tag", "", "Filter by log tag (contains match, overrides config)")
	rootCmd.Flags().StringVar(&textFlag, "text", "", "Filter by log text (contains match, overrides config)")
}

func configOpts() *config.ResolveOptions {
	if cliProjectConfig == "" && cliLocalConfig == "" {
		return nil
	}
	return &config.ResolveOptions{
		ProjectConfigPath: cliProjectConfig,
		LocalConfigPath:   cliLocalConfig,
	}
}

func resolveAppConfig() (config.Config, error) {
	return config.Resolve(configOpts())
}

func Execute() error {
	return rootCmd.Execute()
}
