package bingo

import "math/rand"

// Game represents a bingo game.  The zero value can be used to start a new game.
type Game struct {
	numbers      [int(MaxNumber)]Number
	numbersDrawn int
}

// NumbersLeft reports how many available numbers in the game can be drawn.
func (g Game) NumbersLeft() int {
	if len(g.numbers) <= g.numbersDrawn {
		return 0
	}
	if g.numbers[0] == 0 {
		g.Reset()
	}
	return len(g.numbers) - g.numbersDrawn
}

// DrawnNumbers is the numbers in the game that have been drawn
func (g Game) DrawnNumbers() []Number {
	return g.numbers[:g.numbersDrawn]
}

// DrawNumber move the next available number to DrawnNumbers.
// The game is reset if no numbers are available or have been drawn.
func (g *Game) DrawNumber() {
	switch {
	case g.numbersDrawn <= 0:
		g.Reset()
		g.numbersDrawn = 1
	case g.numbersDrawn < len(g.numbers):
		g.numbersDrawn++
	}
}

// Reset clears drawn numbers and resets/shuffles all the possible available numbers.
func (g *Game) Reset() {
	for i := range g.numbers {
		g.numbers[i] = Number(i + 1)
	}
	rand.Shuffle(len(g.numbers), func(i, j int) {
		g.numbers[i], g.numbers[j] = g.numbers[j], g.numbers[i]
	})
	g.numbersDrawn = 0
}

// Columns partitions the drawn numbers by columns in the order that they were drawn.
func (g Game) Columns() map[int][]Number {
	cols := make(map[int][]Number, 5)
	for _, n := range g.DrawnNumbers() {
		cols[n.Column()] = append(cols[n.Column()], n)
	}
	return cols
}
