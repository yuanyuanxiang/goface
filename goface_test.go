package main

import (
	"fmt"
	"image"
	"image/jpeg"
	"os"
	"testing"

	pigo "github.com/esimov/pigo/core"
)

// save the result of DetectFace.
func save(src *image.NRGBA, dets []pigo.Detection) []image.Image {
	for i, v := range dets {
		x, y, w := v.Col, v.Row, v.Scale/2
		img := src.SubImage(image.Rect(x-w, y-w, x+w, y+w))
		file, err := os.Create(fmt.Sprintf("%d-%.2f", i, v.Q) + ".jpg")
		if err != nil {
			fmt.Println(err)
			continue
		}
		defer file.Close()
		if err = jpeg.Encode(file, img, &jpeg.Options{Quality: 100}); err != nil {
			fmt.Println(err)
		}
	}
	return nil
}

// https://github.com/esimov/pigo
func TestGoface(t *testing.T) {
	initClassifier()
	file, err := os.Open("./test.jpg")
	if err != nil {
		fmt.Println(err)
	}
	defer file.Close()
	img, err := jpeg.Decode(file)
	if err != nil {
		fmt.Println(err)
	}
	DetectFace(img, save)
}
