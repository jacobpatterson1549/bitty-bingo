package bingo

import (
	"errors"
	"math/rand"
	"strconv"
	"strings"
)

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

func (g Game) ID() (string, error) {
	if g.numbersDrawn <= 0 {
		return "0", nil
	}
	if !g.validNumbers() {
		return "", errors.New("game numbers not valid")
	}
	// TODO: make byte array smaller, numbers should only range [0,75)
	// this means each number takes 7 digits: ceil(log2(75)) digits = ceil(6.2288)
	// this means only 7*75=525 bits are needed, and 525/8 = 65 5/8, so only 66 bytes are needed
	b := make([]byte, len(g.numbers))
	for i, n := range g.numbers {
		b[i] = byte(n)
	}
	nums := base64Encoding.EncodeToString(b)
	id := strconv.Itoa(g.numbersDrawn) + "-" + nums
	return id, nil
}

func GameFromID(id string) (*Game, error) {
	if id == "0" {
		return new(Game), nil
	}
	i := strings.IndexAny(id, "-")
	if i < 0 || i >= len(id) {
		return nil, errors.New("could not split id string into numbersDrawn and numbers")
	}
	numbersDrawnStr, numsStr := id[:i], id[i+1:]
	numbersDrawn, err := strconv.Atoi(numbersDrawnStr)
	if err != nil {
		return nil, errors.New("parsing numbersLeft: " + err.Error())
	}
	b, err := base64Encoding.DecodeString(numsStr)
	if err != nil {
		return nil, errors.New("decoding game numbers: " + err.Error())
	}
	var g Game
	if len(b) != len(g.numbers) {
		return nil, errors.New("decoded numbers too large/small")
	}
	for i, n := range b {
		g.numbers[i] = Number(n)
	}
	if !g.validNumbers() {
		return nil, errors.New("game numbers not valid")
	}
	g.numbersDrawn = numbersDrawn
	return &g, nil
}

// validNumbers determines if the all the valid numbers are in the game and there are no duplicates.
func (g Game) validNumbers() bool {
	m := make(map[Number]struct{}, MaxNumber)
	for _, n := range g.numbers {
		if n < MinNumber || n > MaxNumber {
			return false // not valid
		}
		if _, ok := m[n]; ok {
			return false // duplicate
		}
		m[n] = struct{}{}
	}
	return true
}
