package input

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"dungeon/shared"
)

type StreamSet struct {
	Input     chan IEvent
	CursorPos chan IEvent
	Quit      chan bool
}

type EventKind uint

const (
	EventKindMouse EventKind = iota
	EventKindKey
)

type IEvent interface {
	implEvent()
}

type ImplEvent struct{}

func (_ ImplEvent) implEvent() {}

// mouse event
type MouseButton byte

const (
	MouseButtonLeft MouseButton = iota
	MouseButtonMiddle
	MouseButtonRight
)
const (
	MouseButtonWheelUp MouseButton = iota + 0x40
	MouseButtonWheelDown
	MouseButtonWheelLeft
	MouseButtonWheelRight
)

type MouseEvent struct {
	ImplEvent
	Button MouseButton
	Pos    shared.XY
}

// key event
type KeyMod uint

const (
	KeyModNone KeyMod = iota
	KeyModCtrl
	KeyModAlt
)

type KeyEvent struct {
	ImplEvent
	Key rune
	Mod KeyMod
}

// control key event
type NonalphaKey uint

const (
	NonalphaKeyUp NonalphaKey = iota
	NonalphaKeyDown
	NonalphaKeyLeft
	NonalphaKeyRight
)

type NonalphaKeyEvent struct {
	ImplEvent
	Key NonalphaKey
}

type CursorPosEvent struct {
	ImplEvent
	Pos shared.XY
}

func StartPollingInput(streamSet StreamSet) {
	for {
		buf := make([]byte, 64)
		n, err := os.Stdin.Read(buf)
		if err != nil {
			streamSet.Quit <- true
		}

		if n == 0 {
			continue
		}

		chunk := buf[:n]

		if chunk[0] == 3 {
			streamSet.Quit <- true
		}

		// DEBUG
		// var out strings.Builder
		// out.WriteString(" input:")
		// for _, byte := range chunk {
		// 	if byte < ' ' || byte >= 0x7f {
		// 		// out.WriteString(fmt.Sprintf("\\x%x", byte))
		// 		fmt.Fprintf(&out, "\\x%x", byte)
		// 	} else {
		// 		out.WriteString(string(byte))
		// 	}
		// }
		// os.Stdout.Write([]byte(out.String()))

		events := getEventsFromInput(chunk)
		for _, e := range events {
			switch e := e.(type) {
			case KeyEvent:
				if e.Key == 'C' && e.Mod == KeyModCtrl {
					streamSet.Quit <- true
				} else {
					streamSet.Input <- e
				}
			case MouseEvent:
				streamSet.Input <- e
			case NonalphaKeyEvent:
				streamSet.Input <- e
			case CursorPosEvent:
				streamSet.CursorPos <- e
			}
		}
	}
}

type ParserState uint

const (
	ParserStateInit ParserState = iota
	ParserStateEsc
	ParserStateCsi
)

// employs a state-machine parser heavily inspired by that of tcell's (see tcell's [`input.go`](https://github.com/gdamore/tcell/blob/v3.1.2/input.go#L440))
//
// this is essentially a stripped down version of that one
func getEventsFromInput(in []byte) []IEvent {
	var events []IEvent
	state := ParserStateInit
	csiParams := make([]byte, 0, 64)
	for _, b := range in {
		switch state {
		case ParserStateInit:
			switch b {
			case '\x1b':
				state = ParserStateEsc
			case '\t':
			case '\b':
				events = append(events, newKeyEvent(rune(b), KeyModNone))
			case '\r':
				events = append(events, newKeyEvent('\n', KeyModNone))
			case 0:
				events = append(events, newKeyEvent(' ', KeyModNone))
			default:
				if b < ' ' {
					events = append(events, newKeyEvent(rune(b+0x40), KeyModCtrl))
				} else if b == 0x7f {
					events = append(events, newKeyEvent('\b', KeyModNone))
				} else {
					events = append(events, newKeyEvent(rune(b), KeyModNone))
				}
			}
		case ParserStateEsc:
			switch b {
			case '[':
				state = ParserStateCsi
			default:
				events = append(events, newKeyEvent(rune(b), KeyModAlt))
				state = ParserStateInit
			}
		case ParserStateCsi:
			if b >= 0x30 && b <= 0x3f {
				csiParams = append(csiParams, b)
			} else if b >= 0x40 && b <= 0x7f {
				events = append(events, getEventFromCsi(b, csiParams))
				state = ParserStateInit
			} else {
				events = append(events, getNilEvent())
				state = ParserStateInit
			}
		}
	}
	if state == ParserStateEsc {
		events = append(events, newKeyEvent('\x1b', KeyModNone))
	}
	return events
}

func getEventFromCsi(mode byte, paramsRaw []byte) IEvent {
	pstr := paramsRaw
	var params []int
	hasLt := false
	if len(pstr) > 0 && pstr[0] == '<' {
		hasLt = true
		pstr = pstr[1:]
	}

	// create list of integer params
	if len(pstr) > 0 && pstr[0] >= '0' && pstr[0] <= '9' {
		parts := strings.SplitSeq(string(pstr), ";")
		for part := range parts {
			if part == "" {
				params = append(params, 0)
			} else {
				if n, e := strconv.ParseInt(part, 10, 32); e == nil {
					params = append(params, int(n))
				}
			}
		}
	}

	if hasLt {
		switch mode {
		case 'm', 'M':
			return getEventFromMouseInput(mode, params)
		}
	}
	switch mode {
	case 'A':
		return newNonalphaKeyEvent(NonalphaKeyUp)
	case 'B':
		return newNonalphaKeyEvent(NonalphaKeyDown)
	case 'C':
		return newNonalphaKeyEvent(NonalphaKeyRight)
	case 'D':
		return newNonalphaKeyEvent(NonalphaKeyLeft)
	case 'R':
		return newCursorPosEvent(params)
	}
	return getNilEvent()
}

func getEventFromMouseInput(mode byte, params []int) IEvent {
	if len(params) < 3 {
		return getNilEvent()
	}
	button := MouseButton(params[0] & 0b1100_0011)
	x := max(params[1], 0)
	y := max(params[2], 0)

	if mode == 'M' {
		return newMouseEvent(button, shared.XY{X: x, Y: y})
	}
	return getNilEvent()
}

func newCursorPosEvent(params []int) CursorPosEvent {
	return CursorPosEvent{
		ImplEvent{},
		shared.XY{X: params[1], Y: params[0]},
	}
}

func newKeyEvent(key rune, mod KeyMod) KeyEvent {
	return KeyEvent{
		ImplEvent{},
		key,
		mod,
	}
}
func newNonalphaKeyEvent(key NonalphaKey) NonalphaKeyEvent {
	return NonalphaKeyEvent{
		ImplEvent{},
		key,
	}
}
func newMouseEvent(button MouseButton, pos shared.XY) MouseEvent {
	return MouseEvent{
		ImplEvent{},
		button,
		pos,
	}
}

func getNilEvent() IEvent {
	return ImplEvent{}
}
