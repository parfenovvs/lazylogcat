package cmd

import (
	"fmt"
	"runtime/debug"

	"github.com/spf13/cobra"
)

var (
	Version = ""
)

func getVersion() string {
	// If version was set via ldflags
	if Version != "" {
		return Version
	}

	// Otherwise, try to get version from build info (when installed via go install)
	if info, ok := debug.ReadBuildInfo(); ok {
		if info.Main.Version != "" && info.Main.Version != "(devel)" {
			return info.Main.Version
		}
	}

	return "dev"
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Example: `  lazylogcat version`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(getVersion())
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
