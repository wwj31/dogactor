//go:build !debug || disable_TColor
// +build !debug disable_TColor

package l

type TColor = int

const (
	Blue TColor = iota + 1
	Yellow
	Green
	Magenta
	Cyan
	Gray
	White
	Red
)

var color = map[TColor]string{}
