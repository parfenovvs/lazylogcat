package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/parfenovvs/lazylogcat/internal/skillinstall"
	"github.com/parfenovvs/lazylogcat/skills"
)

var (
	skillAgent   string
	skillUser    bool
	skillProject bool
)

var skillCmd = &cobra.Command{
	Use:   "skill",
	Short: "Manage bundled agent skills",
}

var skillInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Install the embedded lazylogcat skill for an AI agent",
	Long: `Copies the bundled skill from this binary into the agent-specific skills directory.

You must pass --agent and exactly one of --user or --project.
  --user    install under your home directory (global for that agent)
  --project install under the current working directory (project-local; cwd matters)`,
	Example: `  lazylogcat skill install --agent cursor --user
  lazylogcat skill install --agent claude --project`,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		if skillUser == skillProject {
			return fmt.Errorf("exactly one of --user or --project is required")
		}
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("user home: %w", err)
		}
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("working directory: %w", err)
		}
		dest, err := skillinstall.Destination(skillAgent, skillUser, home, cwd)
		if err != nil {
			return err
		}
		if err := skillinstall.CopyEmbeddedSkill(skills.Lazylogcat, dest); err != nil {
			return err
		}
		fmt.Fprintln(os.Stdout, dest)
		return nil
	},
}

func init() {
	skillInstallCmd.Flags().StringVar(&skillAgent, "agent", "", "Target agent: "+skillinstall.AgentCursor+" or "+skillinstall.AgentClaude)
	skillInstallCmd.Flags().BoolVar(&skillUser, "user", false, "Install to user skills dir (~/.<agent>/skills/...)")
	skillInstallCmd.Flags().BoolVar(&skillProject, "project", false, "Install to project skills dir (./.<agent>/skills/...)")
	_ = skillInstallCmd.MarkFlagRequired("agent")

	skillCmd.AddCommand(skillInstallCmd)
	rootCmd.AddCommand(skillCmd)
}
