package main

import (
	"fmt"
	"image"
	"os"
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
	s := ""
	for i := 0; i < n; i++ {
		s += "█"
	}
	return s
}

func main() {
	fmt.Println("🧪 STUDY 2 – Scaling Workers")
	fmt.Println("OilPaint filter")
	img := loadImage("input.jpg")
	workersList := []int{1, 2, 4, 8, 16, 32, 64}

	var tSeq float64

	for _, w := range workersList {
		start := time.Now()
		_ = Sobel(img, w)
		t := time.Since(start).Seconds() * 1000

		if w == 1 {
			tSeq = t
		}

		speedup := tSeq / t
		fmt.Printf("W=%2d | %6.1f ms | x%.2f | %s\n",
			w, t, speedup, bar(t))
	}
}
