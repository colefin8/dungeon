package buffer

import (
	"fmt"
	"os"

	"dungeon/client/ansi"
	"dungeon/client/input"
	"dungeon/shared"
)

type InputBuffer struct {
	viewPos   shared.XY
	viewSize  shared.XY
	bgCol     shared.Color
	text      string
	cursorIdx int
	submit    chan string
}

// NOTE: must call OnResize() at least once before the buffer can display any actual text
func NewInputBuffer(
	bgCol shared.Color,
	submit chan string,
) InputBuffer {
	return InputBuffer{
		shared.XY{},
		shared.XY{},
		bgCol,
		"",
		0,
		submit,
	}
}

func (ip *InputBuffer) Reset() {
	ip.text = ""
	ip.cursorIdx = 0
	ip.Render()
}

func (ip *InputBuffer) Update(e input.IEvent) {
	switch e := e.(type) {
	case input.KeyEvent:
		switch e.Key {
		case '\b':
			if ip.cursorIdx > 0 {
				os.Stdout.WriteString("\x1b[D")
				os.Stdout.WriteString(ip.text[ip.cursorIdx:] + " ")
				os.Stdout.WriteString(fmt.Sprintf("\x1b[%dD", (len(ip.text)-ip.cursorIdx)+1))
				ip.cursorIdx--
				ip.text = ip.text[:ip.cursorIdx] + ip.text[ip.cursorIdx+1:]
			}
		case '\n':
			if len(ip.text) > 0 {
				ip.submit <- ip.text
				ip.Reset()
			}
		case '\x1b':
			ip.Reset()
		default:
			switch e.Mod {
			case input.KeyModNone:
				if len(ip.text) < ip.viewSize.X {
					os.Stdout.WriteString(string(e.Key))
					if ip.cursorIdx == len(ip.text) {
						ip.text += string(e.Key)
					} else {
						os.Stdout.WriteString(ip.text[ip.cursorIdx:])
						os.Stdout.WriteString(fmt.Sprintf("\x1b[%dD", len(ip.text)-ip.cursorIdx))
						ip.text = ip.text[:ip.cursorIdx] + string(e.Key) + ip.text[ip.cursorIdx:]
					}
					ip.cursorIdx++
				}
			}
		}
	case input.NonalphaKeyEvent:
		switch e.Key {
		case input.NonalphaKeyLeft:
			if ip.cursorIdx > 0 {
				ip.cursorIdx--
				os.Stdout.WriteString("\x1b[D")
			}
		case input.NonalphaKeyRight:
			if ip.cursorIdx < len(ip.text) {
				ip.cursorIdx++
				os.Stdout.WriteString("\x1b[C")
			}
		}
	}
}

func (ip *InputBuffer) OnResize(viewPos shared.XY, viewSize shared.XY) {
	ip.viewPos = viewPos
	ip.viewSize = viewSize
	// TEMP: in the future I will make the input textarea multi-line, for now I'm just gonna truncate the msg if you shrink the window lol
	if len(ip.text) > ip.viewSize.X {
		ip.text = ip.text[:ip.viewSize.X]
	}
	ip.cursorIdx = min(ip.cursorIdx, len(ip.text))
}

func (ip *InputBuffer) Render() {
	ansi.MoveCursorTo(ip.viewPos.X, ip.viewPos.Y)
	ansi.Set24BitBgCol(ip.bgCol)
	ansi.SetFgCol(ansi.AnsiColorWhite, true)
	fmt.Print("\x1b[K")
	fmt.Print(ip.text)
	ansi.MoveCursorHorTo(ip.viewPos.X + ip.cursorIdx)
	ansi.ShowCursor()
}
