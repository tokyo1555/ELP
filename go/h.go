package main

import (
	"fmt"
	"image"
)

type BenchmarkFilter struct {
	Name     string
	Parallel func(img image.Image, workers int) *image.RGBA
	Seq      func(img image.Image) *image.RGBA
}

type BenchmarkData struct {
	Size      int     `json:"size"`
	Filter    string  `json:"filter"`
	Workers   int     `json:"workers"`
	TimeMs    float64 `json:"time_ms"`
	Speedup   float64 `json:"speedup"`
	BlockSize int     `json:"blocksize,omitempty"` // Pour Pixelate
}

type ChartData struct {
	Size     int
	Filter   string
	Seq      float64
	Workers  []int
	Times    []float64
	Speedups []float64
}

func main() {
	fmt.Println("🚀 BENCHMARKS FILTRAGE PARALLÈLE ELP")
	fmt.Println("====================================")

	images := loadTestImages()

	// TES FILTRES CORRIGÉS (Pixelate à la place de Bilateral)
	filters := []BenchmarkFilter{
		{"Grayscale", Grayscale, GrayscaleSeq},
		{"GaussianBlur", GaussianBlur, GaussianBlurSeq},
		{"Sobel", Sobel, SobelSeq},
		{"Median", MedianFilter, MedianFilterSeq},
		// PIXELATE avec blockSize=16 (fixe pour fair-play benchmark)
		{"Pixelate", func(img image.Image, workers int) *image.RGBA {
			return Pixelate(img, workers, 16) // blockSize=16 fixe
		}, func(img image.Image) *image.RGBA {
			return PixelateSeq(img, 16) // blockSize=16 fixe
		}},
		{"OilPaint", func(img image.Image, workers int) *image.RGBA {
			return OilPaint(img, workers, 5)
		}, func(img image.Image) *image.RGBA {
			return OilPaintSeq(img, 5)
		}},
	}
	workersList := []int{1, 2, 4, 8, 16}

	allData := runAllBenchmarks(images, filters, workersList)

	fmt.Println("\n📊 Génération 4 études graphiques...")
	generateTimeCharts(allData, workersList)
	generateSpeedupCharts(allData, workersList)
	generateComparisonChart(allData)
	generateFullHTMLReport(allData, workersList)

	fmt.Println("\n✅ Graphiques sauvés → Copie dans rapport ELP !")
	fmt.Println("   📈 results_time/*.png")
	fmt.Println("   📈 results_speedup/*.png")
	fmt.Println("   📊 comparison.png")
	fmt.Println("   🌐 results.html ← MEILLEUR !")
}
