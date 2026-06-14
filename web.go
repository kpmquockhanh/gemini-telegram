//go:build !dev
// +build !dev

package main

import (
	"embed"
	"io/fs"
)

//go:embed all:dashboard/dist
var embeddedFS embed.FS

var webFS fs.FS = embeddedFS
