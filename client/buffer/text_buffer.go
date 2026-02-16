package buffer

import (
	"dungeon/client/ansi"
	"dungeon/shared"
	"fmt"
	"strings"
)

type TextBuffer struct {
	viewPos     shared.XY
	viewSize    shared.XY
	bgCol       shared.Color
	fgCol       shared.Color
	text        string
	reflowLines []string // memoized for performance
	scrollY     int
}

// NOTE: must call OnResize() at least once before the buffer can display any actual text
func NewTextBuffer(
	bgCol shared.Color,
	fgCol shared.Color,
	text string,
) TextBuffer {
	return TextBuffer{
		shared.XY{},
		shared.XY{},
		bgCol,
		fgCol,
		text,
		[]string{},
		0,
	}
}

func (b *TextBuffer) ScrollUp() bool {
	if len(b.reflowLines) > b.viewSize.Y+b.scrollY {
		b.scrollY++
		return true
	}
	return false
}
func (b *TextBuffer) ScrollDown() bool {
	if b.scrollY > 0 {
		b.scrollY--
		return true
	}
	return false
}

func (b *TextBuffer) Append(txt string) {
	b.text += txt
	b.setReflowLines()
}

func (b *TextBuffer) setReflowLines() {
	lines := strings.Split(b.text, "\n")
	newLines := []string{}
	tail := fmt.Sprintf("\x1b[K\r\n\x1b[%dC", b.viewPos.X-1)
	if b.viewPos.X == 0 {
		tail = "\r\n"
	}
	for _, line := range lines {
		if len(line) > b.viewSize.X {
			ind := 0
			for {
				nextPart := line[ind:min(len(line), ind+b.viewSize.X)]
				if len(nextPart) == 0 {
					break
				}
				newLines = append(newLines, nextPart+tail)
				ind += b.viewSize.X
				if ind >= len(line) {
					break
				}
			}
		} else {
			newLines = append(newLines, line+tail)
		}
	}
	b.reflowLines = newLines
}
func (b *TextBuffer) OnResize(viewPos shared.XY, viewSize shared.XY) {
	b.viewPos = viewPos
	b.viewSize = viewSize
	b.setReflowLines()
}
func (b *TextBuffer) Render() {
	lines, drawYPos := b.getVisibleTextAndYPos()
	ansi.MoveCursorTo(b.viewPos.X, drawYPos)
	ansi.Set24BitFgCol(b.fgCol)
	ansi.Set24BitBgCol(b.bgCol)
	fmt.Print(lines)
}
func (b *TextBuffer) getVisibleTextAndYPos() (string, int) {
	yPos := b.viewPos.Y
	lines := b.reflowLines[:]
	bufNumLines := len(lines)
	if bufNumLines > b.viewSize.Y {
		startLine := bufNumLines - b.viewSize.Y
		startLine -= b.scrollY
		endLine := startLine + b.viewSize.Y
		lines = lines[startLine:endLine]
	} else {
		yPos += (b.viewSize.Y - bufNumLines)
	}

	return strings.Join(lines, ""), yPos
}
