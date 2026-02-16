package buffer

import (
	"fmt"
	"strings"

	"dungeon/client/ansi"
	"dungeon/shared"
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

// NOTE: must call `OnResize()` at least once before the buffer can display any actual text
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
		tail = "\x1b[K\r\n"
	}
linesLoop:
	for _, line := range lines {
		if len(line) > b.viewSize.X {
			// lots of logic goes into a proper soft wrap
			lenLine := 0
			inEscSeq := false
			inCsiSeq := false
			lastWordBoundaryInd := 0
			lenSinceLastWordBoundary := 0
			var currentLine strings.Builder
			for i, r := range line {
				if !inEscSeq {
					switch r {
					case '\x1b':
						inEscSeq = true
						continue
					case ' ':
						currentLine.WriteString(line[lastWordBoundaryInd:i])
						lastWordBoundaryInd = i
						lenSinceLastWordBoundary = 0
					}
					if r >= 0x20 && r <= 0x7e {
						lenLine++
						lenSinceLastWordBoundary++
						if lenLine == b.viewSize.X {
							newLines = append(newLines, currentLine.String()+tail)
							currentLine.Reset()
							lastWordBoundaryInd++ // skip over space character where newline goes
							lenLine = lenSinceLastWordBoundary - 1
						}
					}
				} else if inCsiSeq {
					// https://en.wikipedia.org/wiki/ANSI_escape_code#Control_Sequence_Introducer_commands
					if r >= 0x40 && r <= 0x7e {
						inEscSeq = false
						inCsiSeq = false
					}
				} else {
					if r == '[' {
						inCsiSeq = true
					} else {
						break linesLoop // error state
					}
				}
			}
			currentLine.WriteString(line[lastWordBoundaryInd:]) // flush what's left
			newLines = append(newLines, currentLine.String()+tail)
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
