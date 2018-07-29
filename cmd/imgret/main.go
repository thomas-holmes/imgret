package main

import (
	"image"
	"image/color"
	"image/png"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

func main() {
	tempDir, err := ioutil.TempDir("", "image")
	if err != nil {
		log.Panicln(err)
	}

	outFile, err := os.Create(filepath.Join(tempDir, "picture3.png"))
	if err != nil {
		log.Panicln(err)
	}

	func() {
		defer outFile.Close()

		if err = png.Encode(outFile, rainbowPNG{}); err != nil {
			log.Panicln(err)
		}
	}()

	mustOpen(outFile.Name())

}

func mustOpen(fileName string) {
	cmd := exec.Command("xdg-open", fileName)
	err := cmd.Run()
	if err != nil {
		log.Panicln(err)
	}
}

func createImage(dir string, img io.Reader) (*os.File, error) {
	file, err := os.Create(filepath.Join(dir, "picture.png"))
	if err != nil {
		return nil, err
	}
	log.Println("New file", file.Name())

	io.Copy(file, img)

	_, err = file.Seek(0, 0)
	if err != nil {
		defer file.Close()
		return nil, err
	}
	return file, nil
}

func useImageFuncs(imgData io.Reader, imgTarget io.Writer) error {
	img, err := png.Decode(imgData)
	if err != nil {
		return err
	}

	err = png.Encode(imgTarget, img)
	if err != nil {
		return err
	}

	return nil
}

type rainbowPNG struct{}

func (rainbowPNG) ColorModel() color.Model {
	return color.RGBAModel
}

var rect = image.Rectangle{
	Min: image.Point{0, 0},
	Max: image.Point{800, 600},
}

func (rainbowPNG) Bounds() image.Rectangle {
	return rect
}

func (rainbowPNG) At(x, y int) color.Color {
	r := x % 256
	g := (r + x) % 256
	b := (r + g + x) % 256

	return color.RGBA{
		A: 255,
		R: uint8(r),
		G: uint8(g),
		B: uint8(b),
	}
}
