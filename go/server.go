package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"flag"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"net"
	"runtime"
	"strings"
	"time"
)

func main() {
	addr := flag.String("addr", ":5000", "address to listen on, e.g. :5000 or 0.0.0.0:5000")
	defaultWorkers := flag.Int("workers", 0, "default workers if client sends 0 (0 => NumCPU)")
	flag.Parse()

	ln, err := net.Listen("tcp", *addr)
	if err != nil {
		panic(err)
	}
	defer ln.Close()

	fmt.Printf("Server listening on %s\n", *addr)
	printServerAddresses(*addr)

	for {
		conn, err := ln.Accept()
		if err != nil {
			continue
		}
		go handleConn(conn, *defaultWorkers)
	}
}

// Affiche les IP locales utilisables par un client
func printServerAddresses(listenAddr string) {
	// Récupère le port depuis ":5000" / "0.0.0.0:5000" / "192.168.x.x:5000"
	port := listenAddr
	if !strings.HasPrefix(port, ":") {
		_, p, err := net.SplitHostPort(listenAddr)
		if err == nil {
			port = ":" + p
		} else {
			// fallback (si format bizarre)
			if idx := strings.LastIndex(listenAddr, ":"); idx != -1 {
				port = listenAddr[idx:]
			}
		}
	}

	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return
	}

	fmt.Println("Clients can connect using:")
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

func handleConn(conn net.Conn, defaultWorkers int) {
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(120 * time.Second))

	r := bufio.NewReader(conn)

	filterName, radius, workers, imgBytes, err := readRequest(r)
	if err != nil {
		writeError(conn, fmt.Sprintf("read request: %v", err))
		return
	}

	img, _, err := image.Decode(bytes.NewReader(imgBytes))
	if err != nil {
		writeError(conn, "decode image failed (expected jpg/png/etc)")
		return
	}

	if workers <= 0 {
		if defaultWorkers > 0 {
			workers = defaultWorkers
		} else {
			workers = runtime.NumCPU()
		}
	}

	// Appliquer filtre (PARALLÈLE)
	var out *image.RGBA
	switch filterName {
	case "grayscale":
		out = Grayscale(img, workers)
	case "invert":
		out = Invert(img, workers)
	case "blur":
		if radius < 1 {
			radius = 1
		}
		out = Blur(img, workers, radius)
	case "gaussian":
		out = GaussianBlur(img, workers)
	case "sobel":
		out = Sobel(img, workers)
	case "median":
		out = MedianFilter(img, workers)
	case "bilateral":
		out = BilateralFilter(img, workers, 1.5, 25.0) // valeurs par défaut
	case "oilpaint":
		if radius < 3 {
			radius = 5 // brushSize par défaut
		}
		out = OilPaint(img, workers, radius)
	default:
		writeError(conn, "unknown filter. use: grayscale|invert|blur|gaussian|sobel|median|bilateral|oilpaint")
		return
	}

	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, out, &jpeg.Options{Quality: 95}); err != nil {
		writeError(conn, "encode jpeg failed")
		return
	}

	_ = writeOK(conn, buf.Bytes())
}

func readRequest(r *bufio.Reader) (name string, radius int, workers int, img []byte, err error) {
	var nameLen uint32
	if err = binary.Read(r, binary.BigEndian, &nameLen); err != nil {
		return
	}
	if nameLen == 0 || nameLen > 64 {
		err = fmt.Errorf("invalid filter name length %d", nameLen)
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
		err = fmt.Errorf("image too large or empty: %d bytes", imgSize)
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
