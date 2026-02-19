package main

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"net"
	"os"
	"os/signal"
	"syscall"

	"dungeon/client/ansi"
	"dungeon/client/input"
	"dungeon/shared"

	"golang.org/x/term"
)

const PROMPT = "\x1b[33m\u2767 \x1b[0m"
const MANICULE = "\u270e"

// I despise Go enums
type ProgramModeKind = uint
type DimensionKind = uint

const (
	ProgramModeWelcome ProgramModeKind = iota
	ProgramModeMud
)
const (
	DimensionXl DimensionKind = iota
	DimensionTall
	DimensionM
	DimensionS
)

var ProgramMode ProgramModeKind
var Dimension DimensionKind

const (
	DIMENSION_XL_MIN_WIDTH    = 120
	DIMENSION_XL_MIN_HEIGHT   = 32
	DIMENSION_TALL_MIN_WIDTH  = 64
	DIMENSION_TALL_MIN_HEIGHT = 42
	DIMENSION_M_MIN_WIDTH     = 55
	DIMENSION_M_MIN_HEIGHT    = 24
)

type IView interface {
	Init()
	Render()
	Update()
	ProcessServerMessage(data []byte)
}

var scenes = map[ProgramModeKind]IView{
	ProgramModeWelcome: WelcomeView{},
	ProgramModeMud:     MudView{},
}

var tty *os.File
var bgCol = shared.Color{R: 8, G: 8, B: 8}
var txtCol = shared.Color{R: 96, G: 96, B: 96}
var TermSize shared.XY
var prevTermState *term.State
var MudConnection net.Conn

// input channels
var inputStreamSet = input.StreamSet{
	Input:     make(chan input.IEvent),
	CursorPos: make(chan input.IEvent),
	Quit:      make(chan bool),
}

func main() {
	fmt.Printf("dialing socket %s...\n", shared.SOCKET_PATH)
	conn, err := net.Dial("unix", shared.SOCKET_PATH)
	if err != nil {
		if errors.Is(err, syscall.ECONNREFUSED) {
			fmt.Println("ERROR: server is not running!")
		} else {
			fmt.Printf("ERROR: could not dial unix socket %s: %v\n", shared.SOCKET_PATH, err)
		}
		os.Exit(1)
	}
	MudConnection = conn
	defer MudConnection.Close()

	ProgramMode = ProgramModeWelcome

	tty, err = os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		fmt.Println("ERROR: could not open /dev/tty for reading and writing")
		os.Exit(1)
	}
	defer tty.Close()

	ansi.SwitchToAlternateScreenBuffer(tty)
	defer ansi.SwitchToMainScreenBuffer(tty)

	// prep manual input handling
	if prevTermState, err = term.MakeRaw(int(tty.Fd())); err != nil {
		tty.Close()
		fmt.Println("ERROR: could not make /dev/tty raw")
		os.Exit(1)
	}
	defer term.Restore(int(tty.Fd()), prevTermState)
	ansi.EnableMouseInput(tty)
	defer ansi.DisableMouseInput(tty)
	ansi.HideCursor()
	defer ansi.ShowCursor()

	go input.StartPollingInput(tty, inputStreamSet)
	go receiveMessagesFromServer()

	// check for quit
	go func() {
		<-inputStreamSet.Quit
		quit(1)
	}()

	for {
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

		scenes[ProgramMode].Update()
	}
}

func receiveMessagesFromServer() {
	// this will allocate 65,536 bytes of RAM per user. My raspberry pi has 16GB of RAM total, 14GB roughly it says is free. This means about 229,376 users could be online at a time.
	// even the most popular MUDs in the world in 2026 struggle to hit 100 users online at a time so this is plenty
	dataBuf := make([]byte, math.MaxUint16+1)
	for {
		n, err := MudConnection.Read(dataBuf)
		lenData := binary.LittleEndian.Uint16(dataBuf)
		if err == nil && int(lenData) == n-2 {
			data := dataBuf[2:n]
			scenes[ProgramMode].ProcessServerMessage(data)
		}
	}
}

func render() {
	TermSize.X, TermSize.Y, _ = term.GetSize(int(os.Stdin.Fd()))
	ansi.ClearScreenWithCol(TermSize.X, TermSize.Y, bgCol)
	Dimension = DimensionS
	if TermSize.X >= DIMENSION_XL_MIN_WIDTH && TermSize.Y >= DIMENSION_XL_MIN_HEIGHT {
		Dimension = DimensionXl
	} else if TermSize.X >= DIMENSION_TALL_MIN_WIDTH && TermSize.Y >= DIMENSION_TALL_MIN_HEIGHT {
		Dimension = DimensionTall
	} else if TermSize.X >= DIMENSION_M_MIN_WIDTH && TermSize.Y >= DIMENSION_M_MIN_HEIGHT {
		Dimension = DimensionM
	}

	// DEBUG
	ansi.MoveCursorToTopLeft()
	ansi.Set24BitFgCol(shared.Color{R: 12, G: 12, B: 12})
	fmt.Println(TermSize.X, "x", TermSize.Y)

	scenes[ProgramMode].Render()
}

func quit(code int) {
	ansi.ShowCursor()
	ansi.DisableMouseInput(tty)
	term.Restore(int(os.Stdin.Fd()), prevTermState)
	os.Exit(code)
}
