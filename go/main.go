package main

import (
	"fmt"
	"image"
	"image/jpeg"
	"os"
	"time"
)

func main() {
	// Charge image
	file, err := os.Open("input.jpg")
	if err != nil {
		panic("Place input.jpg dans le dossier")
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		panic("Erreur image")
	}
	fmt.Println("🖼️  Test des 3 nouveaux filtres SÉQUENTIELS")
	fmt.Println("⏱️   Mesure temps + sauvegarde images...\n")

	// 1. MEDIAN
	start := time.Now()
	median := OilPaintSeq(img, 5)
	fmt.Printf("✅ Median SEQ: %.0fms → output_median.jpg\n", time.Since(start).Seconds()*1000)
	saveImage(median, "output_median.jpg")

	// 1. MEDIAN PARALLÈLE
	start1 := time.Now()
	medianPar := OilPaint(img, 8, 5)
	fmt.Printf("🚀 Median Parallel W8: %.0fms → output_median_w8.jpg\n", time.Since(start1).Seconds()*1000)
	saveImage(medianPar, "output_median_w8.jpg")

	// 2. BILATERAL PARALLÈLE

}

func saveImage(img *image.RGBA, filename string) {
	f, _ := os.Create(filename)
	defer f.Close()
	jpeg.Encode(f, img, &jpeg.Options{Quality: 95})
}

// ApplyFilter choisit et applique le bon filtre en fonction du nom.
// name : "grayscale", "invert", "blur"
// workers : nombre de goroutines internes
// radius : utilisé seulement pour "blur"
func ApplyFilter(img image.Image, name string, workers int, radius int) (*image.RGBA, error) {
	switch name {
	case "grayscale":
		return Grayscale(img, workers), nil
	case "invert":
		return Invert(img, workers), nil
	case "blur":
		return Blur(img, workers, radius), nil
	default:
		return nil, fmt.Errorf("filtre non reconnu: %s", name)
	}
}
