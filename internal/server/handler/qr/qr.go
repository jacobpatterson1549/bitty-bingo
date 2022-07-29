// Package qr encodes board text into images
package qr

import (
	"fmt"
	"image"

	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/qr"
)

// Image creates an QR-code image of the text that has the specified dimensions.
func Image(text string, width, height int) (image.Image, error) {
	// qrCode, err := aztec.Encode([]byte(text), aztec.DEFAULT_EC_PERCENT, aztec.DEFAULT_LAYERS)
	// qrCode, err := datamatrix.Encode(text)
	qrCode, err := qr.Encode(text, qr.L, qr.Unicode)
	if err != nil {
		return nil, fmt.Errorf("unexpected problem encoding QR code image: %v", err)
	}
	qrCode, err = barcode.Scale(qrCode, width, height)
	if err != nil {
		return nil, fmt.Errorf("unexpected problem scaling QR code image: %v", err)
	}
	return qrCode, nil
}
