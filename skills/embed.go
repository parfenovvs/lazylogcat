package skills

import "embed"

// Lazylogcat is the bundled agent skill (lazylogcat/SKILL.md and optional subdirs).
//
//go:embed lazylogcat
var Lazylogcat embed.FS
