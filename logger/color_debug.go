//go:build debug && !disable_color

package logger

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

var color map[TColor]struct{ B, E string }

func init() {
	color = map[TColor]struct{ B, E string }{
		Blue:    {B: "\x1b[0;34m", E: "\x1b[0m"},
		Yellow:  {B: "\x1b[0;33m", E: "\x1b[0m"},
		Green:   {B: "\x1b[0;32m", E: "\x1b[0m"},
		Magenta: {B: "\x1b[0;35m", E: "\x1b[0m"},
		Cyan:    {B: "\x1b[0;36m", E: "\x1b[0m"},
		Gray:    {B: "\x1b[0;37m", E: "\x1b[0m"},
		White:   {B: "\x1b[1;37m", E: "\x1b[0m"},
		Red:     {B: "\x1b[0;31m", E: "\x1b[0m"},
	}
}
