package cmd

import (
	"encoding/json"
	"fmt"
	"os"
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

var versionJSON bool

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Example: `  lazylogcat version
  lazylogcat version --json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if versionJSON {
			b, err := json.Marshal(struct {
				Version string `json:"version"`
			}{Version: getVersion()})
			if err != nil {
				return err
			}
			fmt.Fprintln(os.Stdout, string(b))
			return nil
		}
		fmt.Println(getVersion())
		return nil
	},
}

func init() {
	versionCmd.Flags().BoolVar(&versionJSON, "json", false, "Print version as a JSON object")
	rootCmd.AddCommand(versionCmd)
}
