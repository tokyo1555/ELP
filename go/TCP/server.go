// server.go
package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"flag"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"net"
	"runtime"
	"strings"
	"time"
)

func main() {
	// =========================
	// 1) Flags / configuration
	// =========================
	addr := flag.String("addr", ":5000", "adresse d'écoute, ex: :5000 ou 0.0.0.0:5000")
	defaultWorkers := flag.Int("workers", 0, "workers par défaut si le client envoie 0 (0 => NumCPU)")
	flag.Parse()

	// =========================
	// 2) TCP listen + accept
	// =========================
	ln, err := net.Listen("tcp", *addr)
	if err != nil {
		panic(err)
	}
	defer ln.Close()

	fmt.Printf("Serveur TCP en écoute sur %s\n", *addr)
	printServerAddresses(*addr)

	for {
		conn, err := ln.Accept()
		if err != nil {
			continue
		}
		go handleConn(conn, *defaultWorkers)
	}
}

// =========================
// Affichage des IP locales
// =========================
func printServerAddresses(listenAddr string) {
	// Récupère le port depuis ":5000" / "0.0.0.0:5000" / "192.168.x.x:5000"
	port := listenAddr
	if !strings.HasPrefix(port, ":") {
		_, p, err := net.SplitHostPort(listenAddr)
		if err == nil {
			port = ":" + p
		} else {
			// fallback
			if idx := strings.LastIndex(listenAddr, ":"); idx != -1 {
				port = listenAddr[idx:]
			}
		}
	}

	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return
	}

	fmt.Println("Les clients peuvent se connecter via :")
	for _, addr := range addrs {
		ipnet, ok := addr.(*net.IPNet)
		if !ok || ipnet.IP.IsLoopback() {
			continue
		}
		ip := ipnet.IP.To4()
		if ip == nil {
			continue
		}
		fmt.Printf("  %s%s\n", ip.String(), port)
	}
}

// =========================
// Gestion d'une connexion
// =========================
func handleConn(conn net.Conn, defaultWorkers int) {
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(120 * time.Second))

	r := bufio.NewReader(conn)

	// 1) Lire requête
	filterName, radius, workers, imgBytes, err := readRequest(r)
	if err != nil {
		writeError(conn, fmt.Sprintf("lecture requête: %v", err))
		return
	}

	// 2) Décoder l'image (supporte jpg/png/gif/etc selon les imports)
	img, format, err := image.Decode(bytes.NewReader(imgBytes))
	if err != nil {
		writeError(conn, "échec décodage image (jpg/png/gif/etc)")
		return
	}

	// 3) Choisir workers
	if workers <= 0 {
		if defaultWorkers > 0 {
			workers = defaultWorkers
		} else {
			workers = runtime.NumCPU()
		}
	}

	// 4) Appliquer filtre (PARALLÈLE)
	out, err := ApplyFilter(img, filterName, workers, radius)
	if err != nil {
		writeError(conn, err.Error())
		return
	}

	// 5) Ré-encoder dans le MÊME format que l'entrée
	encoded, err := encodeSameFormat(out, format)
	if err != nil {
		writeError(conn, fmt.Sprintf("échec encodage (%s): %v", format, err))
		return
	}

	_ = writeOK(conn, encoded)
}

// =========================
// Protocole (request/response)
// =========================
func readRequest(r *bufio.Reader) (name string, radius int, workers int, img []byte, err error) {
	var nameLen uint32
	if err = binary.Read(r, binary.BigEndian, &nameLen); err != nil {
		return
	}
	if nameLen == 0 || nameLen > 64 {
		err = fmt.Errorf("longueur nom filtre invalide: %d", nameLen)
		return
	}
	nameBytes := make([]byte, nameLen)
	if _, err = io.ReadFull(r, nameBytes); err != nil {
		return
	}
	name = string(nameBytes)

	var rad32 int32
	if err = binary.Read(r, binary.BigEndian, &rad32); err != nil {
		return
	}
	radius = int(rad32)

	var w32 int32
	if err = binary.Read(r, binary.BigEndian, &w32); err != nil {
		return
	}
	workers = int(w32)

	var imgSize uint64
	if err = binary.Read(r, binary.BigEndian, &imgSize); err != nil {
		return
	}
	if imgSize == 0 || imgSize > 200*1024*1024 {
		err = fmt.Errorf("image vide ou trop grande: %d octets", imgSize)
		return
	}
	img = make([]byte, imgSize)
	_, err = io.ReadFull(r, img)
	return
}

func writeError(w io.Writer, msg string) {
	_ = binary.Write(w, binary.BigEndian, uint32(1))
	_ = binary.Write(w, binary.BigEndian, uint32(len(msg)))
	_, _ = w.Write([]byte(msg))
}

func writeOK(w io.Writer, img []byte) error {
	if err := binary.Write(w, binary.BigEndian, uint32(0)); err != nil {
		return err
	}
	if err := binary.Write(w, binary.BigEndian, uint64(len(img))); err != nil {
		return err
	}
	_, err := w.Write(img)
	return err
}

// =========================
// Encodage au même format
// =========================
func encodeSameFormat(img image.Image, format string) ([]byte, error) {
	var buf bytes.Buffer

	switch strings.ToLower(format) {
	case "jpeg", "jpg":
		err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 95})
		return buf.Bytes(), err

	case "png":
		err := png.Encode(&buf, img)
		return buf.Bytes(), err

	case "gif":
		// GIF: palette + dithering gérés par l'encodeur standard.
		err := gif.Encode(&buf, img, nil)
		return buf.Bytes(), err

	default:
		// Fallback: si format inconnu, on choisit PNG (sans perte).
		err := png.Encode(&buf, img)
		return buf.Bytes(), err
	}
}

// =========================
// Routeur de filtres (serveur)
// =========================
// name : "grayscale|invert|blur|gaussian|sobel|median|pixelate|oilpaint"
// workers : nombre de goroutines
// radius : intensité / paramètre selon filtre
func ApplyFilter(img image.Image, name string, workers int, radius int) (*image.RGBA, error) {
	switch name {
	case "grayscale":
		return Grayscale(img, workers), nil

	case "invert":
		return Invert(img, workers), nil

	case "blur":
		if radius < 1 {
			radius = 1
		}
		return Blur(img, workers, radius), nil

	case "gaussian":
		return GaussianBlur(img, workers), nil

	case "sobel":
		return Sobel(img, workers), nil

	case "median":
		return MedianFilter(img, workers), nil

	case "pixelate":
		if radius < 2 {
			radius = 2
		}
		return Pixelate(img, workers, radius), nil

	case "oilpaint":
		if radius < 3 {
			radius = 5 // brushSize par défaut
		}
		return OilPaint(img, workers, radius), nil

	default:
		return nil, fmt.Errorf("filtre inconnu. Utilise: grayscale|invert|blur|gaussian|sobel|median|pixelate|oilpaint")
	}
}

