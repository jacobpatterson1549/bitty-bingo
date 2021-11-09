package bingo

import (
	"math/rand"
	"reflect"
	"testing"
)

func TestGameNumbersLeft(t *testing.T) {
	for i, test := range gameNumbersLeftTests {
		if got := test.game.NumbersLeft(); test.want != got {
			t.Errorf("test %v: wanted %v numbers left, got %v", i, test.want, got)
		}
	}
}

func TestGameDrawNumber(t *testing.T) {
	rand.Seed(1257894000) // seed the available numbers for the first test
	for i, test := range gameDrawNumberTests {
		test.game.DrawNumber()
		if got := test.game; !reflect.DeepEqual(test.want, got) {
			t.Errorf("test %v: games not equal after number drawn:\nwanted: %v\ngot:    %v", i, test.want, got)
		}
	}
}

func TestResetGame(t *testing.T) {
	g := Game{
		DrawnNumbers: []Number{23},
		availableNumbers: []Number{47, 39},
	}
	g.Reset()
	if want, got := 75, len(g.availableNumbers); want != got {
		t.Errorf("wanted %v available numbers after reset, got %v", want, got)
	}
	if want, got := 0, len(g.DrawnNumbers); want != got {
		t.Errorf("wanted %v drawn numbers after reset, got %v", want, got)
	}
}

var gameNumbersLeftTests = []struct {
	game Game
	want int
}{
	{},
	{
		game: Game{
			availableNumbers: []Number{1, 2, 3},
		},
		want: 3,
	},
	{
		game: Game{
			DrawnNumbers:     []Number{6, 43, 8},
			availableNumbers: []Number{19},
		},
		want: 1,
	},
}

var gameDrawNumberTests = []struct {
	game Game
	want Game
}{
	{ // shuffle numbers if game is not initialized; random generator is seeded at 1257894000 for these results
		want: Game{
			DrawnNumbers:     []Number{24},
			availableNumbers: []Number{20, 64, 54, 6, 62, 25, 43, 22, 57, 10, 40, 28, 29, 30, 73, 75, 69, 68, 23, 2, 37, 36, 15, 38, 26, 8, 18, 51, 49, 53, 42, 1, 32, 52, 71, 16, 65, 5, 35, 31, 9, 12, 59, 34, 4, 33, 39, 17, 41, 27, 67, 70, 11, 55, 56, 13, 72, 46, 19, 58, 3, 47, 14, 74, 45, 66, 48, 44, 63, 21, 50, 61, 60, 7},
		},
	},
	{ // NOOP if the game has drawn numbers and there are no available numbers to draw
		game: Game{
			DrawnNumbers: []Number{1},
		},
		want: Game{
			DrawnNumbers: []Number{1},
		},
	},
	{ // first draw
		game: Game{
			DrawnNumbers:     []Number{},
			availableNumbers: []Number{8, 14, 3},
		},
		want: Game{
			DrawnNumbers:     []Number{8},
			availableNumbers: []Number{14, 3},
		},
	},
	{ // draw from front of available numbers, add to end of drawn numbers
		game: Game{
			DrawnNumbers:     []Number{7, 28, 3},
			availableNumbers: []Number{21, 34, 18},
		},
		want: Game{
			DrawnNumbers:     []Number{7, 28, 3, 21},
			availableNumbers: []Number{34, 18},
		},
	},
}
