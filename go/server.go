package main

import (
	"bufio"
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"net"
	"strconv"
	"strings"
)

func main() {
	ln, err := net.Listen("tcp", ":8000")
	if err != nil {
		panic(err)
	}
	fmt.Println("Serveur TCP en écoute sur :8000")

	for {
		conn, err := ln.Accept()
		if err != nil {
			continue
		}
		go handleClient(conn)
	}
}

func handleClient(conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)

	// 1) Choix filtre
	conn.Write([]byte("Choisissez le filtre: blur / grayscale / invert\n"))
	filter, _ := reader.ReadString('\n')
	filter = strings.TrimSpace(strings.ToLower(filter))

	// 2) Radius si blur
	radius := 3
	if filter == "blur" {
		conn.Write([]byte("Radius ?\n"))
		line, _ := reader.ReadString('\n')
		if v, err := strconv.Atoi(strings.TrimSpace(line)); err == nil {
			radius = v
		}
	}

	conn.Write([]byte("OK\n"))

	// 3) Taille image
	sizeLine, _ := reader.ReadString('\n')
	sizeStr := strings.TrimPrefix(strings.TrimSpace(sizeLine), "SIZE=")
	size, _ := strconv.Atoi(sizeStr)

	// 4) Lecture image
	imgBytes := make([]byte, size)
	_, err := io.ReadFull(reader, imgBytes)
	if err != nil {
		return
	}

	img, _, err := image.Decode(bytes.NewReader(imgBytes))
	if err != nil {
		return
	}

	workers := 4
	var result *image.RGBA

	// 5) Application filtre
	if filter == "blur" {
		seqImg, parImg, seqDur, parDur := CompareBlur(img, workers, radius)
		_ = seqImg // utilisé seulement pour comparaison
		result = parImg

		conn.Write([]byte(fmt.Sprintf(
			"SEQ=%d;PAR=%d\n",
			seqDur.Nanoseconds(),
			parDur.Nanoseconds(),
		)))
	} else {
		result, _ = ApplyFilter(img, filter, workers, radius)
	}

	// 6) Envoi image
	var buf bytes.Buffer
	jpeg.Encode(&buf, result, &jpeg.Options{Quality: 90})

	conn.Write([]byte(fmt.Sprintf("SIZE=%d\n", buf.Len())))
	conn.Write(buf.Bytes())
}
