package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

func main() {
	fmt.Println("Hello world")

	dir, err := ioutil.TempDir("", "img")
	if err != nil {
		log.Panicln(err)
	}

	file, err := os.Create(filepath.Join(dir, "picture.png"))
	if err != nil {
		log.Panicln(err)
	}
	defer file.Close()

	log.Println("New file", file.Name())

	pic, err := os.Open("purple.png")
	if err != nil {
		log.Panicln(err)
	}
	defer pic.Close()

	log.Println("Copying")
	io.Copy(file, pic)
}
