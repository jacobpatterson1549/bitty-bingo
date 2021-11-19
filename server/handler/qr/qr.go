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
	qrcode, err := qr.Encode(text, qr.L, qr.Unicode)
	if err != nil {
		return nil, fmt.Errorf("unexpected problem ecoding QR code image: %v", err)
	}
	qrcode, err = barcode.Scale(qrcode, width, height)
	if err != nil {
		return nil, fmt.Errorf("unexpected problem scaling QR code image: %v", err)
	}
	return qrcode, nil
}
