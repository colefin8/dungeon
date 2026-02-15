package buffer

import (
	"dungeon/shared"
	"fmt"
	"strings"
)

type TextBuffer struct {
	viewSize    shared.XY
	text        string
	reflowLines []string // memoized for performance
	scrollY     int
}

// NOTE: must call OnResize() at least once before the buffer can display any actual text
func NewTextBuffer(text string) TextBuffer {
	return TextBuffer{
		shared.XY{},
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

func (b *TextBuffer) setReflowLines() {
	lines := strings.Split(b.text, "\n")
	newLines := []string{}
	for _, line := range lines {
		if len(line) > b.viewSize.X {
			ind := 0
			for {
				nextPart := line[ind:min(len(line), ind+b.viewSize.X)]
				if len(nextPart) == 0 {
					break
				}
				newLines = append(newLines, nextPart+fmt.Sprintf("\x1b[K\x1b[%dD\x1b[B", len(nextPart)))
				ind += b.viewSize.X
				if ind >= len(line) {
					break
				}
			}
		} else {
			tail := fmt.Sprintf("\x1b[K\x1b[%dD\x1b[B", len(line))
			if len(line) == 0 {
				tail = "\x1b[K\x1b[B"
			}
			newLines = append(newLines, line+tail)
		}
	}
	b.reflowLines = newLines
}
func (b *TextBuffer) OnResize(viewSize shared.XY) {
	b.viewSize = viewSize
	b.setReflowLines()
}
func (b *TextBuffer) GetVisibleTextAndYPos(viewWidth int, viewHeight int) (string, int) {
	yPos := 1
	lines := b.reflowLines[:]
	bufNumLines := len(lines)
	if bufNumLines > viewHeight {
		startLine := bufNumLines - viewHeight
		startLine -= b.scrollY
		endLine := startLine + viewHeight
		lines = lines[startLine:endLine]
	} else {
		yPos += (viewHeight - bufNumLines)
	}

	return strings.Join(lines, ""), yPos
}
