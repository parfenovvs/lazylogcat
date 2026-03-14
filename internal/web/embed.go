package web

import "embed"

//go:generate bun run --cwd ../../web-ui build

//go:embed static/*
var staticFS embed.FS
