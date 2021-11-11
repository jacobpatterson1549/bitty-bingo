package bingo

import (
	"math/rand"
	"reflect"
	"testing"
)

func TestGameNumbersLeft(t *testing.T) {
	for i, test := range gameTests {
		if want, got := test.wantNumbersLeft, test.game.NumbersLeft(); want != got {
			t.Errorf("test %v: numbers left not equal: wanted %v, got %v", i, want, got)
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
	hasAllAvailableNumbers := func(numbers []Number) bool {
		m := make(map[Number]struct{}, MaxNumber)
		for _, n := range numbers {
			if n < MinNumber || n > MaxNumber {
				return false // not valid
			}
			if _, ok := m[n]; ok {
				return false // duplicate
			}
			m[n] = struct{}{}
		}
		return true
	}
	for i, test := range gameTests {
		test.game.Reset()
		if got := test.game.DrawnNumbers(); len(got) != 0 {
			t.Errorf("test %v: drawn numbers not empty after reset: got %v", i, got)
		}
		if !hasAllAvailableNumbers(test.game.numbers[:]) {
			t.Errorf("test %v: not all numbers available after reset: %v", i, test.game.numbers)
		}
	}
}

func TestColumns(t *testing.T) {
	for i, test := range gameTests {
		if want, got := test.wantColumns, test.game.Columns(); !reflect.DeepEqual(want, got) {
			t.Errorf("test %v: columns not equal:\nwanted: %v\ngot:    %v", i, want, got)
		}
	}
}

var gameTests = []struct {
	game                   Game
	wantAvailableAfterDraw Game
	wantNumbersLeft        int
	wantColumns            map[int][]Number
}{
	{ // shuffle numbers if game is not initialized; random generator is seeded at 1257894000 for these results
		wantAvailableAfterDraw: Game{
			numbers:      [MaxNumber]Number{24, 20, 64, 54, 6, 62, 25, 43, 22, 57, 10, 40, 28, 29, 30, 73, 75, 69, 68, 23, 2, 37, 36, 15, 38, 26, 8, 18, 51, 49, 53, 42, 1, 32, 52, 71, 16, 65, 5, 35, 31, 9, 12, 59, 34, 4, 33, 39, 17, 41, 27, 67, 70, 11, 55, 56, 13, 72, 46, 19, 58, 3, 47, 14, 74, 45, 66, 48, 44, 63, 21, 50, 61, 60, 7},
			numbersDrawn: 1,
		},
		wantNumbersLeft: 75,
		wantColumns:     map[int][]Number{},
	},
	{ // NOOP if the game has drawn numbers and there are no available numbers to draw
		game: Game{
			numbers:      [MaxNumber]Number{66},
			numbersDrawn: int(MaxNumber),
		},
		wantAvailableAfterDraw: Game{
			numbers:      [MaxNumber]Number{66},
			numbersDrawn: int(MaxNumber),
		},
		wantNumbersLeft: 0,
		wantColumns: map[int][]Number{
			4: {66},
			// 74 zeroes at the end because numbers is a fixed sized array and all have been drawn
			0: {0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		},
	},
	{ // draw from front of available numbers, add to end of drawn numbers
		game: Game{
			numbers:      [MaxNumber]Number{7, 28, 3, 21, 34, 18},
			numbersDrawn: 3,
		},
		wantAvailableAfterDraw: Game{
			numbers:      [MaxNumber]Number{7, 28, 3, 21, 34, 18},
			numbersDrawn: 4,
		},
		wantNumbersLeft: 72,
		wantColumns: map[int][]Number{
			0: {7, 3},
			1: {28},
		},
	},
	{ // al numbers, shuffled
		game: Game{
			numbers:      [MaxNumber]Number{58, 29, 33, 59, 44, 61, 36, 60, 16, 12, 46, 50, 41, 47, 26, 67, 57, 55, 30, 34, 53, 24, 21, 38, 11, 56, 35, 48, 15, 52, 4, 27, 3, 42, 39, 8, 13, 2, 1, 45, 51, 49, 25, 32, 72, 31, 37, 40, 17, 69, 18, 43, 23, 65, 54, 7, 63, 28, 19, 5, 6, 9, 22, 62, 14, 20, 10, 66, 74, 68, 71, 73, 75, 70, 64},
			numbersDrawn: 75,
		},
		wantAvailableAfterDraw: Game{
			numbers:      [MaxNumber]Number{58, 29, 33, 59, 44, 61, 36, 60, 16, 12, 46, 50, 41, 47, 26, 67, 57, 55, 30, 34, 53, 24, 21, 38, 11, 56, 35, 48, 15, 52, 4, 27, 3, 42, 39, 8, 13, 2, 1, 45, 51, 49, 25, 32, 72, 31, 37, 40, 17, 69, 18, 43, 23, 65, 54, 7, 63, 28, 19, 5, 6, 9, 22, 62, 14, 20, 10, 66, 74, 68, 71, 73, 75, 70, 64},
			numbersDrawn: 75,
		},
		wantNumbersLeft: 0,
		wantColumns: map[int][]Number{
			0: {12, 11, 15, 4, 3, 8, 13, 2, 1, 7, 5, 6, 9, 14, 10},
			1: {29, 16, 26, 30, 24, 21, 27, 25, 17, 18, 23, 28, 19, 22, 20},
			2: {33, 44, 36, 41, 34, 38, 35, 42, 39, 45, 32, 31, 37, 40, 43},
			3: {58, 59, 60, 46, 50, 47, 57, 55, 53, 56, 48, 52, 51, 49, 54},
			4: {61, 67, 72, 69, 65, 63, 62, 66, 74, 68, 71, 73, 75, 70, 64},
		},
	},
}
