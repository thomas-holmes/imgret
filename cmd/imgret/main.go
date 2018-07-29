package main

import (
	"crypto/sha256"
	"image"
	"image/color"
	"image/png"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	doHashThings(strings.NewReader("Thomas Holmes"))
}

func mustOpen(fileName string) {
	cmd := exec.Command("xdg-open", fileName)
	err := cmd.Run()
	if err != nil {
		log.Panicln(err)
	}
}

func doHashThings(r io.Reader) error {
	h := sha256.New()
	io.Copy(h, r)

	sum := h.Sum(nil)
	log.Println(sum)
	bp := bitPNG{bytes: sum, mult: 16}

	dir, err := ioutil.TempDir("", "image")
	if err != nil {
		return err
	}

	outFile, err := os.Create(filepath.Join(dir, "hashed.png"))
	if err != nil {
		return err
	}
	defer outFile.Close()

	if err := png.Encode(outFile, bp); err != nil {
		return err
	}

	outFile.Close()

	mustOpen(outFile.Name())

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

	return color.RGBA{
		A: 255, R: r, G: g, B: b,
	}
}
