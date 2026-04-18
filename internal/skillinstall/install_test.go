package skillinstall

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/parfenovvs/lazylogcat/skills"
)

func TestDestination(t *testing.T) {
	t.Parallel()
	home := "/home/testuser"
	cwd := "/proj/repo"

	tests := []struct {
		name      string
		agent     string
		userScope bool
		want      string
		wantErr   bool
	}{
		{
			name:      "cursor_user",
			agent:     AgentCursor,
			userScope: true,
			want:      filepath.Join(home, ".cursor", "skills", "lazylogcat"),
		},
		{
			name:      "cursor_project",
			agent:     AgentCursor,
			userScope: false,
			want:      filepath.Join(cwd, ".agents", "skills", "lazylogcat"),
		},
		{
			name:      "claude_user",
			agent:     AgentClaude,
			userScope: true,
			want:      filepath.Join(home, ".claude", "skills", "lazylogcat"),
		},
		{
			name:      "claude_project",
			agent:     AgentClaude,
			userScope: false,
			want:      filepath.Join(cwd, ".claude", "skills", "lazylogcat"),
		},
		{
			name:      "claude_code_user",
			agent:     "claude-code",
			userScope: true,
			want:      filepath.Join(home, ".claude", "skills", "lazylogcat"),
		},
		{
			name:      "codex_user",
			agent:     "codex",
			userScope: true,
			want:      filepath.Join(home, ".codex", "skills", "lazylogcat"),
		},
		{
			name:      "codex_project",
			agent:     "codex",
			userScope: false,
			want:      filepath.Join(cwd, ".agents", "skills", "lazylogcat"),
		},
		{
			name:      "openclaw_project",
			agent:     "openclaw",
			userScope: false,
			want:      filepath.Join(cwd, "skills", "lazylogcat"),
		},
		{
			name:      "openclaw_user",
			agent:     "openclaw",
			userScope: true,
			want:      filepath.Join(home, ".openclaw", "skills", "lazylogcat"),
		},
		{
			name:      "antigravity_user",
			agent:     "antigravity",
			userScope: true,
			want:      filepath.Join(home, ".gemini", "antigravity", "skills", "lazylogcat"),
		},
		{
			name:      "windsurf_user",
			agent:     "windsurf",
			userScope: true,
			want:      filepath.Join(home, ".codeium", "windsurf", "skills", "lazylogcat"),
		},
		{
			name:      "trae_cn_project",
			agent:     "trae-cn",
			userScope: false,
			want:      filepath.Join(cwd, ".trae", "skills", "lazylogcat"),
		},
		{
			name:      "trae_cn_user",
			agent:     "trae-cn",
			userScope: true,
			want:      filepath.Join(home, ".trae-cn", "skills", "lazylogcat"),
		},
		{
			name:      "unknown_agent",
			agent:     "other",
			userScope: true,
			wantErr:   true,
		},
		{
			name:      "empty_agent",
			agent:     "",
			userScope: true,
			wantErr:   true,
		},
		{
			name:      "whitespace_only_agent",
			agent:     "  \t  ",
			userScope: true,
			wantErr:   true,
		},
		{
			name:      "cursor_trimmed",
			agent:     "  cursor  ",
			userScope: true,
			want:      filepath.Join(home, ".cursor", "skills", "lazylogcat"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := Destination(tt.agent, tt.userScope, home, cwd)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("Destination() err = nil, want error")
				}
				return
			}
			if err != nil {
				t.Fatalf("Destination() err = %v", err)
			}
			if got != tt.want {
				t.Errorf("Destination() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDestinationAllRegisteredAgents(t *testing.T) {
	t.Parallel()
	home := "/home/h"
	cwd := "/proj/wd"
	for _, agent := range ValidAgentIDs() {
		t.Run(agent+"_user", func(t *testing.T) {
			t.Parallel()
			if _, err := Destination(agent, true, home, cwd); err != nil {
				t.Fatalf("user: %v", err)
			}
		})
		t.Run(agent+"_project", func(t *testing.T) {
			t.Parallel()
			if _, err := Destination(agent, false, home, cwd); err != nil {
				t.Fatalf("project: %v", err)
			}
		})
	}
}

func TestValidAgentIDsIncludesClaudeAlias(t *testing.T) {
	t.Parallel()
	ids := ValidAgentIDs()
	if !slices.Contains(ids, AgentClaude) {
		t.Fatalf("ValidAgentIDs missing %q", AgentClaude)
	}
	if !slices.Contains(ids, "claude-code") {
		t.Fatal("ValidAgentIDs missing claude-code")
	}
}

func TestCopyEmbeddedSkill(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	if err := CopyEmbeddedSkill(skills.Lazylogcat, dir); err != nil {
		t.Fatalf("CopyEmbeddedSkill: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(dir, "SKILL.md"))
	if err != nil {
		t.Fatalf("read SKILL.md: %v", err)
	}
	if len(data) < 50 {
		t.Fatalf("SKILL.md too short: %d bytes", len(data))
	}
}
