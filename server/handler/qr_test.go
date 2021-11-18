package handler

import (
	"strings"
	"testing"

	"github.com/jacobpatterson1549/bitty-bingo/bingo"
)

func TestFreeSpace(t *testing.T) {
	// the base64 encoding of the png image of the qr codes all appear to have this prefix
	// according https://www.w3.org/TR/PNG/#5PNG-file-signature,
	//  -> all png files start with [137 80 78 71 13 10 26 10] which encodes to "iVBORw0KGgo"
	const wantDataPrefix = "iVBORw0KGgoAAAANSUhEUgAAAFAAAABQEAAAAAD76nEyAAAB"
	tests := []struct {
		name   string
		board  bingo.Board
		wantOk bool
	}{
		{
			name: "invalid board: zero value",
		},
		{
			board:  *bingo.NewBoard(),
			name:   "random board",
			wantOk: true,
		},
	}
	for i, test := range tests {
		data, err := freeSpace(test.board)
		switch {
		case !test.wantOk:
			if err == nil {
				t.Errorf("test %v (%v): wanted error", i, test.name)
			}
		case err != nil:
			t.Errorf("test %v (%v): unwanted error: %v", i, test.name, err)
		case !strings.HasPrefix(data, wantDataPrefix):
			id, _ := test.board.ID() // ignore error because it did not cause an error generating the free space
			t.Errorf("test %v (%v): prefix of free space not equal for board of ID=%q:\n"+
				"the base64 encoding of the png image of the qr code of the board id was unwanted:\nwanted: %v\ngot:    %v",
				i, test.name, id, wantDataPrefix, data)
		}
	}
}
