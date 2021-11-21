package handler

import (
	"image"
	"image/color"
	"testing"
)

func TestNewTransparentImage(t *testing.T) {
	b := image.Rect(0, 0, 0, 0)
	tests := []struct {
		name string
		m    image.Image
	}{
		{
			name: "valid gray16 image",
			m:    image.NewGray16(b),
		},
		{
			name: "colorful image RGBA",
			m:    image.NewRGBA(b),
		},
		{
			name: "colorful image RGB",
			m:    image.NewNRGBA(b),
		},
	}
	for i, test := range tests {
		got := newTransparentImage(test.m)
		if want, got := test.m.ColorModel().Convert(color.Black), got.blackColor; want != got {
			t.Errorf("test %v (%v): black color is not correct for the color model:\nwanted: %v\ngot:    %v", i, test.name, want, got)
		}
	}
}

func TestTransparentImageColorsBlackAndTransparent(t *testing.T) {
	m := image.NewNRGBA(image.Rect(0, 0, 2, 2))
	m.Set(0, 0, color.Black)
	m.Set(0, 1, color.White)
	m.Set(1, 0, color.Transparent)
	m.Set(1, 1, color.Opaque)
	black := m.ColorModel().Convert(color.Black)
	transparent := m.ColorModel().Convert(color.Transparent)
	got := newTransparentImage(m)
	b := got.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			c := got.At(x, y)
			switch c {
			case black, transparent:
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
		t.Errorf("wanted color model that has non-alpha-premultiplied colors:\nwanted: %v\ngot:    %v", want, got)
	}
}
