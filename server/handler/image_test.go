package handler

import (
	"errors"
	"image"
	"strings"
	"testing"

	"github.com/jacobpatterson1549/bitty-bingo/bingo"
)

// init applies a mock QRCode function to make tests IN THIS PACKAGE faster by not calling external libaries
func init() {
	qrCode = func(text string, width, height int) (image.Image, error) {
		r := image.Rect(0, 0, 1, 1)
		img := image.NewGray(r)
		return img, nil
	}
}

func TestFreeSpace(t *testing.T) {
	// according https://www.w3.org/TR/PNG/#5PNG-file-signature,
	//  -> all png files start with [137 80 78 71 13 10 26 10] which encodes to "iVBORw0KGgo"
	prevQREncoder := qrCode
	defer func() {
		qrCode = prevQREncoder
	}()
	const wantDataPrefix = "iVBORw0KGgo"
	tests := []struct {
		name   string
		board  bingo.Board
		wantOk bool
		genQR  qrEncoder
	}{
		{
			name: "invalid board: zero value",
		},
		{
			board:  *bingo.NewBoard(),
			name:   "random board",
			wantOk: true,
			genQR: func(text string, width, height int) (image.Image, error) {
				return image.NewGray(image.Rect(0, 0, 1, 1)), nil
			},
		},
		{
			board:  *bingo.NewBoard(),
			name:   "qr encode error",
			wantOk: false,
			genQR: func(text string, width, height int) (image.Image, error) {
				return nil, errors.New("qr code error")
			},
		},
		{
			board:  *bingo.NewBoard(),
			name:   "qr encode error",
			wantOk: false,
			genQR: func(text string, width, height int) (image.Image, error) {
				return nil, errors.New("qr code error")
			},
		},
		{
			board:  *bingo.NewBoard(),
			name:   "page encode error (empty image",
			wantOk: false,
			genQR: func(text string, width, height int) (image.Image, error) {
				return image.NewGray(image.Rect(0, 0, 0, 0)), nil
			},
		},
	}
	for i, test := range tests {
		qrCode = test.genQR
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
