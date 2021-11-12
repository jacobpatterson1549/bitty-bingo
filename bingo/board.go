package bingo

import (
	"encoding/base64"
	"errors"
	"strconv"
)

type Board [25]Number

func NewBoard() *Board {
	var g Game
	columnRowCounts := 0
	var b Board
	for columnRowCounts < 055555 {
		g.DrawNumber()
		n := g.numbers[g.numbersDrawn-1]
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
	s := make(map[Number]struct{}, g.numbersDrawn+1)
	s[0] = struct{}{} // free cell
	drawnNumbers := g.DrawnNumbers()
	for _, n := range drawnNumbers {
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

// base64Encoding is used to encode/decode boards/ids.  The ids can be put in urls.
var base64Encoding = base64.URLEncoding

// ID encodes the board into a base64 string.
// Each two numbers can be shrunk to a 0-14 number, concatenated, and converted to a byte.
// This results in a byte array that is (25-1)/2 = 12 characters long
// Since there are 8 bits in a byte the array uses 8 * 12 = 96 bits.
// Base 64 uses 6 bits for each character, so the string will be 96 / 6 = 16 characters long
func (b Board) ID() (string, error) {
	if b.hasDuplicateNumbers() {
		return "", errors.New("board has duplicate numbers")
	}
	for i, n := range b {
		c := i / 5
		if col := n.Column(); i != 12 && col != c {
			return "", errors.New("board has number at incorrect column at index " + strconv.Itoa(i))
		}
	}
	tinyBoard := make([]byte, 0, 12)
	for i := 0; i < len(b); i++ {
		l := encodeNumber(b[i])
		i++
		r := encodeNumber(b[i])
		ch := byte(l<<4 | r)
		tinyBoard = append(tinyBoard, ch)
		if i == 11 { // free cell
			i++
		}
	}
	id := base64Encoding.EncodeToString(tinyBoard)
	return id, nil
}

// BoardFromID converts the board id to a Board.
// An error is returned if the id is for an invalid board.
func BoardFromID(id string) (*Board, error) {
	if len(id) != 16 {
		return nil, errors.New("id must be 16 characters long")
	}
	tinyBoard, err := base64Encoding.Strict().DecodeString(id)
	if err != nil {
		return nil, errors.New("decoding board from id: " + err.Error())
	}
	var b Board
	i := 0
	for _, ch := range tinyBoard {
		l, err := decodeNumber(ch>>4, i)
		if err != nil {
			return nil, err
		}
		r, err := decodeNumber(ch&15, i+1)
		if err != nil {
			return nil, err
		}
		b[i], b[i+1] = l, r
		i += 2
		if i == 12 {
			i++
		}
	}
	if b.hasDuplicateNumbers() {
		return nil, errors.New("board has duplicate numbers")
	}
	return &b, nil
}

// encodeNumber converts the number to one in [0,15)
func encodeNumber(n Number) int {
	h := int(n-1) % 15 // (mod 15 is same as subtract n.Column()*15)
	return h
}

// decodeNumber converts the [0,15) byte back to a number at the index in the board
func decodeNumber(h byte, i int) (Number, error) {
	if h == 15 {
		return 0, errors.New("board has invalid number at index " + strconv.Itoa(i))
	}
	c := i / 5
	n := Number(int(h+1) + c*15)
	return n, nil
}

// hasDuplicateNumbers determines if the board has duplicate numbers
func (b Board) hasDuplicateNumbers() bool {
	m := make(map[Number]struct{}, len(b))
	for _, n := range b {
		if _, ok := m[n]; ok {
			return true
		}
		m[n] = struct{}{}
	}
	return false
}
