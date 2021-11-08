package bingo

import "math/rand"

type Game struct {
	DrawnNumbers     []Number
	availableNumbers []Number
}

// NumbersLeft reports how many available numbers in the game can be drawn.
func (g *Game) NumbersLeft() int {
	return len(g.availableNumbers)
}

// DrawNumber move the next available number to DrawnNumbers.
// The game is reset if no numbers are available or have been drawn.
func (g *Game) DrawNumber() {
	if len(g.availableNumbers) == 0 {
		if len(g.DrawnNumbers) != 0 {
			return
		}
		g.Reset()
	}
	g.DrawnNumbers = append(g.DrawnNumbers, g.availableNumbers[0])
	g.availableNumbers = g.availableNumbers[1:]
}

// reset clears drawn numbers and
func (g *Game) Reset() {
	const c = int(MaxNumber - MinNumber + 1)
	arr := make([]Number, c)
	for i := 0; i < c; i++ {
		arr[i] = Number(i + 1)
	}
	rand.Shuffle(c, func(i, j int) {
		arr[i], arr[j] = arr[j], arr[i]
	})
	g.availableNumbers = arr
	g.DrawnNumbers = nil
}
