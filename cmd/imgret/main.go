package main

import (
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

	pic, err := os.Open("purple.png")
	if err != nil {
		log.Panicln(err)
	}
	defer pic.Close()

	file, err := createImage(tempDir, pic)
	if err != nil {
		log.Panicln(err)
	}
	err = file.Close()
	if err != nil {
		log.Panicln(err)
	}

	cmd := exec.Command("xdg-open", file.Name())
	err = cmd.Run()
	if err != nil {
		log.Panicln(err)
	}

	out, err := os.Create(filepath.Join(tempDir, "picture2.png"))
	if err != nil {
		log.Panicln(err)
	}
	defer out.Close()

	if _, err = pic.Seek(0, 0); err != nil {
		log.Panicln("seek", err)
	}

	err = useImageFuncs(pic, out)
	if err != nil {
		log.Panicln("useImageFuncs", err)
	}

	out.Close()
	cmd2 := exec.Command("xdg-open", out.Name())
	err = cmd2.Run()
	if err != nil {
		log.Panicln("cmd2", err)
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
