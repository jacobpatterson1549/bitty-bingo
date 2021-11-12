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
	for i, test := range gameTests {
		test.game.Reset()
		if got := test.game.DrawnNumbers(); len(got) != 0 {
			t.Errorf("test %v: drawn numbers not empty after reset: got %v", i, got)
		}
		if !test.game.validNumbers() {
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

func TestGameID(t *testing.T) {
	t.Run("ok games", func(t *testing.T) {
		for i, test := range gameTests {
			want := test.wantID
			got, err := test.game.ID()
			switch {
			case err != nil:
				t.Errorf("test %v: unwanted error getting game id: %v", i, err)
			case want != got:
				t.Errorf("test %v: ids not equal:\nwanted: %q\ngot:    %q", i, want, got)
			}
		}
	})
	t.Run("invalid games", func(t *testing.T) {
		games := []Game{
			{numbersDrawn: 1}, // first number is invalid
			{numbers: [int(MaxNumber)]Number{9999}, numbersDrawn: 1}, // first number is invalid
			{numbers: [int(MaxNumber)]Number{1, 1}, numbersDrawn: 1}, // duplicate numbers
			// duplicate numbers (first and last)
			{numbers: [int(MaxNumber)]Number{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32, 33, 34, 35, 36, 37, 38, 39, 40, 41, 42, 43, 44, 45, 46, 47, 48, 49, 50, 51, 52, 53, 54, 55, 56, 57, 58, 59, 60, 61, 62, 63, 64, 65, 66, 67, 68, 69, 70, 71, 72, 73, 74, 1}, numbersDrawn: 1},
		}
		for i, game := range games {
			_, err := game.ID()
			if err == nil {
				t.Errorf("test %v: wanted error getting id", i)
			}
		}
	})
}

func TestGameFromID(t *testing.T) {
	t.Run("ok ids", func(t *testing.T) {
		for i, test := range gameTests {
			want := &test.game
			got, err := GameFromID(test.wantID)
			switch {
			case err != nil:
				t.Errorf("test %v: GameFromID(%q): unwanted error : %v", i, test.wantID, err)
			case !reflect.DeepEqual(want, got):
				t.Errorf("test %v: GameFromID(%q):\nwanted: %v\ngot:    %v", i, test.wantID, want, got)
			}
		}
	})
	t.Run("invalid ids", func(t *testing.T) {
		ids := []string{
			"1",          // no hyphen
			"-",          // bad numbersDrawn
			"a-",         // bad numbersDrawn
			"1-",         // no numbers
			"1-!@#%*!@$", // bad numbersLeft
			// numbers not valid (all zeroes):
			"75-AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA",
		}
		for i, id := range ids {
			_, err := GameFromID(id)
			if err == nil {
				t.Errorf("test %v: wanted error getting game from %q", i, id)
			}
		}
	})
}

var gameTests = []struct {
	game                   Game
	wantAvailableAfterDraw Game
	wantNumbersLeft        int
	wantColumns            map[int][]Number
	wantID                 string
}{
	{ // shuffle numbers if game is not initialized; random generator is seeded at 1257894000 for these results
		game: Game{},
		wantAvailableAfterDraw: Game{
			numbers:      [MaxNumber]Number{24, 20, 64, 54, 6, 62, 25, 43, 22, 57, 10, 40, 28, 29, 30, 73, 75, 69, 68, 23, 2, 37, 36, 15, 38, 26, 8, 18, 51, 49, 53, 42, 1, 32, 52, 71, 16, 65, 5, 35, 31, 9, 12, 59, 34, 4, 33, 39, 17, 41, 27, 67, 70, 11, 55, 56, 13, 72, 46, 19, 58, 3, 47, 14, 74, 45, 66, 48, 44, 63, 21, 50, 61, 60, 7},
			numbersDrawn: 1,
		},
		wantNumbersLeft: 75,
		wantColumns:     map[int][]Number{},
		wantID:          "0",
	},
	{ // draw from front of available numbers, add to end of drawn numbers
		game: Game{
			numbers:      [MaxNumber]Number{65, 35, 44, 73, 18, 1, 37, 41, 69, 62, 72, 13, 9, 30, 14, 60, 2, 16, 64, 71, 24, 21, 6, 75, 55, 29, 61, 54, 12, 23, 53, 42, 48, 43, 28, 70, 15, 49, 46, 63, 68, 27, 31, 47, 67, 52, 56, 25, 11, 4, 39, 59, 66, 19, 26, 74, 22, 36, 45, 10, 50, 34, 3, 5, 57, 20, 32, 17, 40, 8, 58, 7, 51, 38, 33},
			numbersDrawn: 3,
		},
		wantAvailableAfterDraw: Game{
			numbers:      [MaxNumber]Number{65, 35, 44, 73, 18, 1, 37, 41, 69, 62, 72, 13, 9, 30, 14, 60, 2, 16, 64, 71, 24, 21, 6, 75, 55, 29, 61, 54, 12, 23, 53, 42, 48, 43, 28, 70, 15, 49, 46, 63, 68, 27, 31, 47, 67, 52, 56, 25, 11, 4, 39, 59, 66, 19, 26, 74, 22, 36, 45, 10, 50, 34, 3, 5, 57, 20, 32, 17, 40, 8, 58, 7, 51, 38, 33},
			numbersDrawn: 4,
		},
		wantNumbersLeft: 72,
		wantColumns: map[int][]Number{
			2: {35, 44},
			4: {65},
		},
		wantID: "3-QSMsSRIBJSlFPkgNCR4OPAIQQEcYFQZLNx09NgwXNSowKxxGDzEuP0QbHy9DNDgZCwQnO0ITGkoWJC0KMiIDBTkUIBEoCDoHMyYh",
	},
	{ // do not draw if all numbers have been drawn
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
		wantID: "75-Oh0hOyw9JDwQDC4yKS8aQzk3HiI1GBUmCzgjMA80BBsDKicIDQIBLTMxGSBIHyUoEUUSKxdBNgc_HBMFBgkWPg4UCkJKREdJS0ZA",
	},
	{ // first 5 numbers are for "5zuTsMm6CTZAs7ad" the rest are sequential
		game: Game{
			numbers:      [MaxNumber]Number{15, 8, 4, 12, 10, 19, 27, 16, 28, 25, 42, 41, 31, 40, 49, 52, 50, 46, 57, 64, 72, 67, 70, 74, 1, 2, 3, 5, 6, 7, 9, 11, 13, 14, 17, 18, 20, 21, 22, 23, 24, 26, 29, 30, 32, 33, 34, 35, 36, 37, 38, 39, 43, 44, 45, 47, 48, 51, 53, 54, 55, 56, 58, 59, 60, 61, 62, 63, 65, 66, 68, 69, 71, 73, 75},
			numbersDrawn: 5,
		},
		wantAvailableAfterDraw: Game{
			numbers:      [MaxNumber]Number{15, 8, 4, 12, 10, 19, 27, 16, 28, 25, 42, 41, 31, 40, 49, 52, 50, 46, 57, 64, 72, 67, 70, 74, 1, 2, 3, 5, 6, 7, 9, 11, 13, 14, 17, 18, 20, 21, 22, 23, 24, 26, 29, 30, 32, 33, 34, 35, 36, 37, 38, 39, 43, 44, 45, 47, 48, 51, 53, 54, 55, 56, 58, 59, 60, 61, 62, 63, 65, 66, 68, 69, 71, 73, 75},
			numbersDrawn: 6,
		},
		wantNumbersLeft: 70,
		wantColumns: map[int][]Number{
			0: {15, 8, 4, 12, 10},
		},
		wantID: "5-DwgEDAoTGxAcGSopHygxNDIuOUBIQ0ZKAQIDBQYHCQsNDhESFBUWFxgaHR4gISIjJCUmJyssLS8wMzU2Nzg6Ozw9Pj9BQkRFR0lL",
	},
	{ // first 24 numbers are for "5zuTsMm6CTZAs7ad" the rest are sequential
		game: Game{
			numbers:      [MaxNumber]Number{15, 8, 4, 12, 10, 19, 27, 16, 28, 25, 42, 41, 31, 40, 49, 52, 50, 46, 57, 64, 72, 67, 70, 74, 1, 2, 3, 5, 6, 7, 9, 11, 13, 14, 17, 18, 20, 21, 22, 23, 24, 26, 29, 30, 32, 33, 34, 35, 36, 37, 38, 39, 43, 44, 45, 47, 48, 51, 53, 54, 55, 56, 58, 59, 60, 61, 62, 63, 65, 66, 68, 69, 71, 73, 75},
			numbersDrawn: 24,
		},
		wantAvailableAfterDraw: Game{
			numbers:      [MaxNumber]Number{15, 8, 4, 12, 10, 19, 27, 16, 28, 25, 42, 41, 31, 40, 49, 52, 50, 46, 57, 64, 72, 67, 70, 74, 1, 2, 3, 5, 6, 7, 9, 11, 13, 14, 17, 18, 20, 21, 22, 23, 24, 26, 29, 30, 32, 33, 34, 35, 36, 37, 38, 39, 43, 44, 45, 47, 48, 51, 53, 54, 55, 56, 58, 59, 60, 61, 62, 63, 65, 66, 68, 69, 71, 73, 75},
			numbersDrawn: 25,
		},
		wantNumbersLeft: 51,
		wantColumns: map[int][]Number{
			0: {15, 8, 4, 12, 10},
			1: {19, 27, 16, 28, 25},
			2: {42, 41, 31, 40},
			3: {49, 52, 50, 46, 57},
			4: {64, 72, 67, 70, 74},
		},
		wantID: "24-DwgEDAoTGxAcGSopHygxNDIuOUBIQ0ZKAQIDBQYHCQsNDhESFBUWFxgaHR4gISIjJCUmJyssLS8wMzU2Nzg6Ozw9Pj9BQkRFR0lL",
	},
}
