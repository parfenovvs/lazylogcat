package cmd

import (
	"fmt"
	"os"
	"strings"

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
	Long:  "", // set in init() to include supported agents list
	Example: `  lazylogcat skill install --agent cursor --user
  lazylogcat skill install --agent claude-code --project
  lazylogcat skill install --agent codex --user`,
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
	skillInstallCmd.Long = strings.TrimSpace(`Copies the bundled skill from this binary into the agent-specific skills directory.

Supported agents: ` + strings.Join(skillinstall.ValidAgentIDs(), ", ") + `

You must pass --agent and exactly one of --user or --project.
  --user    install under your home directory (global for that agent)
  --project install under the current working directory (project-local; cwd matters)`)
	skillInstallCmd.Flags().StringVar(&skillAgent, "agent", "", "Target agent name (see command help for the full list; tab completes)")
	skillInstallCmd.Flags().BoolVar(&skillUser, "user", false, "Install to user skills dir (agent-specific path under $HOME)")
	skillInstallCmd.Flags().BoolVar(&skillProject, "project", false, "Install to project skills dir (agent-specific path under cwd)")
	_ = skillInstallCmd.MarkFlagRequired("agent")
	_ = skillInstallCmd.RegisterFlagCompletionFunc("agent", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return skillinstall.ValidAgentIDs(), cobra.ShellCompDirectiveNoFileComp | cobra.ShellCompDirectiveNoSpace
	})

	skillCmd.AddCommand(skillInstallCmd)
	rootCmd.AddCommand(skillCmd)
}
