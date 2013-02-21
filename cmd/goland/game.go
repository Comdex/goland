package main

import (
	"flag"
	"fmt"
	"github.com/nsf/termbox-go"
	"github.com/nsf/tulib"
	"image"
	"log"
	"os"
	"runtime"
	"time"
	"unicode"
)

const (
	NumFPSSamples = 64
	FPSLimit      = 60
)

var (
	fpsSamples    [64]float64
	currentSample = 0
	stats         runtime.MemStats
	logfile       = flag.String("log", "goland.log", "log file")
	debug         = flag.Bool("debug", false, "print debugging info")
)

type Game struct {
	P *Player

	Terminal
	logfile *os.File

	CloseChan chan bool

	// unexported
	fps float64

	Objects []Object

	Map *MapChunk
}

func NewGame() *Game {
	g := Game{}

	g.CloseChan = make(chan bool, 1)

	g.Map = NewMapChunk()
	g.Map.Locations[1][4] = GlyphToTerrain('#')
	g.Map.Locations[2][4] = GlyphToTerrain('#')
	g.Map.Locations[3][4] = GlyphToTerrain('#')
	g.Map.Locations[4][4] = GlyphToTerrain('#')
	g.Map.Locations[4][3] = GlyphToTerrain('#')
	g.Map.Locations[4][2] = GlyphToTerrain('#')
	g.Map.Locations[4][1] = GlyphToTerrain('#')

	g.P = NewPlayer(&g)
	g.P.Pos = image.Pt(2, 2)

	g.Objects = append(g.Objects, g.P)

	u := NewUnit(&g)
	u.Ch.Ch = '@'
	u.Pos = image.Pt(7, 7)

	g.Objects = append(g.Objects, &u)

	return &g
}

func (g *Game) Run() {

	g.Start()

	timer := NewDeltaTimer()
	ticker := time.NewTicker(time.Second / FPSLimit)

	run := true

	for run {
		select {
		case <-ticker.C:
			// frame tick
			delta := timer.DeltaTime()

			if delta.Seconds() > 0.25 {
				delta = time.Duration(250 * time.Millisecond)
			}

			g.Update(delta)
			g.Draw()

			g.Flush()

		case <-g.CloseChan:
			run = false
		}
	}

	g.End()

}

func (g *Game) Start() {
	f, err := os.OpenFile(*logfile, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		log.Fatal(err)
	}

	log.SetOutput(f)
	log.Print("Logging started")

	g.logfile = f

	g.Terminal.Start()

	g.HandleKey(termbox.KeyEsc, func(ev termbox.Event) { g.CloseChan <- false })

	scale := PLAYER_RUN_SPEED

	// convert to func SetupDirections()
	for k, v := range CARDINALS {
		func(c rune, d Direction) {
			g.HandleRune(c, func(_ termbox.Event) {
				g.P.Move(d)
			})

			upperc := unicode.ToUpper(c)
			g.HandleRune(upperc, func(_ termbox.Event) {
				for i := 0; i < scale; i++ {
					g.P.Move(d)
				}
			})
		}(k, v)
	}
}

func (g *Game) End() {
	log.Print("Logging ended")
	g.logfile.Close()
	g.Terminal.End()
}

func (g *Game) Update(delta time.Duration) {
	// update fps
	g.fps = g.calcFPS(delta)
	runtime.ReadMemStats(&stats)

	g.RunInputHandlers()

	for _, o := range g.Objects {
		o.Update(delta)
	}

}

func (g *Game) Draw() {

	// construct a current view of the 2d world and blit it
	viewwidth := g.Terminal.Rect.Width - VIEW_START_X - VIEW_PAD_X
	viewheight := g.Terminal.Rect.Height - VIEW_START_Y - VIEW_PAD_Y
	viewrect := tulib.Rect{VIEW_START_X, VIEW_START_Y, viewwidth, viewheight}
	viewbuf := tulib.NewBuffer(viewwidth, viewheight)
	viewbuf.Fill(viewrect, termbox.Cell{Ch: ' ', Fg: termbox.ColorDefault, Bg: termbox.ColorDefault})

	cam := NewCamera(viewbuf)
	cam.SetCenter(g.P.GetPos())

	// draw terrain
	for y, row := range g.Map.Locations {
		for x, terr := range row {
			pos := image.Pt(x, y)
			if cam.ContainsWorldPoint(pos) {
				cam.Draw(terr, pos)
			}
		}
	}

	// draw other crap
	for _, o := range g.Objects {
		if cam.ContainsWorldPoint(o.GetPos()) {
			cam.Draw(o, o.GetPos())
		}
	}

	// draw labels
	statsparams := &tulib.LabelParams{termbox.ColorRed, termbox.ColorBlack, tulib.AlignLeft, '.', false}
	statsrect := tulib.Rect{1, 0, 60, 1}

	statsstr := fmt.Sprintf("%dx%d TERM %5.2f FPS %5.2f MB %d GC %d GR",
		g.Terminal.Rect.Width, g.Terminal.Rect.Height, g.fps, float64(stats.HeapAlloc)/1000000.0, stats.NumGC, runtime.NumGoroutine())

	playerparams := &tulib.LabelParams{termbox.ColorRed, termbox.ColorBlack, tulib.AlignLeft, '.', false}
	playerrect := tulib.Rect{1, g.Terminal.Rect.Height - 1, g.Terminal.Rect.Width, 1}

	playerstr := fmt.Sprintf("%s Cam.Pos: %s Cam.Rect: %v", g.P, cam.Pos.String(), cam.Rect)

	g.Terminal.DrawLabel(statsrect, statsparams, []byte(statsstr))
	g.Terminal.DrawLabel(playerrect, playerparams, []byte(playerstr))

	// blit
	g.Terminal.Blit(viewrect, 0, 0, &viewbuf)

}

func (g *Game) calcFPS(delta time.Duration) float64 {
	fpsSamples[currentSample%NumFPSSamples] = 1.0 / delta.Seconds()
	currentSample++
	fps := 0.0

	for i := 0; i < NumFPSSamples; i++ {
		fps += fpsSamples[i]
	}

	fps /= NumFPSSamples

	return fps
}
