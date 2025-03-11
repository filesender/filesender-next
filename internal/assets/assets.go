package assets

import "embed"

//go:embed public/*
var EmbeddedPublicFiles embed.FS

//go:embed templates/*
var EmbeddedTemplateFiles embed.FS
