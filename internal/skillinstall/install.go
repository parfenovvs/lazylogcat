package skillinstall

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

const (
	AgentCursor = "cursor"
	AgentClaude = "claude"
)

// Destination returns the directory where the lazylogcat skill should be installed
// (the folder that will contain SKILL.md), e.g. ~/.cursor/skills/lazylogcat.
// home is typically os.UserHomeDir(); cwd is typically os.Getwd() for project scope.
func Destination(agent string, userScope bool, home, cwd string) (string, error) {
	agent = strings.TrimSpace(agent)
	if agent == "" {
		return "", fmt.Errorf("agent name is empty")
	}
	key := normalizeAgent(agent)
	layout, ok := agentLayouts[key]
	if !ok {
		return "", fmt.Errorf("unknown agent %q", agent)
	}
	var skillsRoot string
	if userScope {
		skillsRoot = layout.userRoot(home)
	} else {
		skillsRoot = layout.projectRoot(cwd)
	}
	return filepath.Join(skillsRoot, "lazylogcat"), nil
}

// CopyEmbeddedSkill copies the embedded lazylogcat/ tree from srcFS into destDir.
// srcFS must be the embed.FS that includes a top-level "lazylogcat" directory (see skills package).
func CopyEmbeddedSkill(srcFS fs.FS, destDir string) error {
	sub, err := fs.Sub(srcFS, "lazylogcat")
	if err != nil {
		return fmt.Errorf("skill bundle: %w", err)
	}
	return copyFSIntoDir(sub, destDir)
}

func copyFSIntoDir(fsys fs.FS, destRoot string) error {
	return fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == "." {
			return os.MkdirAll(destRoot, 0o755)
		}
		out := filepath.Join(destRoot, filepath.FromSlash(path))
		if d.IsDir() {
			return os.MkdirAll(out, 0o755)
		}
		if err := os.MkdirAll(filepath.Dir(out), 0o755); err != nil {
			return err
		}
		data, err := fs.ReadFile(fsys, path)
		if err != nil {
			return err
		}
		return os.WriteFile(out, data, 0o644)
	})
}
