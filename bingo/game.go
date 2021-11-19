// Package bingo provides structures to simulate bingo games, boards, and number values.
package bingo

import (
	"errors"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

type (
	// Game represents a bingo game.  The zero value can be used to start a new game.
	Game struct {
		numbers      [MaxNumber - MinNumber + 1]Number
		numbersDrawn int
	}
	// Resetter resets games to valid, shuffled states.  It can be seeded to be predictable reset the next reset game.
	Resetter interface {
		// Reset resets the game.
		Reset(g *Game)
		// Seed sets the GameResetter to reset the next game from a starting point.
		Seed(seed int64)
	}
	// shuffler is the internal implementation of GameResetter.
	// It uses a random source to randomly swap numbers when shuffling.
	shuffler struct {
		*rand.Rand
		swap func(numbers []Number) func(i, j int)
	}
)

// GameResetter shuffles the game numbers.  It is seeded to the time it is created; it should only be used when testing.
var GameResetter Resetter = &shuffler{
	Rand: rand.New(rand.NewSource(time.Now().UnixNano())),
	swap: func(numbers []Number) func(i, j int) {
		return func(i, j int) {
			numbers[i], numbers[j] = numbers[j], numbers[i]
		}
	},
}

// NumbersLeft reports how many available numbers in the game can be drawn.
func (g Game) NumbersLeft() int {
	g.normalizeNumbersDrawn()
	return len(g.numbers) - g.numbersDrawn
}

// DrawnNumbers is the numbers in the game that have been drawn
func (g Game) DrawnNumbers() []Number {
	g.normalizeNumbersDrawn()
	return g.numbers[:g.numbersDrawn]
}

// DrawNumber move the next available number to DrawnNumbers.
// The game is reset if no numbers have been drawn.
func (g *Game) DrawNumber() {
	g.normalizeNumbersDrawn()
	switch {
	case g.numbersDrawn == 0:
		GameResetter.Reset(g)
		g.numbersDrawn = 1
	case g.numbersDrawn < len(g.numbers):
		g.numbersDrawn++
	}
}

// DrawnNumberColumns partitions the drawn numbers by columns in the order that they were drawn.
func (g Game) DrawnNumberColumns() map[int][]Number {
	cols := make(map[int][]Number, 5)
	drawnNumbers := g.DrawnNumbers()
	for _, n := range drawnNumbers {
		c := n.Column()
		cols[c] = append(cols[c], n)
	}
	return cols
}

// PreviousNumberDrawn is the last number drawn, or 0 of no numbers have been drawn.
func (g Game) PreviousNumberDrawn() Number {
	g.normalizeNumbersDrawn()
	if g.numbersDrawn == 0 {
		return 0
	}
	return g.numbers[g.numbersDrawn-1]
}

// Reset clears drawn numbers and resets/shuffles all the possible available numbers.
// To shuffle the numbers to a specific order, call rand.Seed() with a constant value.
func (s *shuffler) Reset(g *Game) {
	for i := range g.numbers {
		g.numbers[i] = Number(i + 1)
	}
	s.Rand.Shuffle(len(g.numbers), func(i, j int) {
		g.numbers[i], g.numbers[j] = g.numbers[j], g.numbers[i]
	})
	g.numbersDrawn = 0
}

// normalizeNumbersDrawn clamps numbersDrawn to [0,75]
func (g *Game) normalizeNumbersDrawn() {
	switch {
	case g.numbersDrawn < 0:
		g.numbersDrawn = 0
	case g.numbersDrawn > len(g.numbers):
		g.numbersDrawn = len(g.numbers)
	}
}

// ID encodes the game into an easy to transport string.
func (g Game) ID() (string, error) {
	g.normalizeNumbersDrawn()
	switch {
	case g.numbersDrawn == 0:
		return "0", nil
	case !validNumbers(g.numbers[:], false):
		return "", errors.New("game has duplicate/invalid numbers")
	}
	data := make([]byte, len(g.numbers))
	for i, n := range g.numbers {
		data[i] = byte(n)
	}
	nums := base64Encoding.EncodeToString(data)
	id := strconv.Itoa(g.numbersDrawn) + "-" + nums
	return id, nil
}

// GameFromID creates a game from the identifying string.
func GameFromID(id string) (*Game, error) {
	i := strings.IndexAny(id, "-")
	switch {
	case id == "0":
		return new(Game), nil
	case i < 0, i >= len(id):
		return nil, errors.New("could not split id string into numbersDrawn and numbers")
	}
	numbersDrawnStr, numsStr := id[:i], id[i+1:]
	numbersDrawn, err := strconv.Atoi(numbersDrawnStr)
	if err != nil {
		return nil, errors.New("parsing numbersLeft: " + err.Error())
	}
	data, err := base64Encoding.DecodeString(numsStr)
	if err != nil {
		return nil, errors.New("decoding game numbers: " + err.Error())
	}
	var g Game
	if len(data) != len(g.numbers) {
		return nil, errors.New("decoded numbers too large/small")
	}
	for i, n := range data {
		g.numbers[i] = Number(n)
	}
	if !validNumbers(g.numbers[:], false) {
		return nil, errors.New("game has duplicate/invalid numbers")
	}
	g.numbersDrawn = numbersDrawn
	return &g, nil
}
