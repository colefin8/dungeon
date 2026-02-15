package main

import (
	"fmt"

	"dungeon/client/ansi"
	"dungeon/client/buffer"
	"dungeon/client/input"
	"dungeon/shared"
)

var promptBarCol = shared.Color{R: 24, G: 24, B: 24}

const CURSOR_MIN_X_POS = 5
const DBG_TXT = `
You've earned this.

Yeah!

I'm a Pollen Jock! And it's a perfect
fit. All I gotta do are the sleeves.

Oh, yeah.

That's our Barry.

Mom! The bees are back!

If anybody needs
to make a call, now's the time.

I got a feeling we'll be
working late tonight!

Here's your change. Have a great
afternoon! Oan I help who's next?

Would you like some honey with that?
It is bee-approved. Don't forget these.

Milk, cream, cheese, it's all me.
And I don't see a nickel!

Sometimes I just feel
like a piece of meat!

I had no idea.

Barry, I'm sorry.
Have you got a moment?

Would you excuse me?
My mosquito associate will help you.

Sorry I'm late.

He's a lawyer too?

I was already a blood-sucking parasite.
All I needed was a briefcase.

Have a great afternoon!

Barry, I just got this huge tulip order,
and I can't get them anywhere.

No problem, Vannie.
Just leave it to me.

You're a lifesaver, Barry.
Can I help who's next?

All right, scramble, jocks!
It's time to fly.

Thank you, Barry!

That bee is living my life!

Let it go, Kenny.

- When will this nightmare end?!
- Let it all go.

- Beautiful day to fly.
- Sure is.

Between you and me,
I was dying to get out of that office.

You have got
to start thinking bee, my friend.

- Thinking bee!
- Me?

Hold it. Let's just stop for a second. Hold it. I'm sorry. I'm sorry, everyone. Can we stop here? I'm not making a major life decision during a production number! All right. Take ten, everybody. Wrap it up, guys. I had virtually no rehearsal for that.`

const PROMPT_FLOWER = "\u2767"
const MSG_BUFFER_HOR_PADDING = 2

var msgBuffer buffer.TextBuffer
var inputBuffer buffer.InputBuffer
var inputSubmit = make(chan string, 512)

type MudView struct{}

func (_ MudView) Init() {
	msgBuffer = buffer.NewTextBuffer(DBG_TXT)
	inputBuffer = buffer.NewInputBuffer(
		promptBarCol,
		inputSubmit,
	)
}

func (_ MudView) Update() {
	go func() {
		for {
			txt := <-inputSubmit
			ansi.MoveCursorTo(5, 5)
			fmt.Print(txt)
			inputBuffer.Render()
		}
	}()

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
	}
}

func (_ MudView) Render() {
	msgBuffer.OnResize(shared.XY{X: TermSize.X - (MSG_BUFFER_HOR_PADDING * 2), Y: TermSize.Y - 4})
	inputBuffer.OnResize(
		shared.XY{X: CURSOR_MIN_X_POS, Y: TermSize.Y - 1},
		shared.XY{X: TermSize.X - (CURSOR_MIN_X_POS + 1), Y: 1},
	)
	drawMessageBuffer()
	drawPromptInputArea()
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
	msgBufferVisibleLines, msgBufferVisibleTextYPos := msgBuffer.GetVisibleTextAndYPos(TermSize.X-(MSG_BUFFER_HOR_PADDING*2), TermSize.Y-4)
	ansi.MoveCursorTo(MSG_BUFFER_HOR_PADDING+1, msgBufferVisibleTextYPos)
	ansi.SetFgCol(ansi.AnsiColorWhite, false)
	ansi.Set24BitBgCol(bgCol)
	ansi.HideCursor()
	fmt.Print(msgBufferVisibleLines)
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
