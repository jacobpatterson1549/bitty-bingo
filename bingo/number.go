package bingo

import (
	"strconv"
)

// Number represents a value that can be drawn in the game.
// Boards are made of numbers.
// If a board has five numbers in a row, column, or diagonal, the board has a BINGO.
type Number int

const (
	// MinNumber is the minimum allowed value of a Number.
	MinNumber Number = 1
	// MaxNumber is the maximum allowed value of a Number.
	MaxNumber Number = 75
	column0   rune   = 'B'
	column1   rune   = 'I'
	column2   rune   = 'N'
	column3   rune   = 'G'
	column4   rune   = 'O'
)

// String is the concatenation of the Number's column letter, a space, and integer value as a string.
func (n Number) String() string {
	if n < MinNumber || n > MaxNumber {
		return "?"
	}
	return string(n.letter()) + " " + strconv.Itoa(n.Value())
}

// letter gets the letter for the column of the number.
// The letter is only set if the column is in [0,4]
func (n Number) letter() (c rune) {
	switch n.Column() {
	case 0:
		c = 'B'
	case 1:
		c = 'I'
	case 2:
		c = 'N'
	case 3:
		c = 'G'
	case 4:
		c = 'O'
	}
	return
}

// Column is the location on the board the Number should be located at.
func (n Number) Column() int {
	return int(n-1) / 15
}

// Value is integer numeric value of the Number.
func (n Number) Value() int {
	return int(n)
}
