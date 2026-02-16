package main

import (
	"fmt"
	"os"
	"time"

	"golang.org/x/sys/unix"

	"dungeon/client/ansi"
	"dungeon/client/buffer"
	"dungeon/shared"
)

const PENCIL = "\u270e"

var hasAllTextAppeared = false
var inputReady = make(chan bool, 1)
var nameInputBuffer buffer.InputBuffer
var nameInputSubmit = make(chan string, 256)

type WelcomeView struct{}

func (_ WelcomeView) Init() {
	nameInputBuffer = buffer.NewInputBuffer(
		bgCol,
		nameInputSubmit,
	)
}

func (_ WelcomeView) Update() {
	<-inputReady

	for {
		ansi.ShowCursor()
		e := <-inputStreamSet.Input
		nameInputBuffer.Update(e)
		select {
		case txt := <-nameInputSubmit:
			MudConnection.Write(append([]byte{shared.RequestTypeLogin}, []byte(txt+"\n")...))
			ProgramMode = ProgramModeMud
			return
		default:
		}
	}
}

func (_ WelcomeView) Render() {
	ansi.HideCursor()
	// display welcome graphic
	welcomeGraphicPos := shared.XY{X: 0, Y: 0}
	const WELCOME_TEXT_AREA_WIDTH = 32
	switch Dimension {
	case DimensionXl:
		welcomeGraphicPos = shared.XY{X: (TermSize.X / 2) - ((WELCOME_TEXT_AREA_WIDTH + ARCHWAY_WIDTH) / 2), Y: (TermSize.Y / 2) - (ARCHWAY_HEIGHT / 2)}
	case DimensionTall:
		welcomeGraphicPos = shared.XY{X: (TermSize.X / 2) - (ARCHWAY_WIDTH / 2), Y: 0}
	}
	if Dimension != DimensionS {
		ansi.MoveCursorTo(welcomeGraphicPos.X+1, welcomeGraphicPos.Y+1)
		f, err := os.Open("archway.bin")
		if err != nil {
			fmt.Printf("could not open archway file: %v\n", err)
			os.Exit(1)
		}
		stat, _ := f.Stat()
		archwaySize := stat.Size()
		archwayBuf := make([]byte, archwaySize)
		f.Read(archwayBuf)
		unix.Write(int(os.Stdout.Fd()), archwayBuf)
	}

	const WELCOME_TEXT = "Welcome...."
	welcomeTextPos := shared.XY{X: (TermSize.X / 2) - (len(WELCOME_TEXT) / 2), Y: 2}
	switch Dimension {
	case DimensionXl:
		welcomeTextPos.X = welcomeGraphicPos.X + ARCHWAY_WIDTH + 4
		welcomeTextPos.Y = TermSize.Y / 2
	case DimensionTall:
		welcomeTextPos.Y = welcomeGraphicPos.Y + ARCHWAY_HEIGHT + 1
	}
	ansi.MoveCursorTo(welcomeTextPos.X, welcomeTextPos.Y)
	ansi.Set24BitFgCol(shared.Color{R: 255, G: 255, B: 255})
	if !hasAllTextAppeared {
		time.Sleep(1 * time.Second)
		RollText(WELCOME_TEXT, RollSpeedSlow)
	} else {
		fmt.Print(WELCOME_TEXT)
	}

	if !hasAllTextAppeared {
		time.Sleep(1 * time.Second)
	}

	regularTextPos := shared.XY{X: 2, Y: welcomeTextPos.Y + 2}
	switch Dimension {
	case DimensionXl:
		regularTextPos.X = welcomeTextPos.X
	}
	ansi.MoveCursorTo(regularTextPos.X, regularTextPos.Y)
	ansi.Set24BitFgCol(shared.Color{R: 116, G: 98, B: 80})
	fmt.Print("What be thy name?")

	ansi.MoveCursorTo(regularTextPos.X, regularTextPos.Y+1)
	fmt.Printf("%s ", PENCIL)
	ansi.Set24BitFgCol(shared.Color{R: 253, G: 213, B: 174})
	ansi.SetBold()
	ansi.ShowCursor()

	nameInputBuffer.OnResize(
		shared.XY{X: regularTextPos.X + 2, Y: regularTextPos.Y + 1},
		shared.XY{X: 32, Y: 1},
	)

	if !hasAllTextAppeared {
		hasAllTextAppeared = true
		inputReady <- true
	}
}
