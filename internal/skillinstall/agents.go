package skillinstall

import (
	"path/filepath"
	"sort"
)

// agentLayout holds path segments (under $HOME or cwd) to the skills root directory
// (the parent of per-skill folders like lazylogcat/).
type agentLayout struct {
	userUnderHome   []string
	projectUnderCwd []string
}

func (a agentLayout) userRoot(home string) string {
	return filepath.Join(append([]string{home}, a.userUnderHome...)...)
}

func (a agentLayout) projectRoot(cwd string) string {
	return filepath.Join(append([]string{cwd}, a.projectUnderCwd...)...)
}

func same(rel ...string) agentLayout {
	r := append([]string(nil), rel...)
	return agentLayout{userUnderHome: r, projectUnderCwd: r}
}

// Paths match Android CLI skill install locations (project + user probes).
var agentLayouts = map[string]agentLayout{
	"adal":           same(".adal", "skills"),
	"antigravity":    {userUnderHome: []string{".gemini", "antigravity", "skills"}, projectUnderCwd: []string{".agents", "skills"}},
	"augment":        same(".augment", "skills"),
	"bob":            same(".bob", "skills"),
	"claude-code":    same(".claude", "skills"),
	"codebuddy":      same(".codebuddy", "skills"),
	"codex":          {userUnderHome: []string{".codex", "skills"}, projectUnderCwd: []string{".agents", "skills"}},
	"command-code":   same(".commandcode", "skills"),
	"common":         same(".agents", "skills"),
	"continue":       same(".continue", "skills"),
	"cortex-code":    {userUnderHome: []string{".snowflake", "cortex", "skills"}, projectUnderCwd: []string{".cortex", "skills"}},
	"crush":          {userUnderHome: []string{".config", "crush", "skills"}, projectUnderCwd: []string{".crush", "skills"}},
	"cursor":         {userUnderHome: []string{".cursor", "skills"}, projectUnderCwd: []string{".agents", "skills"}},
	"deep-agents":    {userUnderHome: []string{".deepagents", "agent", "skills"}, projectUnderCwd: []string{".agents", "skills"}},
	"droid":          same(".factory", "skills"),
	"firebender":     {userUnderHome: []string{".firebender", "skills"}, projectUnderCwd: []string{".agents", "skills"}},
	"gemini":         {userUnderHome: []string{".gemini", "skills"}, projectUnderCwd: []string{".agents", "skills"}},
	"github-copilot": {userUnderHome: []string{".copilot", "skills"}, projectUnderCwd: []string{".agents", "skills"}},
	"goose":          {userUnderHome: []string{".config", "goose", "skills"}, projectUnderCwd: []string{".goose", "skills"}},
	"iflow":          same(".iflow", "skills"),
	"junie":          same(".junie", "skills"),
	"kilo-code":      {userUnderHome: []string{".kilo", "skills"}, projectUnderCwd: []string{".kilocode", "skills"}},
	"kiro":           same(".kiro", "skills"),
	"kode":           same(".kode", "skills"),
	"mcpjam":         same(".mcpjam", "skills"),
	"mistral-vibe":   same(".vibe", "skills"),
	"mux":            same(".mux", "skills"),
	"neovate":        same(".neovate", "skills"),
	"openclaw":       {userUnderHome: []string{".openclaw", "skills"}, projectUnderCwd: []string{"skills"}},
	"opencode":       {userUnderHome: []string{".config", "opencode", "skills"}, projectUnderCwd: []string{".agents", "skills"}},
	"openhands":      same(".openhands", "skills"),
	"pi":             {userUnderHome: []string{".pi", "agent", "skills"}, projectUnderCwd: []string{".pi", "skills"}},
	"pochi":          same(".pochi", "skills"),
	"qoder":          same(".qoder", "skills"),
	"qwen-code":      same(".qwen", "skills"),
	"roo-code":       same(".roo", "skills"),
	"trae":           same(".trae", "skills"),
	"trae-cn":        {userUnderHome: []string{".trae-cn", "skills"}, projectUnderCwd: []string{".trae", "skills"}},
	"universal":      {userUnderHome: []string{".config", "agents", "skills"}, projectUnderCwd: []string{".agents", "skills"}},
	"windsurf":       {userUnderHome: []string{".codeium", "windsurf", "skills"}, projectUnderCwd: []string{".windsurf", "skills"}},
	"zencoder":       same(".zencoder", "skills"),
}

// normalizeAgent maps lazylogcat-only aliases to registry keys.
func normalizeAgent(agent string) string {
	if agent == AgentClaude {
		return "claude-code"
	}
	return agent
}

// ValidAgentIDs returns sorted agent names accepted by --agent (includes "claude" alias).
func ValidAgentIDs() []string {
	out := make([]string, 0, len(agentLayouts)+1)
	for id := range agentLayouts {
		out = append(out, id)
	}
	out = append(out, AgentClaude)
	sort.Strings(out)
	return out
}
