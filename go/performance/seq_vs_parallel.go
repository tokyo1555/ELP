package main

import (
	"fmt"
	"image"
	"os"
	"time"

	_ "image/jpeg"
	_ "image/png"
)

// --------------------
// Chargement image
// --------------------
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

// --------------------
// MAIN
// --------------------
func main() {
	fmt.Println("🧪 STUDY – SEQ vs PAR (workers fixes)")
	fmt.Println("===================================")

	img := loadImage("input.jpg")
	workers := 8

	var tSeq, tPar time.Duration

	// =========================
	// GRAYSCALE
	// =========================
	start := time.Now()
	_ = GrayscaleSeq(img)
	tSeq = time.Since(start)

	start = time.Now()
	_ = Grayscale(img, workers)
	tPar = time.Since(start)

	fmt.Println("\nGrayscale")
	fmt.Printf("SEQ : %.4f s\n", tSeq.Seconds())
	fmt.Printf("PAR (%d workers) : %.4f s\n", workers, tPar.Seconds())
	fmt.Printf("Speedup : x%.2f\n", tSeq.Seconds()/tPar.Seconds())

	// =========================
	// GAUSSIAN BLUR
	// =========================
	start = time.Now()
	_ = GaussianBlurSeq(img)
	tSeq = time.Since(start)

	start = time.Now()
	_ = GaussianBlur(img, workers)
	tPar = time.Since(start)

	fmt.Println("\nGaussian Blur")
	fmt.Printf("SEQ : %.4f s\n", tSeq.Seconds())
	fmt.Printf("PAR (%d workers) : %.4f s\n", workers, tPar.Seconds())
	fmt.Printf("Speedup : x%.2f\n", tSeq.Seconds()/tPar.Seconds())

	// =========================
	// SOBEL EDGE DETECTION
	// =========================
	start = time.Now()
	_ = SobelSeq(img)
	tSeq = time.Since(start)

	start = time.Now()
	_ = Sobel(img, workers)
	tPar = time.Since(start)

	fmt.Println("\nSobel Edge Detection")
	fmt.Printf("SEQ : %.4f s\n", tSeq.Seconds())
	fmt.Printf("PAR (%d workers) : %.4f s\n", workers, tPar.Seconds())
	fmt.Printf("Speedup : x%.2f\n", tSeq.Seconds()/tPar.Seconds())

	// =========================
	// PIXELATE
	// =========================
	blockSize := 5

	start = time.Now()
	_ = PixelateSeq(img, blockSize)
	tSeq = time.Since(start)

	start = time.Now()
	_ = Pixelate(img, workers, blockSize)
	tPar = time.Since(start)

	fmt.Println("\nPixelate (block size = 5)")
	fmt.Printf("SEQ : %.4f s\n", tSeq.Seconds())
	fmt.Printf("PAR (%d workers) : %.4f s\n", workers, tPar.Seconds())
	fmt.Printf("Speedup : x%.2f\n", tSeq.Seconds()/tPar.Seconds())

	// =========================
	// OIL PAINT
	// =========================
	brushSize := 5

	start = time.Now()
	_ = OilPaintSeq(img, brushSize)
	tSeq = time.Since(start)

	start = time.Now()
	_ = OilPaint(img, workers, brushSize)
	tPar = time.Since(start)

	fmt.Println("\nOil Paint (brush size = 5)")
	fmt.Printf("SEQ : %.4f s\n", tSeq.Seconds())
	fmt.Printf("PAR (%d workers) : %.4f s\n", workers, tPar.Seconds())
	fmt.Printf("Speedup : x%.2f\n", tSeq.Seconds()/tPar.Seconds())

}
