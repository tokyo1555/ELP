// parallel.go
package main

import (
	"image"
	"image/color"
	"math"
	"sort"
	"sync"
)

//
// =========================
// OUTILS COMMUNS
// =========================
//

// splitWorkers adapte le nombre de workers à la hauteur de l'image et calcule un découpage par bandes horizontales.
func splitWorkers(bounds image.Rectangle, workers int) (w int, block int, height int) {
	height = bounds.Max.Y - bounds.Min.Y

	if workers < 1 {
		workers = 1
	}
	if workers > height {
		workers = height
	}
	block = height / workers
	return workers, block, height
}

//
// =========================
// FILTRES "PIXEL PAR PIXEL"
// (un pixel de sortie dépend surtout du pixel d'entrée au même endroit)
// =========================
//

// Grayscale convertit l'image en niveaux de gris en parallèle.
func Grayscale(img image.Image, workers int) *image.RGBA {
	bounds := img.Bounds()
	result := image.NewRGBA(bounds)

	w, block, _ := splitWorkers(bounds, workers)
	var wg sync.WaitGroup

	for i := 0; i < w; i++ {
		startY := bounds.Min.Y + i*block
		endY := startY + block
		if i == w-1 {
			endY = bounds.Max.Y
		}

		wg.Add(1)
		go func(startY, endY int) {
			defer wg.Done()
			for y := startY; y < endY; y++ {
				for x := bounds.Min.X; x < bounds.Max.X; x++ {
					r, g, b, _ := img.At(x, y).RGBA()
					gray16 := (r + g + b) / 3
					avg := uint8(gray16 >> 8)
					result.Set(x, y, color.RGBA{avg, avg, avg, 255})
				}
			}
		}(startY, endY)
	}

	wg.Wait()
	return result
}

// Invert applique un négatif (255 - composante) en parallèle.
func Invert(img image.Image, workers int) *image.RGBA {
	bounds := img.Bounds()
	result := image.NewRGBA(bounds)

	w, block, _ := splitWorkers(bounds, workers)
	var wg sync.WaitGroup

	for i := 0; i < w; i++ {
		startY := bounds.Min.Y + i*block
		endY := startY + block
		if i == w-1 {
			endY = bounds.Max.Y
		}

		wg.Add(1)
		go func(startY, endY int) {
			defer wg.Done()
			for y := startY; y < endY; y++ {
				for x := bounds.Min.X; x < bounds.Max.X; x++ {
					r, g, b, _ := img.At(x, y).RGBA()
					result.Set(x, y, color.RGBA{
						255 - uint8(r>>8),
						255 - uint8(g>>8),
						255 - uint8(b>>8),
						255,
					})
				}
			}
		}(startY, endY)
	}

	wg.Wait()
	return result
}

//
// =========================
// FILTRES "VOISINAGE"
// (un pixel de sortie dépend d'un voisinage autour du pixel d'entrée)
// =========================
//

// Blur applique un flou "box blur" de rayon donné (radius >= 1) en parallèle.
// radius = 1 -> ~3x3, radius = 5 -> ~11x11, etc.
func Blur(img image.Image, workers int, radius int) *image.RGBA {
	bounds := img.Bounds()
	result := image.NewRGBA(bounds)

	if radius < 1 {
		radius = 1
	}

	w, block, _ := splitWorkers(bounds, workers)
	var wg sync.WaitGroup

	for i := 0; i < w; i++ {
		startY := bounds.Min.Y + i*block
		endY := startY + block
		if i == w-1 {
			endY = bounds.Max.Y
		}

		wg.Add(1)
		go func(startY, endY int) {
			defer wg.Done()

			for y := startY; y < endY; y++ {
				for x := bounds.Min.X; x < bounds.Max.X; x++ {

					var sumR, sumG, sumB uint32
					var count uint32

					// Voisinage (2*radius+1) x (2*radius+1)
					for ny := y - radius; ny <= y+radius; ny++ {
						if ny < bounds.Min.Y || ny >= bounds.Max.Y {
							continue
						}
						for nx := x - radius; nx <= x+radius; nx++ {
							if nx < bounds.Min.X || nx >= bounds.Max.X {
								continue
							}
							r, g, b, _ := img.At(nx, ny).RGBA()
							sumR += r
							sumG += g
							sumB += b
							count++
						}
					}

					if count == 0 {
						result.Set(x, y, img.At(x, y))
						continue
					}

					avgR := uint8((sumR / count) >> 8)
					avgG := uint8((sumG / count) >> 8)
					avgB := uint8((sumB / count) >> 8)

					result.Set(x, y, color.RGBA{avgR, avgG, avgB, 255})
				}
			}
		}(startY, endY)
	}

	wg.Wait()
	return result
}

// GaussianBlur applique un flou gaussien 5x5 (rayon 2) en parallèle.
func GaussianBlur(img image.Image, workers int) *image.RGBA {
	kernel := [][]float64{
		{1, 4, 6, 4, 1},
		{4, 16, 24, 16, 4},
		{6, 24, 36, 24, 6},
		{4, 16, 24, 16, 4},
		{1, 4, 6, 4, 1},
	}
	const kernelSum = 256.0
	radius := 2

	bounds := img.Bounds()
	out := image.NewRGBA(bounds)

	w, block, _ := splitWorkers(bounds, workers)
	var wg sync.WaitGroup

	for i := 0; i < w; i++ {
		wg.Add(1)
		startY := bounds.Min.Y + i*block
		endY := startY + block
		if i == w-1 {
			endY = bounds.Max.Y
		}

		go func(startY, endY int) {
			defer wg.Done()

			for y := startY; y < endY; y++ {
				for x := bounds.Min.X + radius; x < bounds.Max.X-radius; x++ {
					var rSum, gSum, bSum float64

					for ky := -radius; ky <= radius; ky++ {
						for kx := -radius; kx <= radius; kx++ {
							if y+ky < bounds.Min.Y || y+ky >= bounds.Max.Y ||
								x+kx < bounds.Min.X || x+kx >= bounds.Max.X {
								continue
							}
							r, g, b, _ := img.At(x+kx, y+ky).RGBA()
							weight := kernel[ky+radius][kx+radius]
							rSum += float64(r>>8) * weight
							gSum += float64(g>>8) * weight
							bSum += float64(b>>8) * weight
						}
					}

					out.Set(x, y, color.RGBA{
						uint8(rSum / kernelSum),
						uint8(gSum / kernelSum),
						uint8(bSum / kernelSum),
						255,
					})
				}
			}
		}(startY, endY)
	}

	wg.Wait()
	return out
}

// Sobel détecte les contours (approx gradient) en parallèle.
func Sobel(img image.Image, workers int) *image.RGBA {
	bounds := img.Bounds()
	out := image.NewRGBA(bounds)

	gx := [][]int{
		{-1, 0, 1},
		{-2, 0, 2},
		{-1, 0, 1},
	}
	gy := [][]int{
		{1, 2, 1},
		{0, 0, 0},
		{-1, -2, -1},
	}

	w, block, _ := splitWorkers(bounds, workers)
	var wg sync.WaitGroup

	for i := 0; i < w; i++ {
		wg.Add(1)
		startY := bounds.Min.Y + i*block
		endY := startY + block
		if i == w-1 {
			endY = bounds.Max.Y
		}

		go func(startY, endY int) {
			defer wg.Done()

			for y := startY; y < endY; y++ {
				for x := bounds.Min.X + 1; x < bounds.Max.X-1; x++ {
					var sumX, sumY int

					for ky := -1; ky <= 1; ky++ {
						for kx := -1; kx <= 1; kx++ {
							if y+ky < bounds.Min.Y || y+ky >= bounds.Max.Y ||
								x+kx < bounds.Min.X || x+kx >= bounds.Max.X {
								continue
							}
							r, _, _, _ := img.At(x+kx, y+ky).RGBA()
							gray := int(r >> 8)
							sumX += gray * gx[ky+1][kx+1]
							sumY += gray * gy[ky+1][kx+1]
						}
					}

					magnitude := uint8(math.Min(
						255,
						math.Sqrt(float64(sumX*sumX+sumY*sumY)),
					))

					out.Set(x, y, color.RGBA{magnitude, magnitude, magnitude, 255})
				}
			}
		}(startY, endY)
	}

	wg.Wait()
	return out
}

// MedianFilter applique un filtre médian 3x3 (réduction du bruit impulsionnel) en parallèle.
func MedianFilter(img image.Image, workers int) *image.RGBA {
	bounds := img.Bounds()
	out := image.NewRGBA(bounds)

	w, block, _ := splitWorkers(bounds, workers)
	var wg sync.WaitGroup

	for i := 0; i < w; i++ {
		startY := bounds.Min.Y + i*block
		endY := startY + block
		if i == w-1 {
			endY = bounds.Max.Y
		}

		wg.Add(1)
		go func(startY, endY int) {
			defer wg.Done()

			for y := startY; y < endY; y++ {
				for x := bounds.Min.X; x < bounds.Max.X; x++ {
					var reds, greens, blues [9]uint8
					idx := 0

					for dy := -1; dy <= 1; dy++ {
						for dx := -1; dx <= 1; dx++ {
							nx, ny := x+dx, y+dy
							if nx >= bounds.Min.X && nx < bounds.Max.X &&
								ny >= bounds.Min.Y && ny < bounds.Max.Y {
								r, g, b, _ := img.At(nx, ny).RGBA()
								reds[idx] = uint8(r >> 8)
								greens[idx] = uint8(g >> 8)
								blues[idx] = uint8(b >> 8)
							}
							idx++
						}
					}

					sort.Slice(reds[:], func(i, j int) bool { return reds[i] < reds[j] })
					sort.Slice(greens[:], func(i, j int) bool { return greens[i] < greens[j] })
					sort.Slice(blues[:], func(i, j int) bool { return blues[i] < blues[j] })

					out.Set(x, y, color.RGBA{reds[4], greens[4], blues[4], 255})
				}
			}
		}(startY, endY)
	}

	wg.Wait()
	return out
}

// OilPaint applique un effet "peinture à l'huile" (couleur dominante dans un pinceau) en parallèle.
func OilPaint(img image.Image, workers int, brushSize int) *image.RGBA {
	bounds := img.Bounds()
	out := image.NewRGBA(bounds)

	if brushSize < 3 {
		brushSize = 3
	}
	radius := brushSize / 2

	w, block, _ := splitWorkers(bounds, workers)
	var wg sync.WaitGroup

	for i := 0; i < w; i++ {
		startY := bounds.Min.Y + i*block
		endY := startY + block
		if i == w-1 {
			endY = bounds.Max.Y
		}

		wg.Add(1)
		go func(startY, endY int) {
			defer wg.Done()

			for y := startY; y < endY; y++ {
				for x := bounds.Min.X; x < bounds.Max.X; x++ {
					// Histogramme des couleurs dans le pinceau
					counts := make(map[color.RGBA]int)

					for dy := -radius; dy <= radius; dy++ {
						for dx := -radius; dx <= radius; dx++ {
							nx, ny := x+dx, y+dy
							if nx >= bounds.Min.X && nx < bounds.Max.X &&
								ny >= bounds.Min.Y && ny < bounds.Max.Y {
								r, g, b, _ := img.At(nx, ny).RGBA()
								c := color.RGBA{
									R: uint8(r >> 8),
									G: uint8(g >> 8),
									B: uint8(b >> 8),
									A: 255,
								}
								counts[c]++
							}
						}
					}

					// Couleur dominante (mode)
					var mostCommon color.RGBA
					maxCount := 0
					for c, count := range counts {
						if count > maxCount {
							maxCount = count
							mostCommon = c
						}
					}

					out.Set(x, y, mostCommon)
				}
			}
		}(startY, endY)
	}

	wg.Wait()
	return out
}

//
// =========================
// FILTRES "BLOCS"
// (on traite des zones rectangulaires, utile pour la mosaïque)
// =========================
//

// Pixelate applique un effet mosaïque (pixelation) en parallèle.
// blockSize = taille des blocs (>= 2).
func Pixelate(img image.Image, workers int, blockSize int) *image.RGBA {
	bounds := img.Bounds()
	out := image.NewRGBA(bounds)

	if blockSize < 2 {
		blockSize = 2
	}

	w, block, _ := splitWorkers(bounds, workers)
	var wg sync.WaitGroup

	for i := 0; i < w; i++ {
		startY := bounds.Min.Y + i*block
		endY := startY + block
		if i == w-1 {
			endY = bounds.Max.Y
		}

		wg.Add(1)
		go func(startY, endY int) {
			defer wg.Done()

			// On avance par pas de blockSize
			for y := startY; y < endY; y += blockSize {
				for x := bounds.Min.X; x < bounds.Max.X; x += blockSize {

					var sumR, sumG, sumB uint32
					var count uint32

					// 1) moyenne du bloc
					for yy := y; yy < y+blockSize && yy < bounds.Max.Y; yy++ {
						for xx := x; xx < x+blockSize && xx < bounds.Max.X; xx++ {
							r, g, b, _ := img.At(xx, yy).RGBA()
							sumR += r >> 8
							sumG += g >> 8
							sumB += b >> 8
							count++
						}
					}
					if count == 0 {
						continue
					}

					avg := color.RGBA{
						R: uint8(sumR / count),
						G: uint8(sumG / count),
						B: uint8(sumB / count),
						A: 255,
					}

					// 2) remplissage du bloc
					for yy := y; yy < y+blockSize && yy < bounds.Max.Y; yy++ {
						for xx := x; xx < x+blockSize && xx < bounds.Max.X; xx++ {
							out.Set(xx, yy, avg)
						}
					}
				}
			}
		}(startY, endY)
	}

	wg.Wait()
	return out
}
