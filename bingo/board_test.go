package bingo

import (
	"math/rand"
	"reflect"
	"testing"
)

func TestNewBoard(t *testing.T) {
	rand.Seed(1257894001) // seed the available numbers for the first test
	// &[B 15 B 8 B 4 B 12 B 10 I 19 I 27 I 16 I 28 I 25 N 42 N 41 ? N 31 N 40 G 49 G 52 G 50 G 46 G 57 O 64 O 72 O 67 O 70 O 74]
	want := &Board{
		15, 8, 4, 12, 10, // B
		19, 27, 16, 28, 25, // I
		42, 41, 0, 31, 40, // N
		49, 52, 50, 46, 57, // G
		64, 72, 67, 70, 74, // O
	}
	got := NewBoard()
	if !reflect.DeepEqual(want, got) {

		t.Errorf("boards not equal:\nwanted: %v\ngot:    %v", want, got)
	}
}
