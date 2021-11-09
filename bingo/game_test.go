package bingo

import (
	"math/rand"
	"reflect"
	"testing"
)

func TestGameNumbersLeft(t *testing.T) {
	for i, test := range gameTests {
		if want, got := test.wantNumbersLeft, test.game.NumbersLeft(); want != got {
			t.Errorf("test %v: wanted %v numbers left, got %v", i, want, got)
		}
	}
}

func TestGameDrawNumber(t *testing.T) {
	rand.Seed(1257894000) // seed the available numbers for the first test
	for i, test := range gameTests {
		test.game.DrawNumber()
		if want, got := test.wantAvailableAfterDraw, test.game; !reflect.DeepEqual(want, got) {
			t.Errorf("test %v: games not equal after number drawn:\nwanted: %v\ngot:    %v", i, want, got)
		}
	}
}

func TestResetGame(t *testing.T) {
	for i, test := range gameTests {
		test.game.Reset()
		if want, got := 75, len(test.game.availableNumbers); want != got {
			t.Errorf("test %v: wanted %v available numbers after reset, got %v", i, want, got)
		}
	}
}

var gameTests = []struct {
	game                   Game
	wantAvailableAfterDraw Game
	wantNumbersLeft        int
}{
	{ // shuffle numbers if game is not initialized; random generator is seeded at 1257894000 for these results
		wantAvailableAfterDraw: Game{
			DrawnNumbers:     []Number{24},
			availableNumbers: []Number{20, 64, 54, 6, 62, 25, 43, 22, 57, 10, 40, 28, 29, 30, 73, 75, 69, 68, 23, 2, 37, 36, 15, 38, 26, 8, 18, 51, 49, 53, 42, 1, 32, 52, 71, 16, 65, 5, 35, 31, 9, 12, 59, 34, 4, 33, 39, 17, 41, 27, 67, 70, 11, 55, 56, 13, 72, 46, 19, 58, 3, 47, 14, 74, 45, 66, 48, 44, 63, 21, 50, 61, 60, 7},
		},
		wantNumbersLeft: 0,
	},
	{ // NOOP if the game has drawn numbers and there are no available numbers to draw
		game: Game{
			DrawnNumbers: []Number{1},
		},
		wantAvailableAfterDraw: Game{
			DrawnNumbers: []Number{1},
		},
		wantNumbersLeft: 0,
	},
	{ // first draw
		game: Game{
			DrawnNumbers:     []Number{},
			availableNumbers: []Number{8, 14, 3},
		},
		wantAvailableAfterDraw: Game{
			DrawnNumbers:     []Number{8},
			availableNumbers: []Number{14, 3},
		},
		wantNumbersLeft: 3,
	},
	{ // draw from front of available numbers, add to end of drawn numbers
		game: Game{
			DrawnNumbers:     []Number{7, 28, 3},
			availableNumbers: []Number{21, 34, 18},
		},
		wantAvailableAfterDraw: Game{
			DrawnNumbers:     []Number{7, 28, 3, 21},
			availableNumbers: []Number{34, 18},
		},
		wantNumbersLeft: 3,
	},
}
