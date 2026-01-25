package main

import (
	"image"
	"image/color"
	"math"
	"sort"
)

// GRAYSCALE
func GrayscaleSeq(img image.Image) *image.RGBA {
	bounds := img.Bounds()
	result := image.NewRGBA(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			gray16 := (r + g + b) / 3
			avg := uint8(gray16 >> 8)
			result.Set(x, y, color.RGBA{avg, avg, avg, 255})
		}
	}

	return result
}

// INVERT
func InvertSeq(img image.Image) *image.RGBA {
	bounds := img.Bounds()
	out := image.NewRGBA(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			out.Set(x, y, color.RGBA{
				255 - uint8(r>>8),
				255 - uint8(g>>8),
				255 - uint8(b>>8),
				255,
			})
		}
	}
	return out
}

// BlurSeq applique un flou "box blur" de manière séquentielle.
func BlurSeq(img image.Image, radius int) *image.RGBA {
	bounds := img.Bounds()
	result := image.NewRGBA(bounds)

	if radius < 1 {
		radius = 1
	}

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {

			var sumR, sumG, sumB uint32
			var count uint32

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

	return result
}

func GaussianBlurSeq(img image.Image) *image.RGBA {
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

	for y := bounds.Min.Y + radius; y < bounds.Max.Y-radius; y++ {
		for x := bounds.Min.X + radius; x < bounds.Max.X-radius; x++ {

			var rSum, gSum, bSum float64

			for ky := -radius; ky <= radius; ky++ {
				for kx := -radius; kx <= radius; kx++ {
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
	return out
}

func SobelSeq(img image.Image) *image.RGBA {
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

	for y := bounds.Min.Y + 1; y < bounds.Max.Y-1; y++ {
		for x := bounds.Min.X + 1; x < bounds.Max.X-1; x++ {

			var sumX, sumY int

			for ky := -1; ky <= 1; ky++ {
				for kx := -1; kx <= 1; kx++ {
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
	return out
}

func MedianFilterSeq(img image.Image) *image.RGBA {
	bounds := img.Bounds()
	out := image.NewRGBA(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			// Collecte 9 voisins dans 3x3
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

			// Tri + médiane
			sort.Slice(reds[:], func(i, j int) bool { return reds[i] < reds[j] })
			sort.Slice(greens[:], func(i, j int) bool { return greens[i] < greens[j] })
			sort.Slice(blues[:], func(i, j int) bool { return blues[i] < blues[j] })

			medianR, medianG, medianB := reds[4], greens[4], blues[4]
			out.Set(x, y, color.RGBA{medianR, medianG, medianB, 255})
		}
	}
	return out
}

func OilPaintSeq(img image.Image, brushSize int) *image.RGBA {
	bounds := img.Bounds()
	out := image.NewRGBA(bounds)

	radius := brushSize / 2

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
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

			// Couleur la plus fréquente (mode)
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
	return out
}
