package handler

import (
	"errors"
	"image"
	"image/color"
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

func TestNewTransparentImage(t *testing.T) {
	b := image.Rect(0, 0, 0, 0)
	tests := []struct {
		name   string
		m      image.Image
		wantOk bool
	}{
		{
			name:   "valid gray16 image",
			m:      image.NewGray16(b),
			wantOk: true,
		},
		{
			name:   "colorful image",
			m:      image.NewRGBA(b),
			wantOk: false,
		},
	}
	for i, test := range tests {
		got, err := newTransparentImage(test.m)
		switch {
		case !test.wantOk:
			if err == nil {
				t.Errorf("test %v (%v): wanted error", i, test.name)
			}
		case err != nil:
			t.Errorf("test %v (%v): unwanted error: %v", i, test.name, err)
		case got == nil:
			t.Errorf("test %v (%v): wanted image", i, test.name)
		}
	}
}

func TestTransparentImageColorsBlackAndTransparent(t *testing.T) {
	m := image.NewGray16(image.Rect(0, 0, 2, 2))
	m.Set(0, 0, color.Black)
	m.Set(0, 1, color.White)
	m.Set(1, 0, color.Transparent)
	m.Set(1, 1, color.Opaque)
	got := transparentImage{m}
	b := got.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			c := got.At(x, y)
			switch c {
			case color.Black, color.Transparent:
				// NOOP (wanted color)
			default:
				t.Fatalf("color at pixel [%v,%v] is not black or transparent: %#v", x, y, c)
			}
		}
	}
}

func TestTransparentImageColorModel(t *testing.T) {
	var m transparentImage
	if want, got := color.NRGBAModel, m.ColorModel(); want != got {
		t.Errorf("colors models not equal - wanted color model that has transparency with png:\nwanted: %v\ngot:    %v", want, got)
	}
}
