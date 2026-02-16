package main

import (
	"encoding/binary"
	"fmt"
	"math"
	"strings"

	"dungeon/client/ansi"
	"dungeon/client/buffer"
	"dungeon/client/input"
	"dungeon/shared"
)

var promptBarCol = shared.Color{R: 24, G: 24, B: 24}

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
		promptBarCol,
		inputSubmit,
	)
}

func listenForMessages() {
	// this will allocate 65,536 bytes of RAM per user. My raspberry pi has 16GB of RAM total, 14GB roughly it says is free. This means about 229,376 users could be online at a time.
	// even the most popular MUDs in the world in 2026 struggle to hit 100 users online at a time so this is plenty
	dataBuf := make([]byte, math.MaxUint16+1)
	for {
		n, err := MudConnection.Read(dataBuf)
		lenData := binary.LittleEndian.Uint16(dataBuf)
		if err == nil && int(lenData) == n-2 {
			data := dataBuf[2:n]
			switch data[0] {
			case shared.ResponseTypeLogin:
				lenUsername := binary.LittleEndian.Uint16(data[1:3])
				username := string(data[3 : 3+lenUsername])
				msgBuffer.Append(fmt.Sprintf("\n\n\x1b[97;1m%s\x1b[39;22m has entered the dungeon!", username))
				drawMessageBuffer()
			case shared.ResponseTypeLogout:
				lenUsername := binary.LittleEndian.Uint16(data[1:3])
				username := string(data[3 : 3+lenUsername])
				msgBuffer.Append(fmt.Sprintf("\n\n\x1b[97;1m%s\x1b[39;22m has left the dungeon....", username))
				drawMessageBuffer()
			case shared.ResponseTypeLoggedInUsers:
				ind := 1
				numUsers := binary.LittleEndian.Uint16(data[ind:])
				ind += 2
				usersWord := "users"
				if numUsers == 1 {
					usersWord = "user"
				}
				var whoStr strings.Builder
				fmt.Fprintf(&whoStr, "\n\n\x1b[33;1m%d\x1b[39:22m %s currently in the dungeon:", numUsers, usersWord)
				for range numUsers {
					lenUsername := int(binary.LittleEndian.Uint16(data[ind:]))
					ind += 2
					username := string(data[ind : ind+lenUsername])
					fmt.Fprintf(&whoStr, "\n  \x1b[97;1m%s\x1b[39;22m", username)
					ind += lenUsername
				}
				msgBuffer.Append(whoStr.String())
				drawMessageBuffer()
			case shared.ResponseTypeLook:
				lenTitle := binary.LittleEndian.Uint16(data[1:])
				title := string(data[3 : 3+lenTitle])
				lenDescription := binary.LittleEndian.Uint16(data[3+lenTitle:])
				description := string(data[5+lenTitle : 5+lenTitle+lenDescription])
				msgBuffer.Append(fmt.Sprintf("\n\n\x1b[37;1m%s\n\x1b[90;22m%s", title, description))
				drawMessageBuffer()
			case shared.ResponseTypeSay:
				lenUsername := binary.LittleEndian.Uint16(data[1:3])
				username := string(data[3 : 3+lenUsername])
				lenMsg := binary.LittleEndian.Uint16(data[3+lenUsername : 5+lenUsername])
				msg := string(data[5+lenUsername : 5+lenUsername+lenMsg])
				saysWord := "says"
				switch msg[len(msg)-1:] {
				case "!":
					saysWord = "shouts"
				case "?":
					saysWord = "asks"
				}
				msgBuffer.Append(fmt.Sprintf("\n\n\x1b[32;1m\"%s\"\x1b[90;22m, %s \x1b[37;1m%s\x1b[90;22m.", msg, saysWord, username))
				drawMessageBuffer()
			}
		} else if err != nil {
			fmt.Printf("server disconnected!! %v\n", err)
			inputStreamSet.Quit <- true
			return
		}
	}

}

func (_ MudView) Update() {
	go listenForMessages()

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
			command := strings.Split(txt, " ")[0]
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
				MudConnection.Write(append([]byte{shared.RequestTypeWho}, '\n'))
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

	fmt.Print("Hint: enter ")
	ansi.SetFgCol(ansi.AnsiColorWhite, true)
	switch hintStep {
	case "who":
		fmt.Print("who")
		ansi.Set24BitFgCol(shared.Color{R: 73, G: 64, B: 45})
		fmt.Print(" to see which users are currently logged in")
	case "say":
		fmt.Print("say Hello my friends")
		ansi.Set24BitFgCol(shared.Color{R: 73, G: 64, B: 45})
		fmt.Print(" to say something to everyone else in the same room as you")
	case "look":
		fmt.Print("look")
		ansi.Set24BitFgCol(shared.Color{R: 73, G: 64, B: 45})
		fmt.Print(" to get a description of the room you're standing in")
	}
	fmt.Print("\x1b[K")

	inputBuffer.Render()
}

func drawPromptInputArea() {
	ansi.MoveCursorTo(1, TermSize.Y-2)
	ansi.ClearLineWithCol(TermSize.X, promptBarCol)
	fmt.Println()
	ansi.ClearLineWithCol(TermSize.X, promptBarCol)
	fmt.Println()
	ansi.ClearLineWithCol(TermSize.X, promptBarCol)
	ansi.MoveCursorTo(2, TermSize.Y-1)
	ansi.SetFgCol(ansi.AnsiColorYellow, false)
	fmt.Print(PROMPT_FLOWER)
}
