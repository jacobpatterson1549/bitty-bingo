package bingo

import (
	"reflect"
	"testing"
)

func TestNewBoard(t *testing.T) {
	GameResetter.Seed(1257894001) // seed the available numbers for the first test
	b := board1257894001
	if want, got := &b, NewBoard(); !reflect.DeepEqual(want, got) {
		t.Errorf("boards not equal:\nwanted: %v\ngot:    %v", want, got)
	}
}

func TestHasLine(t *testing.T) {
	b := board1257894001
	for i, test := range hasLineTests {
		g := createTestGame(t, test.nums)
		if want, got := test.want, b.HasLine(g); want != got {
			t.Errorf("test %v (%v): wanted %v, got %v", i, test.name, want, got)
		}
	}
}

func TestIsFilled(t *testing.T) {
	b := board1257894001
	t.Run("more-than-line", func(t *testing.T) {
		for i, test := range hasLineTests {
			g := createTestGame(t, test.nums)
			if want, got := false, b.IsFilled(g); want != got {
				t.Errorf("test %v (%v): wanted isFilled() = %v, got %v", i, test.name, want, got)
			}
		}
	})
	t.Run("same as board", func(t *testing.T) {
		g := createTestGame(t, b[:])
		if want, got := true, b.IsFilled(g); want != got {
			t.Errorf("wanted isFilled() = %v, got %v", want, got)
		}
	})
	t.Run("all numbers drawn", func(t *testing.T) {
		nums := []Number{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32, 33, 34, 35, 36, 37, 38, 39, 40, 41, 42, 43, 44, 45, 46, 47, 48, 49, 50, 51, 52, 53, 54, 55, 56, 57, 58, 59, 60, 61, 62, 63, 64, 65, 66, 67, 68, 69, 70, 71, 72, 73, 74, 75}
		g := createTestGame(t, nums)
		if want, got := true, b.IsFilled(g); want != got {
			t.Errorf("wanted isFilled() = %v, got %v", want, got)
		}
	})
}

func TestBoardID(t *testing.T) {
	t.Run("board1257894001", func(t *testing.T) {
		b := board1257894001
		id, err := b.ID()
		if err != nil {
			t.Errorf("unwanted error encoding board: %v", err)
		}
		if want, got := board1257894001ID, id; want != got {
			t.Errorf("ids not equal:\nwanted: %q\ngot:    %q", want, got)
		}
		if want, got := 16, len(id); want != got {
			t.Errorf("sanity check to ensure math in documentation is correct: id lengths not equal: wanted %v, got %v", want, got)
		}
	})
	t.Run("numbers in wrong columns", func(t *testing.T) {
		var b Board
		copy(b[:], board1257894001[:])
		b[0], b[len(b)-1] = b[len(b)-1], b[0]
		if id, err := b.ID(); err == nil {
			t.Errorf("wanted error when swapping first and last values of board (B and O columns), got %q", id)
		}
	})
	t.Run("duplicate numbers", func(t *testing.T) {
		var b Board
		copy(b[:], board1257894001[:])
		b[1] = b[0]
		if id, err := b.ID(); err == nil { // 7juTsMm6CTZAs7ad
			t.Errorf("wanted when first number repeated, got %q", id)
		}
	})
}

func TestBoardFromID(t *testing.T) {
	t.Run("board1257894001", func(t *testing.T) {
		id := board1257894001ID
		want := &board1257894001
		got, err := BoardFromID(id)
		switch {
		case err != nil:
			t.Errorf("unwanted error decoding board from id: %v", err)
		case !reflect.DeepEqual(want, got):
			t.Errorf("decodedBoards not equal:\nwanted: %v\ngot:    %v", want, got)
		}
	})
	t.Run("invalid ids", func(t *testing.T) {
		invalidIds := []struct {
			id   string
			name string
		}{
			{"", "too short"},
			{board1257894001ID + "_", "too long"},
			{"INVALID B64 CHAR", "spaces not allowed"},
			{"9zuTsMm6CTZAs7ad", "first number is too large (15)"},
			{"7zuTsMm6CTZAs7ad", "second number is too large (15)"},
			{"7juTsMm6CTZAs7ad", "first number duplicated"},
		}
		for i, test := range invalidIds {
			if _, err := BoardFromID(test.id); err == nil {
				t.Errorf("test %v (%v): wanted id to be invalid", i, test.name)
			}
		}
	})
}

var board1257894001 = Board{
	15, 8, 4, 12, 10, // B
	19, 27, 16, 28, 25, // I
	42, 41, 0, 31, 40, // N
	49, 52, 50, 46, 57, // G
	64, 72, 67, 70, 74, // O
}

const board1257894001ID = "5zuTsMm6CTZAs7ad"

func createTestGame(t *testing.T, nums []Number) Game {
	t.Helper()
	var g Game
	copy(g.numbers[:], nums)
	g.numbersDrawn = len(nums)
	return g
}

var hasLineTests = []struct {
	name string
	nums []Number
	want bool
}{
	{
		name: "B column",
		nums: []Number{15, 8, 4, 12, 10},
		want: true,
	},
	{
		name: "I column",
		nums: []Number{19, 27, 16, 28, 25},
		want: true,
	},
	{
		name: "N column",
		nums: []Number{42, 41, 31, 40},
		want: true,
	},
	{
		name: "G column",
		nums: []Number{49, 52, 50, 46, 57},
		want: true,
	},
	{
		name: "O column",
		nums: []Number{64, 72, 67, 70, 74},
		want: true,
	},
	{
		name: "row 1",
		nums: []Number{15, 19, 42, 49, 64},
		want: true,
	},
	{
		name: "row 2",
		nums: []Number{8, 27, 41, 52, 72},
		want: true,
	},
	{
		name: "row 3",
		nums: []Number{4, 16, 50, 67},
		want: true,
	},
	{
		name: "row 4",
		nums: []Number{12, 28, 31, 46, 70},
		want: true,
	},
	{
		name: "row 5",
		nums: []Number{10, 25, 40, 57, 74},
		want: true,
	},
	{
		name: "diagonal 1",
		nums: []Number{15, 27, 46, 74},
		want: true,
	},
	{
		name: "diagonal 2",
		nums: []Number{10, 28, 52, 64},
		want: true,
	},
	{
		name: "clover leaf, no corners",
		nums: []Number{8, 19, 27, 12, 28, 25, 49, 52, 72, 46, 57, 70},
		want: false,
	},
	{
		name: "no numbers",
	},
}
