package main

import (
	"encoding/binary"
	"fmt"
	"strings"
	"time"

	"dungeon/client/ansi"
	"dungeon/client/buffer"
	"dungeon/client/input"
	"dungeon/shared"
)

var promptBgCol = shared.Color{R: 24, G: 24, B: 24}
var promptFgCol = shared.Color{R: 255, G: 255, B: 255}

const CURSOR_MIN_X_POS = 5
const PROMPT_FLOWER = "\u2767"
const MSG_BUFFER_PADDING_HOR = 2
const MSG_BUFFER_MAX_WIDTH = 64

var msgBuffer buffer.TextBuffer
var inputBuffer buffer.InputBuffer
var inputSubmit = make(chan string, 512)

type MudView struct{}

func (_ MudView) Init() {
	msgBuffer = buffer.NewTextBuffer(
		bgCol,
		txtCol,
		"",
	)
	inputBuffer = buffer.NewInputBuffer(
		promptBgCol,
		promptFgCol,
		inputSubmit,
	)
}

func (_ MudView) ProcessServerMessage(data []byte) {
	switch data[0] {
	case shared.ResponseTypeLogin:
		lenUsername := binary.LittleEndian.Uint16(data[1:3])
		username := string(data[3 : 3+lenUsername])
		msgBuffer.Append(fmt.Sprintf("\n\n\x1b[97;1m%s\x1b[90;22m has entered the dungeon!", username))
		drawMessageBuffer()
	case shared.ResponseTypeLogout:
		lenUsername := binary.LittleEndian.Uint16(data[1:3])
		username := string(data[3 : 3+lenUsername])
		msgBuffer.Append(fmt.Sprintf("\n\n\x1b[97;1m%s\x1b[90;22m has left the dungeon....", username))
		drawMessageBuffer()
	case shared.ResponseTypeLoggedInUsers:
		dataIdx := 1
		numUsers := binary.LittleEndian.Uint16(data[dataIdx:])
		dataIdx += 2
		usersWord := "users"
		if numUsers == 1 {
			usersWord = "user"
		}
		var whoStr strings.Builder
		fmt.Fprintf(&whoStr, "\n\n\x1b[33;1m%d\x1b[39:22m %s currently in the dungeon:", numUsers, usersWord)
		for range numUsers {
			lenUsername := int(binary.LittleEndian.Uint16(data[dataIdx:]))
			dataIdx += 2
			username := string(data[dataIdx : dataIdx+lenUsername])
			fmt.Fprintf(&whoStr, "\n  \x1b[97;1m%s\x1b[39;22m", username)
			dataIdx += lenUsername
		}
		msgBuffer.Append(whoStr.String())
		drawMessageBuffer()
	case shared.ResponseTypeLook:
		dataIdx := 1
		lenTitle := int(binary.LittleEndian.Uint16(data[dataIdx:]))
		dataIdx += 2
		title := string(data[dataIdx : dataIdx+lenTitle])
		dataIdx += lenTitle
		lenDescription := int(binary.LittleEndian.Uint16(data[dataIdx : dataIdx+2]))
		dataIdx += 2
		description := string(data[dataIdx : dataIdx+int(lenDescription)])
		dataIdx += lenDescription
		exitsRaw := data[dataIdx]
		dataIdx++
		exits := []string{}
		if exitsRaw&byte(shared.DirectionNorth) != 0 {
			exits = append(exits, "north")
		}
		if exitsRaw&byte(shared.DirectionEast) != 0 {
			exits = append(exits, "east")
		}
		if exitsRaw&byte(shared.DirectionSouth) != 0 {
			exits = append(exits, "south")
		}
		if exitsRaw&byte(shared.DirectionWest) != 0 {
			exits = append(exits, "west")
		}
		exitsTxt := strings.Join(exits, ", ")
		msgBuffer.Append(fmt.Sprintf("\n\n\x1b[37;1m%s\n\x1b[90;22m%s\n\nVisible exits are \x1b[37m%s\x1b[90;22m.", title, description, exitsTxt))
		drawMessageBuffer()
	case shared.ResponseTypeSay:
		lenUsername := binary.LittleEndian.Uint16(data[1:3])
		username := string(data[3 : 3+lenUsername])
		lenMsg := binary.LittleEndian.Uint16(data[3+lenUsername : 5+lenUsername])
		msg := string(data[5+lenUsername : 5+lenUsername+lenMsg])
		saysWord := "says"
		switch msg[len(msg)-1:] {
		case "!":
			saysWord = "exclaims"
		case "?":
			saysWord = "asks"
		}
		switch msg[len(msg)-2:] {
		case "!!":
			saysWord = "shouts"
		case "!?":
			saysWord = "demands"
		case ":O":
			saysWord = "sings"
		}
		msgBuffer.Append(fmt.Sprintf("\n\n\x1b[32;1m\"%s\"\x1b[90;22m, %s \x1b[37;1m%s\x1b[90;22m.", msg, saysWord, username))
		drawMessageBuffer()
	case shared.ResponseTypeCantMove:
		reason := shared.CantMoveReason(data[1])
		switch reason {
		case shared.CantMoveReasonNoExit:
			msgBuffer.Append("\n\n\x1b[31;22mAlas, ye cannot go that way!")
		case shared.CantMoveReasonTM:
			msgBuffer.Append("\n\n\x1b[31;22mAlas, due to thy torn meniscus, ye cannot move. Best take it easy on that leg.")
		}
		drawMessageBuffer()
	}
}

func (_ MudView) Update() {
	// ensure every user sees it every time; no race condition
	time.Sleep(150 * time.Millisecond)
	sendLookRequest()

	for {
		e := <-inputStreamSet.Input
		switch e := e.(type) {
		case input.KeyEvent:
			switch e.Key {
			default:
				switch e.Mod {
				case input.KeyModAlt:
					switch e.Key {
					case 'k':
						scrollMsgBufferUp()
					case 'j':
						scrollMsgBufferDown()
					}
				}
			}
		case input.NonalphaKeyEvent:
			switch e.Key {
			case input.NonalphaKeyUp:
				scrollMsgBufferUp()
			case input.NonalphaKeyDown:
				scrollMsgBufferDown()
			}
		case input.MouseEvent:
			switch e.Button {
			case input.MouseButtonWheelUp:
				scrollMsgBufferUp()
			case input.MouseButtonWheelDown:
				scrollMsgBufferDown()
			}
		}

		inputBuffer.Update(e)
		select {
		case txt := <-inputSubmit:
			command := strings.ToLower(strings.Split(txt, " ")[0])
			switch command {
			case "quit":
				inputStreamSet.Quit <- true
				return
			case "say":
				MudConnection.Write(append([]byte{shared.RequestTypeSay}, []byte(txt[4:]+"\n")...))
				if hintStep == "say" {
					hintStep = "look"
					drawHintText()
				}
			case "who":
				MudConnection.Write([]byte{shared.RequestTypeWho, '\n'})
				if hintStep == "who" {
					hintStep = "say"
					drawHintText()
				}
			case "look":
				sendLookRequest()
				if hintStep == "look" {
					hintStep = "done"
					drawHintText()
				}
			case "n", "north":
				MudConnection.Write([]byte{shared.RequestTypeMovement, byte(shared.DirectionNorth), '\n'})
			case "e", "east":
				MudConnection.Write([]byte{shared.RequestTypeMovement, byte(shared.DirectionEast), '\n'})
			case "s", "south":
				MudConnection.Write([]byte{shared.RequestTypeMovement, byte(shared.DirectionSouth), '\n'})
			case "w", "west":
				MudConnection.Write([]byte{shared.RequestTypeMovement, byte(shared.DirectionWest), '\n'})
			default:
				msgBuffer.Append(fmt.Sprintf("\n\n\x1b[37;1m%s\x1b[31;22m is not a recognized command!\x1b[90;22m", command))
				drawMessageBuffer()
			}
		default:
		}
	}
}

func sendLookRequest() {
	MudConnection.Write(append([]byte{shared.RequestTypeLook}, '\n'))
}

func (_ MudView) Render() {
	msgBufferWidth := min((TermSize.X - (MSG_BUFFER_PADDING_HOR * 2)), MSG_BUFFER_MAX_WIDTH)
	msgBufferX := ((TermSize.X / 2) - (msgBufferWidth / 2)) + 1
	msgBuffer.OnResize(
		shared.XY{X: msgBufferX, Y: 1},
		shared.XY{X: msgBufferWidth, Y: TermSize.Y - 5},
	)
	inputBuffer.OnResize(
		shared.XY{X: CURSOR_MIN_X_POS, Y: TermSize.Y - 1},
		shared.XY{X: TermSize.X - (CURSOR_MIN_X_POS + 1), Y: 1},
	)
	drawPromptInputArea()
	drawMessageBuffer()
	drawHintText()
	inputBuffer.Render()
}

func scrollMsgBufferUp() {
	didScroll := msgBuffer.ScrollUp()
	if didScroll {
		drawMessageBuffer()
	}
}
func scrollMsgBufferDown() {
	didScroll := msgBuffer.ScrollDown()
	if didScroll {
		drawMessageBuffer()
	}
}

func drawMessageBuffer() {
	msgBuffer.Render()
	inputBuffer.Render()
}

var hintStep = "who"

func drawHintText() {
	ansi.Set24BitBgCol(bgCol)
	ansi.Set24BitFgCol(shared.Color{R: 73, G: 64, B: 45})
	ansi.ResetIntensity()
	ansi.MoveCursorTo(2, TermSize.Y-3)

	if hintStep == "done" {
		fmt.Print("\x1b[K")
		return
	}

	isSmallSize := TermSize.X <= 80

	fmt.Print("Hint: enter ")
	ansi.SetFgCol(ansi.AnsiColorWhite, true)
	switch hintStep {
	case "who":
		fmt.Print("who")
		if isSmallSize {
			break
		}
		ansi.Set24BitFgCol(shared.Color{R: 73, G: 64, B: 45})
		fmt.Print(" to see which users are currently logged in")
	case "say":
		fmt.Print("say Hello my friends")
		if isSmallSize {
			break
		}
		ansi.Set24BitFgCol(shared.Color{R: 73, G: 64, B: 45})
		fmt.Print(" to say something to everyone else in the same room as you")
	case "look":
		fmt.Print("look")
		if isSmallSize {
			break
		}
		ansi.Set24BitFgCol(shared.Color{R: 73, G: 64, B: 45})
		fmt.Print(" to get a description of the room you're standing in")
	}
	fmt.Print("\x1b[K")

	inputBuffer.Render()
}

func drawPromptInputArea() {
	ansi.MoveCursorTo(1, TermSize.Y-2)
	ansi.ClearLineWithCol(TermSize.X, promptBgCol)
	fmt.Println()
	ansi.ClearLineWithCol(TermSize.X, promptBgCol)
	fmt.Println()
	ansi.ClearLineWithCol(TermSize.X, promptBgCol)
	ansi.MoveCursorTo(2, TermSize.Y-1)
	ansi.SetFgCol(ansi.AnsiColorYellow, false)
	fmt.Print(PROMPT_FLOWER)
}
