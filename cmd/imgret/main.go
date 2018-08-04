package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"flag"
	"html/template"
	"image"
	"image/color"
	"image/png"
	"io"
	"log"
	"net/http"
	"os/exec"
	"strings"
	"time"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/img/", hashHandler)

	if err := http.ListenAndServe(bind, mux); err != nil {
		log.Panicln(err)
	}
}

func timer(label string, start time.Time) {
	log.Println(label, "after", time.Since(start))
}

func hashHandler(w http.ResponseWriter, r *http.Request) {
	defer timer(r.URL.Path, time.Now())
	type data struct {
		ImageDataBase64 string
	}

	var buf bytes.Buffer
	encoder := base64.NewEncoder(base64.StdEncoding, &buf)

	if err := createImage(strings.NewReader(r.URL.Path), encoder); err != nil {
		log.Panicln(err)
	}

	if err := t.Execute(w, data{buf.String()}); err != nil {
		log.Panicln(err)
	}
}

func mustOpen(fileName string) {
	cmd := exec.Command("xdg-open", fileName)
	err := cmd.Run()
	if err != nil {
		log.Panicln(err)
	}
}

func createImage(r io.Reader, w io.Writer) error {
	h := sha256.New()
	io.Copy(h, r)

	sum := h.Sum(nil)
	bp := bitPNG{bytes: sum, mult: 16}

	if err := png.Encode(w, bp); err != nil {
		return err
	}

	return nil
}

// encodes a 128x128 image
type bitPNG struct {
	mult  int
	bytes []byte
}

func (img bitPNG) ColorModel() color.Model {
	return color.RGBAModel
}

func (img bitPNG) Bounds() image.Rectangle {
	return image.Rectangle{
		image.Point{0, 0},
		image.Point{1024, 1024},
	}
}

func (img bitPNG) At(x, y int) color.Color {
	x /= 1024 / img.mult
	y /= 1024 / img.mult
	pos := x + (y * 16)
	byt := pos / 8
	bit := pos % 8

	mask := byte(1) << byte(bit)

	on := (img.bytes[byt] & mask) > 0

	var r, g, b uint8 = 0, 0, 0
	if on {
		r, g, b = 128, 0, 128
	}

	// Random color shenanigans

	{
		rBit := (bit + 4) % 8
		mask = byte(1) << byte(rBit)

		if on && (img.bytes[byt]&mask) > 0 {
			r, b = 255, 255

		}
	}

	{
		rBit := (bit + 6) % 8
		mask = byte(1) << byte(rBit)

		if on && (img.bytes[byt]&mask) > 0 {
			b = 200
		}
	}

	{
		rBit := (bit + 2) % 8
		mask := byte(1) << byte(rBit)

		if on && (img.bytes[byt]&mask) > 0 {
			r = 64
		}
	}

	{
		rBit := (bit + 7) % 8
		mask := byte(1) << byte(rBit)

		if on && (img.bytes[byt]&mask) > 0 {
			g = 196
			b /= 2
		}
	}

	return color.RGBA{
		A: 255, R: r, G: g, B: b,
	}
}

var doc = `
	<!DOCTYPE html>
	<html>
	<head></head>
	<body>
		<img src="data:image/png;base64,{{ .ImageDataBase64 }}">
	</body>
	</html>
`

var t *template.Template

var bind string

func init() {
	var err error
	t, err = template.New("output.html").Parse(doc)
	if err != nil {
		log.Panicln(err)
	}

	flag.StringVar(&bind, "bind", ":30000", "bind address")
}
