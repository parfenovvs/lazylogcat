package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/parfenovvs/lazylogcat/internal/app"
	"github.com/parfenovvs/lazylogcat/internal/util"
	"github.com/spf13/cobra"
)

var devicesOutput string

var devicesCmd = &cobra.Command{
	Use:   "devices",
	Short: "List connected Android devices",
	Long:  `Runs adb devices -l and prints connected devices. Default output is JSON for scripting.`,
	Example: `  lazylogcat devices
  lazylogcat devices --output text`,
	SilenceUsage: true,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return app.SetupLogging(debugFlag)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := app.PreLaunchChecks(); err != nil {
			return err
		}
		devices, err := util.GetConnectedDevices()
		if err != nil {
			return err
		}
		switch devicesOutput {
		case "json":
			data, err := json.MarshalIndent(devices, "", "  ")
			if err != nil {
				return fmt.Errorf("encode devices: %w", err)
			}
			fmt.Println(string(data))
		case "text":
			for _, d := range devices {
				fmt.Fprintf(os.Stdout, "%s\t%s\n", d.Id, d.Name)
			}
		default:
			return fmt.Errorf("invalid --output %q (use json or text)", devicesOutput)
		}
		return nil
	},
	PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
		return app.CloseLog()
	},
}

func init() {
	devicesCmd.Flags().StringVar(&devicesOutput, "output", "json", "Output format: json or text")
	rootCmd.AddCommand(devicesCmd)
}
