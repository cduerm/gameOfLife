// Package gameOfLife contains all ingredients to run a Game of Life with
// arbitrary board size and rules. Full, empty and periodic boundary
// conditions are supported.
package gameOfLife

import "math/rand"

// BoundaryCondition holds the 3 possible bounday conditons
type BoundaryCondition uint8

// All possbile boundary conditions
const (
	BCPeriodic BoundaryCondition = iota
	BCEmpty
	BCFull
)

// Gol holds all data of a Game of Life.
type Gol struct {
	field                [][]bool
	row, col             int
	step                 int
	alive                int
	ruleCreate, ruleLive GolRule
	boundaryCondition    BoundaryCondition
}

// GolRule contains the number of neighbours for which a Cell
// stays alive or is newly born.
type GolRule [10]bool

// Boundary retrieves the boundary conditon
func (g *Gol) Boundary() BoundaryCondition {
	return g.boundaryCondition
}

// SetBoundary sets the boundary contition
func (g *Gol) SetBoundary(b BoundaryCondition) {
	g.boundaryCondition = b
}

// GetSize returns the size of the field (withut boundaries)
func (g *Gol) GetSize() (w, h int) {
	return g.col, g.row
}

// NewGol initializes a new Gol structure with a specified size, rules to
// create and keep living cells, the boundary conditions and a filling fraction
// p which gives the chance for each cell to be alive in the beginning.
func NewGol(row, col int, ruleCreate, ruleLive GolRule, borderStyle BoundaryCondition, p float32) *Gol {
	g := new(Gol)
	g.row = row
	g.col = col
	g.ruleCreate = ruleCreate
	g.ruleLive = ruleLive
	g.boundaryCondition = borderStyle
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
	return g
}

// MakeRule creates a live or born rule from a slice of integers.
// With a live rule []int{2,3} a cell would survive with either 2 or 3
// neighbours.
func MakeRule(i []int) (rule GolRule) {
	for _, val := range i {
		rule[val] = true
	}
	return
}

// String returns a string representation of the field to implement the
// fmt.Stringer interface
func (g *Gol) String() string {
	var s string
	for i := -1; i < g.col+1; i++ {
		for j := -1; j < g.row+1; j++ {
			if g.Cell(i, j) {
				s += "x"
			} else {
				s += "."
			}
			if j < g.col {
				s += " "
			}
		}
		s += "\n"
	}
	s += "\n"
	return s
}

// Cell returns the state of one cell, considering boundary conditions
func (g *Gol) Cell(i, j int) bool {
	if i >= 0 && i < g.row && j >= 0 && j < g.col {
		return g.field[i][j]
	}

	switch g.boundaryCondition {
	case BCEmpty:
		return false
	case BCFull:
		return true
	case BCPeriodic:
		return g.field[(i+g.row)%g.row][(j+g.col)%g.col]
	}

	return false
}

// Field returns the slice containing all cells, but not the boundary
func (g *Gol) Field() [][]bool {
	return g.field
}

func (g *Gol) neighbours(i, j int) int {
	n := 0
	if g.Cell(i, j) {
		n--
	}

	for i1 := -1; i1 < 2; i1++ {
		for j1 := -1; j1 < 2; j1++ {
			if g.Cell(i+i1, j+j1) {
				n++
			}
		}
	}
	return n
}

// DoStep evolves the board by one step
func (g *Gol) DoStep() {
	var iField [][]bool
	iField = make([][]bool, g.row)
	for i := range iField {
		iField[i] = make([]bool, g.col)
	}

	g.step++

	for i := 0; i < g.row; i++ {
		for j := 0; j < g.col; j++ {
			neighbours := g.neighbours(i, j)

			if g.Cell(i, j) && !g.ruleLive[neighbours] {
				iField[i][j] = false
				g.alive--
			} else if !g.Cell(i, j) && g.ruleCreate[neighbours] {
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
