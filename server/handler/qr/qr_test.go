package qr

import "testing"

func TestQRCode(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test that depends on external library")
	}
	for i, test := range imageTests {
		got, err := QRCode(test.text, test.width, test.height)
		switch {
		case !test.wantOk:
			if err == nil {
				t.Errorf("test %v (%v): wanted error", i, test.name)
			}
		case err != nil:
			t.Errorf("test %v (%v): %v", i, test.name, err)
		case got == nil:
			t.Errorf("test %v (%v): wanted image", i, test.name)
		}
	}
}

var imageTests = []struct {
	name   string
	text   string
	width  int
	height int
	wantOk bool
}{
	{
		name:   "board1257894001 ID",
		text:   "5zuTsMm6CTZAs7ad",
		width:  80,
		height: 80,
		wantOk: true,
	},
	{
		name:   "zero width/height",
		wantOk: false,
	},
	{
		// see https://www.qrcode.com/en/about/version.html
		name:   "too much text, one larger than Version 40, 177x177 modules, ECC Level L, Binary",
		text:   string(make([]byte, 2953+1)),
		width:  1000,
		height: 1000,
		wantOk: false,
	},
}
