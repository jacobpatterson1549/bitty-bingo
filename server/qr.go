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

// freeSpace converts the id of the board to a qr code, to a png image, and then encodes it with standard base64 encoding.
func freeSpace(b bingo.Board) (string, error) {
	id, err := b.ID()
	if err != nil {
		return "", fmt.Errorf("getting board id %v", err)
	}
	qrcode, err := qr.Encode(id, qr.L, qr.Unicode)
	if err != nil {
		return "", fmt.Errorf("unexpected problem ecoding qr image: %v", err)
	}
	qrcode, err = barcode.Scale(qrcode, 80, 80)
	if err != nil {
		return "", fmt.Errorf("unexpected problem scaling qr code: %v", err)
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, qrcode); err != nil {
		return "", fmt.Errorf("unexpected problem converting barcode to png image: %v", err)
	}
	bytes := buf.Bytes()
	data := base64.StdEncoding.EncodeToString(bytes)
	return data, nil
}
