package main

import (
	"flag"
	"fmt"
	"math/rand"
	"time"
	"strconv"
	ui "github.com/gizak/termui"
	gol "github.com/cduerm/gameOfLife"
)

// Randbedingungen
const (
	periodic = 0
	empty    = 1
	full     = 2
)

type gol struct {
	field                [][]bool
	row, col             int
	step                 int
	alive                int
	ruleCreate, ruleLive [10]bool
	borderStyle          int
}

var game gol
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
	game.Init(*size, *size, MakeRule([]int{3}), MakeRule([]int{2, 3}), *border, float32(*filling))

	err := ui.Init()
	if err != nil {
		panic(err)
	}
	defer ui.Close()
	InitScreen()

	InstallHandler()

	ui.Render(header, footer, board, keys, debug)
	ui.Loop()
}

func InstallHandler() {
	ui.Handle("sys/kbd", func(ev ui.Event) {
		key := ev.Path[9:]
		switch key {
		case "q":
			ui.StopLoop()
		case "s":
			if !running {
				game.Evolution()
				board.Text = game.ToString()
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
			if editmode{
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
				bs, _ := strconv.Atoi(key)
				switch bs {
				case empty:
					debug.Text = "empty boundary"
				case full:
					debug.Text = "full boundary"
				case periodic:
					debug.Text = "periodic boundary"
				}
				game.borderStyle = bs
				board.Text = game.ToString()
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

	board = ui.NewPar(game.ToString())
	board.Width = game.col * 2 + 4
	board.Height = game.row + 4
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

func Autorun(g *gol, board *ui.Par) (chan bool) {
	stopchan := make(chan bool)
	go func () {
		runLoop:
		for {
			select {
			case <-stopchan:
				break runLoop
			default:
				g.Evolution()
				board.Text = g.ToString()
				ui.Render(board)
				time.Sleep(interval*time.Millisecond)
			}
		}
	}()
	return stopchan
}

func (g *gol) Init(row, col int, ruleCreate, ruleLive [10]bool, borderStyle int, p float32) {
	g.row = row
	g.col = col
	g.ruleCreate = ruleCreate
	g.ruleLive = ruleLive
	g.borderStyle = borderStyle
	g.step = 0

	g.field = make([][]bool, g.row)
	for r := range g.field {
		g.field[r] = make([]bool, col)
		for c := range g.field[r] {
			if rand.Float32() < p {
				g.field[r][c] = true
				g.alive++
			}
		}
	}
}

func MakeRule(i []int) (rule [10]bool) {
	for _, val := range i {
		rule[val] = true
	}
	return
}

func (g *gol) Print() {
	//fmt.Println("Step: ", g.step)
	for _, v := range g.field {
		for _, item := range v {
			if item {
				fmt.Print("x ")
			} else {
				fmt.Print(". ")
			}
		}
		fmt.Print("\n")
	}
	//fmt.Print("alive: ", g.alive, "\n\n")
}

func (g *gol) ToString() string {
	var border string
	switch g.borderStyle {
		case periodic:
			border = "[#](bg-white,fg-black)"
		case empty:
			border = "[.](bg-white,fg-black)"
		case full:
			border = "[x](bg-white,fg-black)"
		}
	var s string = ""
	//s += fmt.Sprintln("Step: ", g.step)
	for i := 0 ; i < game.col + 2 ; i++ {
		s += border + " "
	}
	s += "\n"
	for _, v := range g.field {
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
	for i := 0 ; i < game.col + 2 ; i++ {
		s += border + " "
	}
	//s += fmt.Sprint("alive: ", g.alive)
	return s
}

func (g *gol) Cell(i, j int) bool {
	if i >= 0 && i < g.row && j >= 0 && j < g.col {
		return g.field[i][j]
	}

	switch g.borderStyle {
	case empty:
		return false
	case full:
		return true
	case periodic:
		return g.field[(i+g.row)%g.row][(j+g.col)%g.col]
	}

	return false
}

func (g *gol) Evolution() {
	var neighbours int
	var iField [][]bool
	iField = make([][]bool, g.row)
	for i := range iField {
		iField[i] = make([]bool, g.col)
	}

	g.step++

	for i := 0; i < g.row; i++ {
		for j := 0; j < g.col; j++ {
			neighbours = 0
			if g.Cell(i, j) {
				neighbours--
			}

			for i1 := -1; i1 < 2; i1++ {
				for j1 := -1; j1 < 2; j1++ {
					if g.Cell(i+i1, j+j1) {
						neighbours++
					}
				}
			}
			if g.Cell(i, j) && !g.ruleLive[neighbours] {
				// fmt.Println("die:  ",i+1,j+1,neighbours)
				iField[i][j] = false
				g.alive--
			} else if !g.Cell(i, j) && g.ruleCreate[neighbours] {
				// fmt.Println("born: ",i+1,j+1,neighbours)
				iField[i][j] = true
				g.alive++
			} else {
				iField[i][j] = g.Cell(i, j)
			}
		}
	}

	for i := 0; i < g.row; i++ {
		for j := 0; j < g.col; j++ {
			g.field[i][j] = iField[i][j]
		}
	}
}
