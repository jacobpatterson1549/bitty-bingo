package barcode

import "testing"

func TestImage(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test that depends on external library")
	}
	for i, test := range imageTests {
		got, err := Image(test.format, test.text, test.width, test.height)
		switch {
		case !test.wantOk:
			if err == nil {
				t.Errorf("test %v (%v): wanted error", i, test.name)
			}
		case err != nil:
			t.Errorf("test %v (%v): %v", i, test.name, err)
		case got == nil, got.Bounds().Dx() != test.width, got.Bounds().Dy() != test.height:
			t.Errorf("test %v (%v): wanted [%vx%v] image, got %v", i, test.name, test.width, test.height, got)
		}
	}
}

var imageTests = []struct {
	format Format
	name   string
	text   string
	width  int
	height int
	wantOk bool
}{
	{
		name:   "board1257894001 ID - default to  QR",
		text:   "5zuTsMm6CTZAs7ad",
		width:  80,
		height: 80,
		wantOk: true,
	},
	{
		format: QR_CODE,
		name:   "board1257894001 ID - QR_CODE",
		text:   "5zuTsMm6CTZAs7ad",
		width:  80,
		height: 80,
		wantOk: true,
	},
	{
		format: AZTEC,
		name:   "board1257894001 ID - AZTEC",
		text:   "5zuTsMm6CTZAs7ad",
		width:  80,
		height: 80,
		wantOk: true,
	},
	{
		format: DATA_MATRIX,
		name:   "board1257894001 ID - AZTEC",
		text:   "5zuTsMm6CTZAs7ad",
		width:  80,
		height: 80,
		wantOk: true,
	},
	{
		format: 1000,
		name:   "board1257894001 ID - bad format",
		text:   "5zuTsMm6CTZAs7ad",
		width:  80,
		height: 80,
		wantOk: false,
	},
	{
		name:   "zero width/height",
		wantOk: false,
	},
	{
		name:   "too much text, one larger than Version 40, 177x177 modules, ECC Level L, Binary (see https://www.qrcode.com/en/about/version.html)",
		text:   string(make([]byte, 2953+1)),
		width:  1000,
		height: 1000,
		wantOk: false,
	},
}
