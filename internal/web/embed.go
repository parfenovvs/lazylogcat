package web

import "embed"

//go:generate bun --cwd ../../web-ui run build

//go:embed static/*
var staticFS embed.FS
