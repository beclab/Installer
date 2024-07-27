//go:build !dev
// +build !dev

package frontend

import "embed"

//go:embed dist/spa/*
var assets embed.FS

func Assets() embed.FS {
	return assets
}
