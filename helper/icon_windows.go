//go:build windows

package helper

import _ "embed"

//go:embed icon.ico
var Icon []byte
var Mime = "image/x-icon"
