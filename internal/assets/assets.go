// Package assets contains all embedded template & public / static files
package assets

import "embed"

// EmbeddedPublicFiles contains public / static files
//
//go:embed public/*
var EmbeddedPublicFiles embed.FS

// EmbeddedTemplateFiles contains template files
//
//go:embed templates/*
var EmbeddedTemplateFiles embed.FS
