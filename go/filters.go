package main

import (
	"fmt"
	"image"
	"image/color"
	"sync"
	"time"
)

// GrayscaleSeq applique un filtre de niveaux de gris de manière séquentielle.
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

// Grayscale applique un filtre de niveaux de gris en parallèle.
func Grayscale(img image.Image, workers int) *image.RGBA {
	bounds := img.Bounds()
	result := image.NewRGBA(bounds)

	height := bounds.Max.Y - bounds.Min.Y
	if workers > height {
		workers = height
	}
	if workers < 1 {
		workers = 1
	}
	block := height / workers

	var wg sync.WaitGroup

	for i := 0; i < workers; i++ {
		startY := bounds.Min.Y + i*block
		endY := startY + block
		if i == workers-1 {
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

// Invert applique un négatif de l'image en parallèle.
func Invert(img image.Image, workers int) *image.RGBA {
	bounds := img.Bounds()
	result := image.NewRGBA(bounds)

	height := bounds.Max.Y - bounds.Min.Y
	if workers > height {
		workers = height
	}
	if workers < 1 {
		workers = 1
	}
	block := height / workers

	var wg sync.WaitGroup

	for i := 0; i < workers; i++ {
		startY := bounds.Min.Y + i*block
		endY := startY + block
		if i == workers-1 {
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

// Blur applique un flou "box blur" de rayon donné (radius >= 1) en parallèle.
// radius = 1 -> ~3x3, radius = 5 -> ~11x11, etc.
func Blur(img image.Image, workers int, radius int) *image.RGBA {
	bounds := img.Bounds()
	result := image.NewRGBA(bounds)

	if radius < 1 {
		radius = 1
	}

	height := bounds.Max.Y - bounds.Min.Y
	if workers > height {
		workers = height
	}
	if workers < 1 {
		workers = 1
	}
	block := height / workers

	var wg sync.WaitGroup

	for i := 0; i < workers; i++ {
		startY := bounds.Min.Y + i*block
		endY := startY + block
		if i == workers-1 {
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
						// sécurité : on recopie le pixel original
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

// CompareBlur applique le blur en séquentiel et en concurrent
// et renvoie les deux images + durées.
func CompareBlur(img image.Image, workers int, radius int) (seqImg, parImg *image.RGBA, seqDur, parDur time.Duration) {
	// Séquentiel
	startSeq := time.Now()
	seqImg = BlurSeq(img, radius)
	seqDur = time.Since(startSeq)

	// Concurrent
	startPar := time.Now()
	parImg = Blur(img, workers, radius)
	parDur = time.Since(startPar)

	return
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
