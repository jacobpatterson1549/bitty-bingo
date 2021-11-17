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
)

// String is the concatenation of the Number's column letter, a space, and integer value as a string.
func (n Number) String() string {
	if !n.Valid() {
		return "?"
	}
	return string("BINGO"[n.Column()]) + " " + strconv.Itoa(n.Value())
}

// Column is the location on the board the Number should be located at.
func (n Number) Column() int {
	return int(n-1) / 15
}

// Value is integer numeric value of the Number.
func (n Number) Value() int {
	return int(n)
}

// Valid returns whether or not the number is valid, that is, it is between 1 and 75.
func (n Number) Valid() bool {
	switch {
	case n < MinNumber, n > MaxNumber:
		return false
	}
	return true
}

// validNumbers determines if the all the valid numbers are in the game and there are no duplicates.
func validNumbers(numbers []Number, allowZeroValue bool) bool {
	m := make(map[Number]struct{}, len(numbers))
	for _, n := range numbers {
		if _, duplicate := m[n]; duplicate || ((n == 0 && allowZeroValue) != !n.Valid()) {
			return false // duplicate or invalid
		}
		m[n] = struct{}{}
	}
	return true
}
