package firedancer

import (
	"embed"
)

//go:embed assets
var assets embed.FS

const (
	assetsInstall = "assets/install"
)