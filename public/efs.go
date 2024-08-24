package ui

import "embed"

// Files provides a means to embed files in the resulting binary
//
//go:embed "css" "font" "img" "js" "txt"
var Files embed.FS
