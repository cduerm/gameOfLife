package main

import (
	"flag"
	"math/rand"
	"strconv"
	"time"

	gol "github.com/cduerm/gameOfLife"
	ui "github.com/gizak/termui"
)

var game gol.Game
var stopchannel chan bool
var running bool = false
var editmode bool = false
var interval time.Duration = 200

var header, footer, board, debug *ui.Par
var keys *ui.List

var keyList = []string{
	"[General](bold)",
	"p - play/pause",
	"e - editmode on/off",
	"s - one step",
	"q - quit",
	"[Playing](bold)",
	"+/- - in-/decrease speed",
	"[Editmode](bold)",
	"c - clear field by 10 %",
	"f - fill field by 10 %",
	"cursor keys - move cursor",
	"space - toggle cell",
	"0 - periodic boundary",
	"1 - empty boundary",
	"2 - filled boundary",
}

func main() {
	var steps = flag.Int("steps", 1, "How many steps to calculate?")
	_ = steps
	var border = flag.Int("border", 0, "Border: 0 - periodic, 1 - empty, 2 - full")
	var size = flag.Int("size", 10, "Number of columns and rows.")
	var filling = flag.Float64("filling", 0.3, "Probability for a living cell in initial conditions.")
	flag.Parse()

	rand.Seed(int64(time.Now().Nanosecond()))
	game = *gol.New(*size, *size, gol.MakeRule([]int{3}), gol.MakeRule([]int{2, 3}), gol.BoundaryCondition(*border), float32(*filling))

	err := ui.Init()
	if err != nil {
		panic(err)
	}
	defer ui.Close()
	InitScreen()

	InstallHandlers()

	ui.Render(header, footer, board, keys, debug)
	ui.Loop()
}

func InstallHandlers() {
	ui.Handle("sys/kbd", func(ev ui.Event) {
		key := ev.Path[9:]
		switch key {
		case "q":
			ui.StopLoop()
		case "s":
			if !running {
				game.DoStep()
				board.Text = ToString(&game)
				ui.Render(board)
			}
		case "e":
			if editmode {
				debug.Text = "edit mode off"
				editmode = false
			} else {
				Stop()
				editmode = true
				debug.Text = "edit mode on"
			}
		case "<up>":
			debug.Text = "up key pressed"
		case "<space>":
			if editmode {
				debug.Text = "toggle"
			}
		case "c":
			if editmode {
				debug.Text = "clear"
			}
		case "f":
			if editmode {
				debug.Text = "fill"
			}
		case "p":
			debug.Text = "play/pause"
			if running {
				Stop()
			} else {
				editmode = false
				Start()
			}
		case "0", "1", "2":
			if editmode {
				number, _ := strconv.Atoi(key)
				bs := gol.BoundaryCondition(number)

				switch bs {
				case gol.BCEmpty:
					debug.Text = "empty boundary"
				case gol.BCFull:
					debug.Text = "full boundary"
				case gol.BCPeriodic:
					debug.Text = "periodic boundary"
				}
				game.SetBoundary(bs)
				board.Text = ToString(&game)
				ui.Render(board)
			}
		case "+":
			interval = time.Duration(float32(interval) / 1.6666)
			if interval < 10 {
				interval = 10
			}
			debug.Text = strconv.Itoa(int(interval))
		case "-":
			interval = time.Duration(float32(interval) * 1.6666)
			debug.Text = strconv.Itoa(int(interval))
		default:
			debug.Text = key
		}
		ui.Render(debug)
	})
}

func InitScreen() {
	header = ui.NewPar("")
	header.BorderLabel = "[Game of Life](fg-bold) in GO by Christoph Dürmann"
	header.Width = ui.TermWidth()
	header.Height = ui.TermHeight()

	footer = ui.NewPar("Controls: q - Exit  p - Pause  s - Step")
	footer.Width = ui.TermWidth()
	footer.Y = ui.TermHeight() - 3
	footer.Height = 0

	board = ui.NewPar(ToString(&game))
	board.Width = game.Cols()*2 + 4
	board.Height = game.Rows() + 4
	board.X = (ui.TermWidth() - 27 - board.Width) / 2
	board.Y = (ui.TermHeight() - board.Height) / 2
	board.Border = false

	keys = ui.NewList()
	keys.Width = 29
	keys.Height = len(keyList) + 2
	keys.Border = true
	keys.BorderLabel = "Hotkeys"
	keys.Items = keyList
	keys.Y = 1
	keys.X = ui.TermWidth() - keys.Width - 2

	debug = ui.NewPar("")
	debug.Width = keys.Width - 2
	debug.Height = 2
	debug.X = keys.X + 1
	debug.Y = keys.Y + keys.Height + 1
	debug.Border = false
}

func Stop() {
	if running {
		stopchannel <- true
		running = false
		debug.Text = "pause"
	}
}

func Start() {
	if !running {
		running = true
		stopchannel = Autorun(&game, board)
		debug.Text = "play"
	}
}

func Autorun(g *gol.Game, board *ui.Par) chan bool {
	stopchan := make(chan bool)
	go func() {
	runLoop:
		for {
			select {
			case <-stopchan:
				break runLoop
			default:
				g.DoStep()
				board.Text = ToString(g)
				ui.Render(board)
				time.Sleep(interval * time.Millisecond)
			}
		}
	}()
	return stopchan
}

func ToString(g *gol.Game) string {
	var border string
	switch g.Boundary() {
	case gol.BCPeriodic:
		border = "[#](bg-white,fg-black)"
	case gol.BCEmpty:
		border = "[.](bg-white,fg-black)"
	case gol.BCFull:
		border = "[x](bg-white,fg-black)"
	}
	var s string = ""
	//s += fmt.Sprintln("Step: ", g.step)
	for i := 0; i < game.Cols()+2; i++ {
		s += border + " "
	}
	s += "\n"
	for _, v := range g.Field() {
		s += border + " "
		for _, item := range v {
			if item {
				s += "x "
			} else {
				s += ". "
			}
		}
		s += border + "\n"
	}
	for i := 0; i < game.Cols()+2; i++ {
		s += border + " "
	}
	//s += fmt.Sprint("alive: ", g.alive)
	return s
}
