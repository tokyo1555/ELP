// parallel.go
package main

import (
	"image"
	"image/color"
	"math"
	"sort"
	"sync"
	"image/draw"
)

// splitWorkers adapte le nombre de workers a la hauteur de l'image et calcule un découpage par bandes horizontales
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

// Grayscale convertit l'image en niveaux de gris
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

// Blur applique un flou "box blur" de rayon donné (radius >= 1)
// radius = 1 -> ~3x3 ect...
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

// Sobel détecte les contours (approx gradient)
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

// MedianFilter applique un filtre médian 3x3 (réduction du bruit impulsionnel)
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

// Pixelate applique un effet mosaïque (pixelation)
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

// PosterizeQuantilesColor applique une posterization couleur basée sur des quantiles globaux,
// séparément sur R, G et B.
// levels = nombre de niveaux par canal (>=2). Couleurs possibles ~ levels^3.
// Complexité : O(N log N) (3 tris : R,G,B).
func PosterizeQuantilesColor(img image.Image, workers int, levels int) *image.RGBA {
	if levels < 2 {
		levels = 2
	}

	bounds := img.Bounds()
	wImg := bounds.Dx()
	hImg := bounds.Dy()

	// Conversion en RGBA pour accès rapide
	src := image.NewRGBA(bounds)
	draw.Draw(src, bounds, img, bounds.Min, draw.Src)

	// 1) Récupérer R,G,B en parallèle (tableaux de taille N)
	n := wImg * hImg
	rs := make([]uint8, n)
	gs := make([]uint8, n)
	bs := make([]uint8, n)

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
				yy := y - bounds.Min.Y
				base := yy * wImg
				for x := bounds.Min.X; x < bounds.Max.X; x++ {
					xx := x - bounds.Min.X
					pi := src.PixOffset(x, y)
					idx := base + xx
					rs[idx] = src.Pix[pi+0]
					gs[idx] = src.Pix[pi+1]
					bs[idx] = src.Pix[pi+2]
				}
			}
		}(startY, endY)
	}
	wg.Wait()

	// 2) Tri global (dominant) et LUT par canal
	sr := make([]uint8, n)
	sg := make([]uint8, n)
	sb := make([]uint8, n)
	copy(sr, rs)
	copy(sg, gs)
	copy(sb, bs)

	sort.Slice(sr, func(i, j int) bool { return sr[i] < sr[j] })
	sort.Slice(sg, func(i, j int) bool { return sg[i] < sg[j] })
	sort.Slice(sb, func(i, j int) bool { return sb[i] < sb[j] })

	lutR := buildQuantileLUT(sr, levels)
	lutG := buildQuantileLUT(sg, levels)
	lutB := buildQuantileLUT(sb, levels)

	// 3) Application parallèle
	out := image.NewRGBA(bounds)

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
				yy := y - bounds.Min.Y
				base := yy * wImg
				for x := bounds.Min.X; x < bounds.Max.X; x++ {
					xx := x - bounds.Min.X
					idx := base + xx

					pi := src.PixOffset(x, y)
					a := src.Pix[pi+3]

					rq := lutR[rs[idx]]
					gq := lutG[gs[idx]]
					bq := lutB[bs[idx]]

					di := out.PixOffset(x, y)
					out.Pix[di+0] = rq
					out.Pix[di+1] = gq
					out.Pix[di+2] = bq
					out.Pix[di+3] = a
				}
			}
		}(startY, endY)
	}
	wg.Wait()

	return out
}

// buildQuantileLUT construit une table 0..255 -> niveau représentatif basé sur les quantiles.
func buildQuantileLUT(sortedVals []uint8, levels int) [256]uint8 {
	var lut [256]uint8
	n := len(sortedVals)

	if n == 0 || levels < 2 {
		for v := 0; v < 256; v++ {
			lut[v] = uint8(v)
		}
		return lut
	}
	if levels > n {
		levels = n
	}

	reps := make([]uint8, levels)
	binMax := make([]uint8, levels)

	for b := 0; b < levels; b++ {
		start := b * n / levels
		end := (b + 1) * n / levels
		if b == levels-1 {
			end = n
		}
		if end <= start {
			end = start + 1
			if end > n {
				end = n
				start = n - 1
			}
		}

		sum := 0
		for i := start; i < end; i++ {
			sum += int(sortedVals[i])
		}
		reps[b] = uint8(sum / (end - start))
		binMax[b] = sortedVals[end-1]
	}

	b := 0
	for v := 0; v < 256; v++ {
		for b < levels-1 && uint8(v) > binMax[b] {
			b++
		}
		lut[v] = reps[b]
	}
	return lut
}
