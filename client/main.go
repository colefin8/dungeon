package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"dungeon/client/ansi"
	"dungeon/client/input"
	"dungeon/shared"

	"golang.org/x/term"
)

const SOCKET = "/home/dungeon/.dungeon.sock"
const PROMPT = "\x1b[33m\u2767 \x1b[0m"
const MANICULE = "\u270e"

const ARCHWAY_WIDTH = 64
const ARCHWAY_HEIGHT = 32

// I despise Go enums
type ProgramModeKind = uint
type DimensionKind = uint

const (
	ProgramModeKindWelcome ProgramModeKind = iota
	ProgramModeKindMud
)
const (
	DimensionKindXl DimensionKind = iota
	DimensionKindTall
	DimensionKindS
)

var ProgramMode ProgramModeKind
var Dimension DimensionKind

const (
	DIMENSION_XL_MIN_WIDTH    = 120
	DIMENSION_XL_MIN_HEIGHT   = 32
	DIMENSION_TALL_MIN_WIDTH  = 64
	DIMENSION_TALL_MIN_HEIGHT = 42
)

type IView interface {
	Init()
	Render()
	Update()
}

var scenes = map[ProgramModeKind]IView{
	ProgramModeKindWelcome: WelcomeView{},
	ProgramModeKindMud:     MudView{},
}

var bgCol = shared.Color{R: 8, G: 8, B: 8}
var TermSize shared.XY
var prevTermState *term.State

// input channels
var inputStreamSet = input.StreamSet{
	Input: make(chan input.IEvent),
	Quit:  make(chan bool),
}

func main() {
	// fmt.Printf("dialing socket %s...\n", SOCKET)
	// conn, err := net.Dial("unix", SOCKET)
	// if err != nil {
	// 	if errors.Is(err, syscall.ECONNREFUSED) {
	// 		fmt.Println("ERROR: server is not running!")
	// 	} else {
	// 		fmt.Printf("ERROR: could not dial unix socket %s: %v\n", SOCKET, err)
	// 	}
	// 	os.Exit(1)
	// }
	// defer conn.Close()

	// ProgramMode = ProgramModeKindMud
	ProgramMode = ProgramModeKindWelcome

	// prep manual input handling
	prevTermState, _ = term.MakeRaw(int(os.Stdin.Fd()))
	defer term.Restore(int(os.Stdin.Fd()), prevTermState)
	ansi.EnableMouseInput()
	defer ansi.DisableMouseInput()
	ansi.HideCursor()
	defer ansi.ShowCursor()

	go input.StartPollingInput(inputStreamSet)

	// check for quit
	go func() {
		<-inputStreamSet.Quit
		quit(1)
	}()

	scenes[ProgramMode].Init()
	render()

	// resize listener
	sigResize := make(chan os.Signal, 1)
	signal.Notify(sigResize, syscall.SIGWINCH)
	go func() {
		for range sigResize {
			render()
		}
	}()

	for {
		scenes[ProgramMode].Update()
	}

	// scanner := bufio.NewScanner(os.Stdin)

	// // get username before entering the dungeon
	// fmt.Println("\nWhat be your name, friend?")
	// fmt.Print(" \x1b[33m", MANICULE, " \x1b[0m")
	// scanner.Scan()
	// username := scanner.Text() + "\n"

	// go printIncomingMessages(conn)

	// bytesToWrite := append([]byte{shared.MessageTypeConnect}, []byte(username)...)
	// conn.Write(bytesToWrite)

	// for {
	// 	fmt.Printf("\x1b[%d;2H", termSize.y-1)
	// 	fmt.Print(PROMPT)
	// 	if !scanner.Scan() {
	// 		break
	// 	}

	// 	line := scanner.Text()
	// 	if line == "quit" {
	// 		return
	// 	}
	// 	conn.Write(fmt.Appendf([]byte{shared.MessageTypeSay}, "%s\n", line))
	// }
}

func printIncomingMessages(conn net.Conn) {
	scanner := bufio.NewScanner(conn)
	for {
		scanStatus := scanner.Scan()
		if scanStatus {
			line := scanner.Text()
			fmt.Print(line)
		} else {
			fmt.Println("server disconnected!!")
			os.Exit(1)
		}
	}
}

func render() {
	TermSize.X, TermSize.Y, _ = term.GetSize(int(os.Stdin.Fd()))
	ansi.ClearScreenWithCol(TermSize.X, TermSize.Y, bgCol)
	Dimension = DimensionKindS
	if TermSize.X >= DIMENSION_XL_MIN_WIDTH && TermSize.Y >= DIMENSION_XL_MIN_HEIGHT {
		Dimension = DimensionKindXl
	} else if TermSize.X >= DIMENSION_TALL_MIN_WIDTH && TermSize.Y >= DIMENSION_TALL_MIN_HEIGHT {
		Dimension = DimensionKindTall
	}

	ansi.MoveCursorToTopLeft()
	ansi.Set24BitFgCol(shared.Color{R: 60, G: 60, B: 60})
	fmt.Println(TermSize.X, "x", TermSize.Y)

	scenes[ProgramMode].Render()
}

func quit(code int) {
	ansi.ShowCursor()
	ansi.DisableMouseInput()
	term.Restore(int(os.Stdin.Fd()), prevTermState)
	os.Exit(code)
}
