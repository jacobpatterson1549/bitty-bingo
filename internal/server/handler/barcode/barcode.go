// Package barcode encodes text into images.
package barcode

import (
	"fmt"
	"image"

	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/aztec"
	"github.com/boombuler/barcode/datamatrix"
	"github.com/boombuler/barcode/qr"
)

// Formats used to create bar codes.  QR_CODE is the default.
type Format int

const (
	QR_CODE Format = iota
	AZTEC
	DATA_MATRIX
)

// Image creates an QR-code image of the text that has the specified dimensions.
func Image(f Format, text string, width, height int) (image.Image, error) {
	bc, err := f.newBarcode(text)
	if err != nil {
		return nil, fmt.Errorf("unexpected problem encoding QR code image: %v", err)
	}
	bc, err = barcode.Scale(bc, width, height)
	if err != nil {
		return nil, fmt.Errorf("unexpected problem scaling QR code image: %v", err)
	}
	return bc, nil
}

// newBarcode creates a barcode with the text.
func (f Format) newBarcode(text string) (barcode.Barcode, error) {
	switch f {
	case QR_CODE:
		return qr.Encode(text, qr.L, qr.Unicode)
	case AZTEC:
		return aztec.Encode([]byte(text), aztec.DEFAULT_EC_PERCENT, aztec.DEFAULT_LAYERS)
	case DATA_MATRIX:
		return datamatrix.Encode(text)
	default:
		return nil, fmt.Errorf("unknown barcode format: %v", f)
	}
}
