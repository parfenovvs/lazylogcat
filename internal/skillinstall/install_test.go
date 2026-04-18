package skillinstall

import (
	"os"
	"path/filepath"
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
			want:      filepath.Join(cwd, ".cursor", "skills", "lazylogcat"),
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
			name:      "unknown_agent",
			agent:     "other",
			userScope: true,
			wantErr:   true,
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
