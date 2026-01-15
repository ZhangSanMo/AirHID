package icon

import (
	_ "embed"
)

//go:embed icon.ico
var iconData []byte

func Data() []byte {
	return iconData
}