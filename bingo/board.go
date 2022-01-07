package bingo

import (
	"encoding/base64"
	"errors"
)

// Board represents a 5*5 square bingo board.
// The middle square (index 12) is left empty (0).
type Board [25]Number

// NewBoard creates a board by drawing numbers from a game.
// Each column of the board (5-cell group) only contains numbers of the same column.
func NewBoard() *Board {
	var g Game
	GameResetter.Reset(&g)
	g.numbersDrawn = g.NumbersLeft() // "draw" all the numbers
	cols := g.DrawnNumberColumns()
	var b Board
	copy(b[0:], cols[0][:5])
	copy(b[5:], cols[1][:5])
	copy(b[10:], cols[2][:2])
	copy(b[13:], cols[2][2:4]) // leave space for free cell, copying numbers and indexes 3 & 4 into rows 4 & 5 of middle column
	copy(b[15:], cols[3][:5])
	copy(b[20:], cols[4][:5])
	return &b
}

// HasLine determines if the board has a five-in-a row line, creating a BINGO for the game.
func (b Board) HasLine(g Game) bool {
	nums := numberSet(g)
	if b.hasDiagonal1(nums) || b.hasDiagonal2(nums) {
		return true
	}
	for i := 0; i < 5; i++ {
		if b.hasColumn(i, nums) || b.hasRow(i, nums) {
			return true
		}
	}
	return false
}

// IsFilled determines if all the numbers in the board have been called in the game.
func (b Board) IsFilled(g Game) bool {
	nums := numberSet(g)
	for _, n := range b {
		if _, ok := nums[n]; !ok {
			return false
		}
	}
	return true
}

// NumberSet creates a map of all the drawn numbers in the game.
// It also includes the zero number to always account for the middle free cell.
func numberSet(g Game) map[Number]struct{} {
	nums := make(map[Number]struct{}, g.numbersDrawn+1)
	nums[0] = struct{}{} // free cell
	drawnNumbers := g.DrawnNumbers()
	for _, n := range drawnNumbers {
		nums[n] = struct{}{}
	}
	return nums
}

// hasColumn checks to see if the column on the board is completely included in nums.
func (b Board) hasColumn(c int, nums map[Number]struct{}) bool {
	for r := 0; r < 5; r++ {
		i := c*5 + r
		if _, ok := nums[b[i]]; !ok {
			return false
		}
	}
	return true
}

// hasColumn checks to see if the row on the board is completely included in nums.
func (b Board) hasRow(r int, nums map[Number]struct{}) bool {
	for c := 0; c < 5; c++ {
		i := c*5 + r
		if _, ok := nums[b[i]]; !ok {
			return false
		}
	}
	return true
}

// hasDiagonal1 checks to see if the leading 5-number diagonal on the board is completely included in nums.
func (b Board) hasDiagonal1(nums map[Number]struct{}) bool {
	for j := 0; j < 5; j++ {
		i := j*5 + j
		if _, ok := nums[b[i]]; !ok {
			return false
		}
	}
	return true
}

// hasColumn checks to see if the trailing 5-number diagonal on the board is completely included in nums.
func (b Board) hasDiagonal2(nums map[Number]struct{}) bool {
	for j := 0; j < 5; j++ {
		i := j*5 + 4 - j
		if _, ok := nums[b[i]]; !ok {
			return false
		}
	}
	return true
}

// ID encodes the board into a base64 string.
// Each two numbers can be shrunk to a 0-14 number, concatenated, and converted to a byte.
// This results in a byte array that is (25-1)/2 = 12 characters long
// Since there are 8 bits in a byte the array uses 8 * 12 = 96 bits.
// Base 64 uses 6 bits for each character, so the string will be 96 / 6 = 16 characters long
func (b Board) ID() (string, error) {
	if !b.isValid() {
		return "", errors.New("board has duplicate/invalid numbers")
	}
	data := make([]byte, 0, 12)
	for i := 0; i < len(b); i++ {
		l := b.encodeNumber(i)
		i++
		r := b.encodeNumber(i)
		ch := byte(l<<4 | r)
		data = append(data, ch)
		if i == 11 { // free cell
			i++
		}
	}
	id := base64.URLEncoding.EncodeToString(data)
	return id, nil
}

// BoardFromID converts the board id to a Board.
// An error is returned if the id is for an invalid board.
func BoardFromID(id string) (*Board, error) {
	if len(id) != 16 {
		return nil, errors.New("id must be 16 characters long")
	}
	data, err := base64.URLEncoding.DecodeString(id)
	if err != nil {
		return nil, errors.New("decoding board from id: " + err.Error())
	}
	var b Board
	i := 0
	for _, ch := range data {
		l, err := decodeBoardNumber(ch>>4, i)
		if err != nil {
			return nil, err
		}
		r, err := decodeBoardNumber(ch&15, i+1)
		if err != nil {
			return nil, err
		}
		b[i], b[i+1] = l, r
		i += 2
		if i == 12 { // free cell
			i++
		}
	}
	if !b.isValid() {
		return nil, errors.New("board has duplicate/invalid numbers")
	}
	return &b, nil
}

// encodeNumber converts the number to one in [0,15)
func (b Board) encodeNumber(index int) int {
	encodedNumber := int(b[index]-1) % 15 // (mod 15 is same as subtract n.Column()*15)
	return encodedNumber
}

// decodeBoardNumber converts the [0,15) byte back to a number at the index in the board
func decodeBoardNumber(encodedNumber byte, index int) (Number, error) {
	c := index / 5
	n := Number(int(encodedNumber+1) + c*15)
	return n, nil
}

func (b Board) numbers() numbers {
	nums := make(numbers, len(b)-1) // do not check the center
	copy(nums, b[:12])
	copy(nums[12:], b[13:])
	return nums
}

// isValid determines if the board has valid numbers, no duplicates, and the center is the zero value
func (b Board) isValid() bool {
	switch {
	case b[12] != 0, !b.numbers().Valid(), !b.numbersInCorrectColumns():
		return false
	}
	return true
}

// numbersInCorrectColumns ensures numbers are in correct columns
func (b Board) numbersInCorrectColumns() bool {
	for i, n := range b {
		if i != 12 && n.Column() != i/5 {
			return false
		}
	}
	return true
}
