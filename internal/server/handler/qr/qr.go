// Package qr encodes board text into images
package qr

import (
	"fmt"
	"image"

	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/aztec"
	"github.com/boombuler/barcode/datamatrix"
	"github.com/boombuler/barcode/qr"
)

// Formats used to create bar codes.  QR_CODE is the default.
const (
	QR_CODE = iota
	AZTEC
	DATA_MATRIX
)

// Image creates an QR-code image of the text that has the specified dimensions.
// The format should be QR_CODE, AZTEC, or DATA_MATRIX
func Image(format int, text string, width, height int) (image.Image, error) {
	var (
		bc  barcode.Barcode
		err error
	)
	switch format {
	case QR_CODE:
		bc, err = qr.Encode(text, qr.L, qr.Unicode)
	case AZTEC:
		bc, err = aztec.Encode([]byte(text), aztec.DEFAULT_EC_PERCENT, aztec.DEFAULT_LAYERS)
	case DATA_MATRIX:
		bc, err = datamatrix.Encode(text)
	default:
		err = fmt.Errorf("unknown barcode format: %v", format)
	}
	if err != nil {
		return nil, fmt.Errorf("unexpected problem encoding QR code image: %v", err)
	}
	bc, err = barcode.Scale(bc, width, height)
	if err != nil {
		return nil, fmt.Errorf("unexpected problem scaling QR code image: %v", err)
	}
	return bc, nil
}
