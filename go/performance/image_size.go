package main

import (
	"fmt"
	"image"
	"image/color"
	"time"
)

func genImage(size int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			img.Set(x, y, color.RGBA{uint8(x % 255), uint8(y % 255), 128, 255})
		}
	}
	return img
}

func main() {
	fmt.Println("🧪 STUDY 4 – Taille Image")

	sizes := []int{512, 1024, 2048}
	workers := 8

	for _, s := range sizes {
		img := genImage(s)

		start := time.Now()
		GaussianBlurSeq(img)
		tSeq := time.Since(start)

		start = time.Now()
		GaussianBlur(img, workers)
		tPar := time.Since(start)

		fmt.Printf("\n%d x %d\n", s, s)
		fmt.Printf("SEQ: %v\n", tSeq)
		fmt.Printf("PAR: %v\n", tPar)
		fmt.Printf("Speedup: x%.2f\n",
			float64(tSeq)/float64(tPar))
	}
}
