package bingo

import (
	"bytes"
	"strconv"
	"strings"
	"testing"
)

func TestSVG(t *testing.T) {
	b := &Board{
		15, 8, 4, 12, 10, // B
		19, 27, 16, 28, 25, // I
		42, 41, 0, 31, 40, // N
		49, 52, 50, 46, 57, // G
		64, 72, 67, 70, 74, // O
	}
	var w bytes.Buffer
	gotErr := b.SVG(&w)
	if gotErr != nil {
		t.Fatalf("unwanted export error: %v", gotErr)
	}
	got := w.String()
	for i, n := range b {
		if i == 12 {
			continue
		}
		s := ">" + strconv.Itoa(int(n)) + "<"
		if !strings.Contains(got, s) {
			t.Errorf("wanted board export to contain %q:\n%v", s, got)
			break
		}
	}
}
