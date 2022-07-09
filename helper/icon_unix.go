//go:build linux || darwin

package helper

import _ "embed"

//go:embed icon.png
var Icon []byte
var Mime = "image/png"
