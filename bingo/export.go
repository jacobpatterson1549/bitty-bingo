package bingo

import (
	_ "embed"
	"io"
	"text/template"
)

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
