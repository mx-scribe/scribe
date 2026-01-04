// Package web provides embedded web assets for the SCRIBE dashboard.
package web

import "embed"

// DistFS contains the embedded web dist files.
// This will be populated at build time.
//
//go:embed all:dist
var DistFS embed.FS
