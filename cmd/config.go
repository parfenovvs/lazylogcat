package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/parfenovvs/lazylogcat/internal/app"
	"github.com/parfenovvs/lazylogcat/internal/config"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Inspect lazylogcat configuration",
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Print merged configuration as JSON",
	Long: `Loads and merges configuration from defaults, user config, project config,
and local override (same rules as the TUI). Warnings about unreadable files go to stderr.`,
	Example: `  lazylogcat config show`,
	SilenceUsage: true,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return app.SetupLogging(debugFlag)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := config.Resolve()
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: %v\n", err)
		}
		data, err := json.MarshalIndent(c, "", "  ")
		if err != nil {
			return fmt.Errorf("encode config: %w", err)
		}
		fmt.Println(string(data))
		return nil
	},
	PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
		return app.CloseLog()
	},
}

func init() {
	configCmd.AddCommand(configShowCmd)
	rootCmd.AddCommand(configCmd)
}
