//go:build dev
// +build dev

package main

import "io/fs"

// webFS is nil in development mode - dashboard is served separately by Vite
var webFS fs.FS
