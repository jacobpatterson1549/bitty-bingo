package bingo

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
