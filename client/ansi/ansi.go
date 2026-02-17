package ansi

import (
	"fmt"
	"os"

	"dungeon/shared"

	"golang.org/x/sys/unix"
)

type AnsiColor uint

const (
	AnsiColorBlack AnsiColor = iota
	AnsiColorRed
	AnsiColorGreen
	AnsiColorYellow
	AnsiColorBlue
	AnsiColorMagenta
	AnsiColorCyan
	AnsiColorWhite
)

const csi = "\x1b["

func SwitchToAlternateScreenBuffer() {
	fmt.Print(csi + "?1049h")
}
func SwitchToMainScreenBuffer() {
	fmt.Print(csi + "?1049l")
}

func SetFgCol(col AnsiColor, isBright bool) {
	brightAdd := 0
	if isBright {
		brightAdd = 60
	}
	fmt.Printf(csi+"%dm", int(col)+30+brightAdd)
}
func SetBgCol(col AnsiColor, isBright bool) {
	brightAdd := 0
	if isBright {
		brightAdd = 60
	}
	fmt.Printf(csi+"%dm", int(col)+40+brightAdd)
}
func Set24BitBgCol(col shared.Color) {
	fmt.Printf(csi+"48;2;%d;%d;%dm", col.R, col.G, col.B)
}
func Set24BitFgCol(col shared.Color) {
	fmt.Printf(csi+"38;2;%d;%d;%dm", col.R, col.G, col.B)
}
func SetBold() {
	fmt.Print(csi + "1m")
}
func SetFaint() {
	fmt.Print(csi + "2m")
}
func ResetIntensity() {
	fmt.Print(csi + "22m")
}

func HideCursor() {
	fmt.Print(csi + "?25l")
}
func ShowCursor() {
	fmt.Print(csi + "?25h")
}

// 1-based (top left is {1, 1})
func MoveCursorTo(x int, y int) {
	fmt.Printf(csi+"%d;%dH", y, x)
}
func MoveCursorToTopLeft() {
	fmt.Print(csi + "H")
}
func MoveCursorHorTo(x int) {
	fmt.Printf(csi+"%dG", x)
}
func MoveCursorToLineStart() {
	fmt.Print(csi + "G")
}

func ClearScreenWithCol(w int, h int, col shared.Color) {
	MoveCursorToTopLeft()
	Set24BitBgCol(col)
	unix.Write(int(os.Stdout.Fd()), []byte(csi+"J"))
}
func ClearLineWithCol(w int, col shared.Color) {
	MoveCursorToLineStart()
	Set24BitBgCol(col)
	unix.Write(int(os.Stdout.Fd()), []byte(csi+"K"))
}

func EnableMouseInput() {
	fmt.Print(csi + "?1000h" + csi + "?1002h" + csi + "?1006h")
}
func DisableMouseInput() {
	fmt.Print(csi + "?1000l" + csi + "?1002l" + csi + "?1006l")
}

func RequestCursorPos() {
	fmt.Print(csi + "6n")
}
