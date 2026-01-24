package main

import (
	"encoding/csv"
	"fmt"
	"image"
	"image/color"
	"math/rand"
	"os"
	"runtime"
	"sort"
	"strconv"
	"time"
)

// BenchmarkFilter = filtre + version SEQ
type BenchmarkFilter struct {
	Name     string
	Parallel func(img image.Image, workers int) *image.RGBA
	Seq      func(img image.Image) *image.RGBA
}

func main() {
	fmt.Println("🚀 BENCHMARKS FILTRAGE PARALLÈLE ELP")
	fmt.Println("=====================================")

	// Images de test (redimensionne input.jpg)
	images := loadTestImages()

	// TES 6 FILTRES (ajoute les tiens depuis filters.go/seq.go)
	filters := []BenchmarkFilter{
		{"Grayscale", Grayscale, GrayscaleSeq},
		{"GaussianBlur", GaussianBlur, GaussianBlurSeq},
		{"Sobel", Sobel, SobelSeq},
		{"Median", MedianFilter, MedianFilterSeq},
		// FIX Bilateral : wrappers sans params
		{"Bilateral", func(img image.Image, workers int) *image.RGBA {
			return BilateralFilter(img, workers, 1.0, 0.1)
		}, func(img image.Image) *image.RGBA {
			return BilateralFilterSeq(img, 1.0, 0.1)
		}},
		// FIX OilPaint : wrappers brushSize=5
		{"OilPaint", func(img image.Image, workers int) *image.RGBA {
			return OilPaint(img, workers, 5)
		}, func(img image.Image) *image.RGBA {
			return OilPaintSeq(img, 5)
		}},
	}
	// Workers à tester
	workersList := []int{1, 2, 4, 8, 16}

	// Lancement études
	runAllBenchmarks(images, filters, workersList)

	fmt.Println("\n✅ CSV sauvé: benchmarks.csv → Excel/Graphs !")
}

func loadTestImages() map[int]image.Image {
	fmt.Println("📥 Chargement images test...")
	images := make(map[int]image.Image)
	sizes := []int{512, 1024, 2048}

	var orig image.Image
	file, err := os.Open("input.jpg")
	if err == nil {
		defer file.Close()
		orig, _, err = image.Decode(file)
		if err != nil {
			fmt.Println("⚠️ input.jpg corrompu → noise")
		} else {
			fmt.Println("✅ input.jpg chargé")
		}
	} else {
		fmt.Println("⚠️ Pas input.jpg → noise coloré")
	}

	for _, size := range sizes {
		img := image.NewRGBA(image.Rect(0, 0, size, size))
		for y := 0; y < size; y++ {
			for x := 0; x < size; x++ {
				if orig != nil {
					sx := int(float64(x) * float64(orig.Bounds().Dx()) / float64(size))
					sy := int(float64(y) * float64(orig.Bounds().Dy()) / float64(size))
					img.Set(x, y, orig.At(sx, sy))
				} else {
					r := uint8(128 + rand.Intn(128))
					img.Set(x, y, color.RGBA{r, r / 2, r * 2 / 3, 255})
				}
			}
		}
		images[size] = img
		fmt.Printf("✅ %dx%d OK\n", size, size)
	}
	return images
}

// Mesure précise (3 runs, médiane)
func benchmark(name string, fn func(image.Image, int) *image.RGBA,
	img image.Image, workers int) float64 {
	times := make([]float64, 3)
	for i := 0; i < 3; i++ {
		runtime.GC() // Clean GC
		start := time.Now()
		fn(img, workers)
		times[i] = time.Since(start).Seconds() * 1000 // ms
	}
	sort.Float64s(times)
	return times[1] // Médiane
}

// Étude 1+2+3 COMPLÈTE
func runAllBenchmarks(images map[int]image.Image, filters []BenchmarkFilter, workersList []int) {
	// Collecte TOUS les données d'abord
	var csvData [][]string
	csvData = append(csvData, []string{"Size", "Filter", "Workers", "Time_ms", "Speedup"})

	// Headers console
	fmt.Println("\n📊 TABLEAUX COMPLETS")
	fmt.Println("Size | Filtre | SEQ | W1 | W2 | W4 | W8 | W16 | MaxSpeedup")
	fmt.Println("-----|--------|-----|----|----|----|----|----|-----------")

	for size, img := range images {
		fmt.Printf("%dx%d | ", size, size)

		for _, f := range filters {
			// 1. SEQ baseline
			tSeq := benchmark(f.Name, func(i image.Image, w int) *image.RGBA {
				return f.Seq(i)
			}, img, 1)

			maxSpeedup := 1.0
			fmt.Printf("%s ", f.Name[:8]) // Nom court

			// 2. PARALLEL scaling
			for _, workers := range workersList {
				tPar := benchmark(f.Name, f.Parallel, img, workers)
				speedup := tSeq / tPar
				if speedup > maxSpeedup {
					maxSpeedup = speedup
				}

				// Console courte
				fmt.Printf("|%3.0f ", tPar)

				// 3. CSV PROPRE
				csvData = append(csvData, []string{
					strconv.Itoa(size),
					f.Name,
					strconv.Itoa(workers),
					strconv.FormatFloat(tPar, 'f', 1, 64),
					strconv.FormatFloat(speedup, 'f', 2, 64),
				})
			}
			fmt.Printf("| x%.1f\n", maxSpeedup)
		}
	}

	// 4. SAUVE CSV FINAL
	file, err := os.Create("benchmarks.csv")
	if err != nil {
		fmt.Printf("❌ CSV erreur: %v\n", err)
		return
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	for _, row := range csvData {
		writer.Write(row)
	}
	writer.Flush()

	fmt.Printf("\n✅ %d lignes CSV sauvées ! Ouvrir benchmarks.csv → Graphs Excel\n", len(csvData))
}
