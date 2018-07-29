package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

func main() {
	file, err := createImage()
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
}

func createImage() (*os.File, error) {
	fmt.Println("Hello world")

	dir, err := ioutil.TempDir("", "img")
	if err != nil {
		return nil, err
	}

	file, err := os.Create(filepath.Join(dir, "picture.png"))
	if err != nil {
		return nil, err
	}
	log.Println("New file", file.Name())

	pic, err := os.Open("purple.png")
	if err != nil {
		return nil, err
	}
	defer pic.Close()

	io.Copy(file, pic)

	_, err = file.Seek(0, 0)
	if err != nil {
		return nil, err
	}
	return file, nil
}
