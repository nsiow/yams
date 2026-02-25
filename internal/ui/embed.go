//go:build !noui

package ui

import "embed"

//go:embed all:dist
var distFS embed.FS
