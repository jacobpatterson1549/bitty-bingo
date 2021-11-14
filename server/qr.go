package server

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image/png"

	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/qr"
	"github.com/jacobpatterson1549/bitty-bingo/bingo"
)

// qrEncoding provides the base64 encoding for encoding png qr images to base64: [A-Za-z0-9+/]
var qrEncoding = base64.StdEncoding

// freeSpace converts the id of the board into a base-64 encoded png image of the qr code of the id of the board.
func freeSpace(b bingo.Board) (string, error) {
	id, err := b.ID()
	if err != nil {
		return "", fmt.Errorf("getting board id %v", err)
	}
	qrcode, err := qr.Encode(id, qr.H, qr.Unicode)
	if err != nil {
		return "", fmt.Errorf("ecoding qr image: %v", err)
	}
	qrcode, err = barcode.Scale(qrcode, 80, 80)
	if err != nil {
		return "", fmt.Errorf("scaling qr code: %v", err)
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, qrcode); err != nil {
		return "", fmt.Errorf("converting barcode to png image: %v", err)
	}
	bytes := buf.Bytes()
	data := qrEncoding.EncodeToString(bytes)
	return data, nil
}
