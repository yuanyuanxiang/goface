package goface

import (
	"fmt"
	"image/jpeg"
	"os"
	"testing"
)

// https://github.com/esimov/pigo
func TestGoface(t *testing.T) {
	file, err := os.Open("./test.jpg")
	if err != nil {
		fmt.Println(err)
	}
	defer file.Close()
	img, err := jpeg.Decode(file)
	if err != nil {
		fmt.Println(err)
	}
	DetectFace(img)
}
