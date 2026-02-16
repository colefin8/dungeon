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
const ARCHWAY_L_WIDTH = 64
const ARCHWAY_S_WIDTH = 55
const ARCHWAY_L_HEIGHT = 32
const ARCHWAY_S_HEIGHT = 16

var hasAllTextAppeared = false
var inputReady = make(chan bool, 1)
var nameInputBuffer buffer.InputBuffer
var nameInputSubmit = make(chan string, 256)

type WelcomeView struct{}

func (_ WelcomeView) Init() {
	nameInputBuffer = buffer.NewInputBuffer(
		bgCol,
		shared.Color{R: 253, G: 213, B: 174},
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

	// load up welcome graphic binaries
	archwayLFile, err := os.Open("archway.bin")
	if err != nil {
		fmt.Printf("could not open archway file: %v\n", err)
		inputStreamSet.Quit <- true
	}
	stat, _ := archwayLFile.Stat()
	fileSize := stat.Size()
	archwayLBuf := make([]byte, fileSize)
	archwayLFile.Read(archwayLBuf)

	archwaySFile, err := os.Open("sword.bin")
	if err != nil {
		fmt.Printf("could not open small archway file: %v\n", err)
		inputStreamSet.Quit <- true
	}
	stat, _ = archwaySFile.Stat()
	fileSize = stat.Size()
	archwaySBuf := make([]byte, fileSize)
	archwaySFile.Read(archwaySBuf)

	// display welcome graphic
	const WELCOME_TEXT_AREA_WIDTH = 32
	welcomeGraphicPos := shared.XY{X: 0, Y: 0}
	switch Dimension {
	case DimensionXl:
		welcomeGraphicPos = shared.XY{X: (TermSize.X / 2) - ((WELCOME_TEXT_AREA_WIDTH + ARCHWAY_L_WIDTH) / 2), Y: (TermSize.Y / 2) - (ARCHWAY_L_HEIGHT / 2)}
	case DimensionTall:
		welcomeGraphicPos = shared.XY{X: (TermSize.X / 2) - (ARCHWAY_L_WIDTH / 2), Y: 0}
	case DimensionM:
		welcomeGraphicPos = shared.XY{X: (TermSize.X / 2) - (ARCHWAY_S_WIDTH / 2), Y: 0}
	}
	if Dimension == DimensionXl || Dimension == DimensionTall {
		ansi.MoveCursorTo(welcomeGraphicPos.X+1, welcomeGraphicPos.Y+1)
		unix.Write(int(os.Stdout.Fd()), archwayLBuf)
	} else if Dimension == DimensionM {
		ansi.MoveCursorTo(welcomeGraphicPos.X+1, welcomeGraphicPos.Y+1)
		unix.Write(int(os.Stdout.Fd()), archwaySBuf)
	}

	const WELCOME_TEXT = "Welcome...."
	welcomeTextPos := shared.XY{X: (TermSize.X / 2) - (len(WELCOME_TEXT) / 2), Y: 2}
	switch Dimension {
	case DimensionXl:
		welcomeTextPos.X = welcomeGraphicPos.X + ARCHWAY_L_WIDTH + 4
		welcomeTextPos.Y = TermSize.Y / 2
	case DimensionTall:
		welcomeTextPos.Y = welcomeGraphicPos.Y + ARCHWAY_L_HEIGHT + 1
	case DimensionM:
		welcomeTextPos.Y = welcomeGraphicPos.Y + ARCHWAY_S_HEIGHT + 1
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
	nameInputBuffer.Render()

	if !hasAllTextAppeared {
		hasAllTextAppeared = true
		inputReady <- true
	}
}
