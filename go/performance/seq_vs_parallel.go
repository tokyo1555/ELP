package main

import (
	"fmt"
	"image"
	"os"
	"time"

	_ "image/jpeg"
	_ "image/png"
)

func loadImage(path string) image.Image {
	f, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	img, _, err := image.Decode(f)
	if err != nil {
		panic(err)
	}
	return img
}

func main() {
	fmt.Println("🧪 STUDY 1 – SEQ vs PAR")

	img := loadImage("input.jpg")
	workers := 8

	start := time.Now()
	_ = GaussianBlurSeq(img)
	tSeq := time.Since(start)

	start = time.Now()
	_ = GaussianBlur(img, workers)
	tPar := time.Since(start)

	fmt.Println("\nGaussian Blur")
	fmt.Printf("SEQ : %v\n", tSeq)
	fmt.Printf("PAR (%d workers): %v\n", workers, tPar)
	fmt.Printf("Speedup: x%.2f\n", float64(tSeq)/float64(tPar))

	start = time.Now()
	_ = SobelSeq(img)
	tSeq = time.Since(start)

	start = time.Now()
	_ = Sobel(img, workers)
	tPar = time.Since(start)

	fmt.Println("\nSobel")
	fmt.Printf("SEQ : %v\n", tSeq)
	fmt.Printf("PAR (%d workers): %v\n", workers, tPar)
	fmt.Printf("Speedup: x%.2f\n", float64(tSeq)/float64(tPar))

	start = time.Now()
	_ = PixelateSeq(img, 5)
	tSeq = time.Since(start)

	start = time.Now()
	_ = Pixelate(img, workers, 5)
	tPar = time.Since(start)

	fmt.Println("\nPixelate block size = 5")
	fmt.Printf("SEQ : %v\n", tSeq)
	fmt.Printf("PAR (%d workers): %v\n", workers, tPar)
	fmt.Printf("Speedup: x%.2f\n", float64(tSeq)/float64(tPar))

	start = time.Now()
	_ = OilPaintSeq(img, 5)
	tSeq = time.Since(start)

	start = time.Now()
	_ = OilPaint(img, workers, 5)
	tPar = time.Since(start)

	fmt.Println("\nOil Paint brush size = 5")
	fmt.Printf("SEQ : %v\n", tSeq)
	fmt.Printf("PAR (%d workers): %v\n", workers, tPar)
	fmt.Printf("Speedup: x%.2f\n", float64(tSeq)/float64(tPar))
}
