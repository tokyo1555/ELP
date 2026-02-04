package main

import (
	"image"
	"image/color"
	"math"
	"sort"
)

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

func PixelateSeq(img image.Image, blockSize int) *image.RGBA {
	bounds := img.Bounds()
	out := image.NewRGBA(bounds)

	if blockSize < 2 {
		blockSize = 2
	}

	for y := bounds.Min.Y; y < bounds.Max.Y; y += blockSize {
		for x := bounds.Min.X; x < bounds.Max.X; x += blockSize {
			var sumR, sumG, sumB uint32
			var count uint32

			// Moyenne bloc
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

			// Remplissage bloc
			for yy := y; yy < y+blockSize && yy < bounds.Max.Y; yy++ {
				for xx := x; xx < x+blockSize && xx < bounds.Max.X; xx++ {
					out.Set(xx, yy, avg)
				}
			}
		}
	}
	return out
}

// PosterizeQuantilesColorSeq : version séquentielle de la posterization couleur
// basée sur des quantiles globaux appliqués séparément à R, G et B.
// levels = nombre de niveaux par canal (>=2)
// Complexité : O(N log N) (tri des valeurs R, G, B)
func PosterizeQuantilesColorSeq(img image.Image, levels int) *image.RGBA {
	if levels < 2 {
		levels = 2
	}

	bounds := img.Bounds()
	wImg := bounds.Dx()
	hImg := bounds.Dy()
	n := wImg * hImg

	// Conversion en RGBA
	src := image.NewRGBA(bounds)
	draw.Draw(src, bounds, img, bounds.Min, draw.Src)

	// 1) Extraire R, G, B
	rs := make([]uint8, n)
	gs := make([]uint8, n)
	bs := make([]uint8, n)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
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

	// 2) Trier chaque canal
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

	// 3) Appliquer les LUT
	out := image.NewRGBA(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
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

	return out
}
