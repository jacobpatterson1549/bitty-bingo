package bingo

import "fmt"

type Number int

const (
	MinNumber Number = 1
	MaxNumber Number = 75
	column0   rune   = 'B'
	column1   rune   = 'I'
	column2   rune   = 'N'
	column3   rune   = 'G'
	column4   rune   = 'O'
)

func (n Number) String() string {
	if n < MinNumber || n > MaxNumber {
		return "?"
	}
	var c rune
	switch (n - 1) / 15 {
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
	return fmt.Sprintf("%c %d", c, n)
}

func (n Number) Column() int {
	return int(n-1) / 15
}

func (n Number) Value() int {
	return int(n)
}
