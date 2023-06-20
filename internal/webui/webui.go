package webui

import (
	"embed"
	"io/fs"
)

//go:embed ui/*
var assets embed.FS

// UI is the embeded frontend
var UI, _ = fs.Sub(assets, "ui")
