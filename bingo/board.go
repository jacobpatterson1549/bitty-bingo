package bingo

import (
	_ "embed"
	"io"
	"text/template"
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
				r++ // skip center cell
			}
			setRowCountMask := r << (3 * c)
			columnRowCounts |= setRowCountMask
		}
	}
	return &b
}

//go:embed template.svg
var svgTemplate string

var t = template.Must(template.New("template.svg").
	Funcs(template.FuncMap{
		"int": func(b Board, i int) int {
			return b[i].Value()
		},
	}).
	Parse(svgTemplate))

func (b Board) SVG(w io.Writer) error {
	return t.Execute(w, b)
}
