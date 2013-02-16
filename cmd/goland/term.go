package main

import (
  "fmt"
  "strings"
  "os"
  "log"
  "github.com/nsf/termbox-go"
  "github.com/nsf/tulib"
)

const (
  superficialSizeLimit = 24
  border = "#"
)

type KeyHandler func(ev termbox.Event)

type Terminal struct {
  tulib.Buffer
  EventChan     chan termbox.Event

  runehandlers  map[rune] KeyHandler
  keyhandlers   map[termbox.Key] KeyHandler
}

func (t *Terminal) Start() error {
  err := termbox.Init()
  if err != nil {
    panic(err)
  }

  t.Buffer = tulib.TermboxBuffer()

  if t.Rect.Height < superficialSizeLimit {
    fmt.Println("terminal too small")
    t.End()
    os.Exit(1)
  }

  t.EventChan = make(chan termbox.Event)

  // event generator
  go func(e chan termbox.Event) {
    for {
      e <- termbox.PollEvent()
    }
  }(t.EventChan)

  t.runehandlers  = make(map[rune] KeyHandler)
  t.keyhandlers   = make(map[termbox.Key] KeyHandler)

  return nil
}

func (t *Terminal) End() {
  termbox.Close()
}

func (t *Terminal) Draw() {
  termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
  t.drawborder()
}

func (t *Terminal) Flush() {
  termbox.Flush()
}

func (t *Terminal) RunInputHandlers() error {
  select {
  case ev := <- t.EventChan:
    log.Printf("Keypress: %s", tulib.KeyToString(ev.Key, ev.Ch, ev.Mod))

    if ev.Ch != 0 { // this is a character
      if handler, ok := t.runehandlers[ev.Ch]; ok {
        handler(ev)
      }
    } else {
      if handler, ok := t.keyhandlers[ev.Key]; ok {
        handler(ev)
      }
    }

  default:
  }

  return nil
}

func (t *Terminal) HandleRune(r rune, h KeyHandler) {
  t.runehandlers[r] = h
}

func (t *Terminal) HandleKey(k termbox.Key, h KeyHandler) {
  t.keyhandlers[k] = h
}

func (t *Terminal) Print(x, y int, fg, bg termbox.Attribute, msg string) {
  for _, c := range msg {
		termbox.SetCell(x, y, c, fg, bg)
		x++
	}
}

func (t *Terminal) Printf(x, y int, fg, bg termbox.Attribute, format string, args ...interface{}) {
	s := fmt.Sprintf(format, args...)
	t.Print(x, y, fg, bg, s)
}

func (t *Terminal) drawborder() {
  var y int

  t.Print(0, 0, termbox.ColorWhite, termbox.ColorBlack, strings.Repeat(border, t.Width))

  for y = 0; y < t.Height - 1; y++ {
    t.Print(0, y, termbox.ColorWhite, termbox.ColorBlack, border)
    t.Print(t.Width - 1, y, termbox.ColorWhite, termbox.ColorBlack, border)
  }

  t.Print(0, y, termbox.ColorWhite, termbox.ColorBlack, strings.Repeat(border, t.Width))

}

