package bingo

import (
	"encoding/base64"
	"fmt"
)

type Board [25]Number

func NewBoard() *Board {
	var g Game
	columnRowCounts := 0
	var b Board
	for columnRowCounts < 055555 {
		g.DrawNumber()
		n := g.DrawnNumbers[len(g.DrawnNumbers)-1]
		c := n.Column()
		r := (columnRowCounts >> (3 * c)) & 07
		if r < 5 {
			b[c*5+r] = n
			clearRowCountMask := 07 << (3 * c)
			columnRowCounts &^= clearRowCountMask
			r++
			if c == 2 && r == 2 {
				r++ // skip center cell (leave free cell as zero)
			}
			setRowCountMask := r << (3 * c)
			columnRowCounts |= setRowCountMask
		}
	}
	return &b
}

func (b Board) HasLine(g Game) bool {
	nums := numberSet(g)
	for i := 0; i < 5; i++ {
		if b.hasColumn(i, nums) || b.hasRow(i, nums) {
			return true
		}
	}
	return b.hasDiagonal1(nums) || b.hasDiagonal2(nums)
}

func (b Board) IsFilled(g Game) bool {
	nums := numberSet(g)
	for _, n := range b {
		if _, ok := nums[n]; !ok {
			return false
		}
	}
	return true
}

func numberSet(g Game) map[Number]struct{} {
	s := make(map[Number]struct{}, len(g.DrawnNumbers)+1)
	s[0] = struct{}{} // free cell
	for _, n := range g.DrawnNumbers {
		s[n] = struct{}{}
	}
	return s
}

func (b Board) hasColumn(c int, nums map[Number]struct{}) bool {
	for r := 0; r < 5; r++ {
		i := c*5 + r
		if _, ok := nums[b[i]]; !ok {
			return false
		}
	}
	return true
}

func (b Board) hasRow(r int, nums map[Number]struct{}) bool {
	for c := 0; c < 5; c++ {
		i := c*5 + r
		if _, ok := nums[b[i]]; !ok {
			return false
		}
	}
	return true
}

func (b Board) hasDiagonal1(nums map[Number]struct{}) bool {
	for j := 0; j < 5; j++ {
		i := j*5 + j
		if _, ok := nums[b[i]]; !ok {
			return false
		}
	}
	return true
}

func (b Board) hasDiagonal2(nums map[Number]struct{}) bool {
	for j := 0; j < 5; j++ {
		i := j*5 + 4 - j
		if _, ok := nums[b[i]]; !ok {
			return false
		}
	}
	return true
}

var base64Encoding = base64.URLEncoding

// ID encodes the board into a base64 string.
// Each two numbers can be shrunk to a 0-14 number, concatenated, and converted to a byte.
// This results in a byte array that is (25-1)/2 = 12 characters long
// Since there are 8 bits in a byte the array uses 8 * 12 = 96 bits.
// Base 64 uses 6 bits for each character, so the string will be 96 / 6 = 16 characters long
func (b Board) ID() (string, error) {
	for i, n := range b {
		c := i / 5
		if col := n.Column(); i != 12 && col != c {
			return "", fmt.Errorf("board value at row %v, column %v invalid: %v", c+1, i%5+1, n)
		}
	}
	tinyBoard := make([]byte, 0, 12)
	enc := func(n Number) int { return int(n-1) % 15 }
	for i := 0; i < len(b); i++ {
		l := enc(b[i])
		i++
		r := enc(b[i])
		ch := byte(l<<4 | r)
		tinyBoard = append(tinyBoard, ch)
		if i == 11 { // free cell
			i++
		}
	}
	return base64Encoding.EncodeToString(tinyBoard), nil
}

// BoardFromID converts the board id to a Board.
// An error is returned if the id is for an invalid board.
func BoardFromID(id string) (*Board, error) {
	if len(id) != 16 {
		return nil, fmt.Errorf("id must be 16 characters long, got %v (%q)", len(id), id)
	}
	tinyBoard, err := base64Encoding.Strict().DecodeString(id)
	if err != nil {
		return nil, fmt.Errorf("decoding board from id: %v", err)
	}
	fmt.Println(tinyBoard)
	var b Board
	i := 0
	dec := func(h byte, c int) Number { return Number(c*15 + int(h+1)) }
	for _, ch := range tinyBoard {
		l, r := ch>>4, ch&15
		if l == 15 {
			return nil, fmt.Errorf("number index %v in board invalid", i)
		}
		if r == 15 {
			return nil, fmt.Errorf("number index %v in board invalid", i)
		}
		b[i], b[i+1] = dec(l, i/5), dec(r, (i+1)/5)
		i += 2
		if i == 12 {
			i++
		}
	}
	return &b, nil
}
