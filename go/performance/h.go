package main

import (
	"fmt"
	"image"
	"os"
	"runtime"
	"sort"
	"strings"
	"time"

	_ "image/jpeg"
)

func loadImage(path string) image.Image {
	f, _ := os.Open(path)
	defer f.Close()
	img, _, _ := image.Decode(f)
	return img
}

func bar(ms float64) string {
	n := int(ms / 10)
	if n > 20 {
		n = 20
	}
	s := ""
	for i := 0; i < n; i++ {
		s += "█"
	}
	return s
}

type FilterConfig struct {
	fn   func(img image.Image, workers int) *image.RGBA
	maxW int
}

func main() {
	fmt.Println("🧪 STUDY 2 – Scaling Workers (Plafonds détectés)")
	fmt.Println("============================================")

	img := loadImage("input.jpg")

	// 3 FILTRES avec brushSize=5 pour OilPaint
	filters := map[string]FilterConfig{
		"GaussianBlur": {GaussianBlur, 64},
		"Sobel":        {Sobel, 16},
		"OilPaint": {func(img image.Image, workers int) *image.RGBA {
			return OilPaint(img, workers, 5) // brushSize=5
		}, 8},
	}

	for name, f := range filters {
		fmt.Printf("\n%s\n", name)
		fmt.Println(strings.Repeat("-", 40))

		tSeq := benchmark(f.fn, img, 1)
		fmt.Printf("SEQ: %6.1f ms | x1.00 | %s\n", tSeq, bar(tSeq))

		workersList := []int{2, 4, 8}
		if f.maxW >= 16 {
			workersList = append(workersList, 16)
		}
		if f.maxW >= 32 {
			workersList = append(workersList, 32)
		}
		if f.maxW >= 64 {
			workersList = append(workersList, 64)
		}

		bestSpeedup := 1.0
		bestW := 1

		for _, w := range workersList {
			tPar := benchmark(f.fn, img, w)
			speedup := tSeq / tPar

			if speedup > bestSpeedup {
				bestSpeedup = speedup
				bestW = w
			}

			fmt.Printf("W=%2d | %6.1f ms | x%.2f | %s\n",
				w, tPar, speedup, bar(tPar))
		}

		fmt.Printf("→ **PIC : W=%d (x%.2f)**\n\n", bestW, bestSpeedup)
	}
}

func benchmark(fn func(img image.Image, workers int) *image.RGBA,
	img image.Image, workers int) float64 {
	runtime.GC()

	times := make([]float64, 3)
	for i := 0; i < 3; i++ {
		runtime.GC()
		start := time.Now()
		fn(img, workers)
		times[i] = time.Since(start).Seconds() * 1000
	}
	sort.Float64s(times)
	return times[1]
}
