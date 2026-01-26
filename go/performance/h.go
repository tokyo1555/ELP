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

// ==========================
// UTILS
// ==========================

func loadImage(path string) image.Image {
	f, err := os.Open(path)
	if err != nil {
		panic("image introuvable")
	}
	defer f.Close()
	img, _, err := image.Decode(f)
	if err != nil {
		panic("image invalide")
	}
	return img
}

func bar(ms float64) string {
	n := int(ms / 10)
	if n > 25 {
		n = 25
	}
	s := ""
	for i := 0; i < n; i++ {
		s += "█"
	}
	return s
}

func benchmark(fn func(image.Image, int) *image.RGBA,
	img image.Image, workers int) float64 {

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

// ==========================
// STRUCT
// ==========================

type FilterConfig struct {
	fn func(img image.Image, workers int) *image.RGBA
}

// ==========================
// MAIN
// ==========================

func main() {
	fmt.Println("🧪 STUDY 2 – Détection AUTOMATIQUE du plafond")
	fmt.Println("============================================")

	img := loadImage("input.jpg")

	filters := map[string]FilterConfig{
		"GaussianBlur": {GaussianBlur},
		"Sobel":        {Sobel},
		"OilPaint": {func(img image.Image, workers int) *image.RGBA {
			return OilPaint(img, workers, 5)
		}},
		"Pixelate": {func(img image.Image, workers int) *image.RGBA {
			return Pixelate(img, workers, 5)
		}},
	}

	for name, f := range filters {
		fmt.Printf("\n%s\n", name)
		fmt.Println(strings.Repeat("-", 45))

		// --- SEQ ---
		tSeq := benchmark(f.fn, img, 1)
		fmt.Printf("SEQ | %6.1f ms | x1.00 | %s\n", tSeq, bar(tSeq))

		prevTime := tSeq
		bestTime := tSeq
		bestW := 1

		workers := 2

		for {
			tPar := benchmark(f.fn, img, workers)
			speedup := tSeq / tPar

			fmt.Printf("W=%2d | %6.1f ms | x%.2f | %s\n",
				workers, tPar, speedup, bar(tPar))

			if tPar < bestTime {
				bestTime = tPar
				bestW = workers
			}

			// 🔴 PLAFOND DÉTECTÉ
			if tPar >= prevTime {
				fmt.Printf("⛔ STOP : temps augmente (plafond atteint)\n")
				break
			}

			prevTime = tPar
			workers *= 2

			if workers > runtime.NumCPU()*4 {
				fmt.Printf("⛔ STOP : limite CPU atteinte\n")
				break
			}
		}

		fmt.Printf("🏁 PIC OPTIMAL : W=%d | %.1f ms | x%.2f\n",
			bestW, bestTime, tSeq/bestTime)
	}
}
