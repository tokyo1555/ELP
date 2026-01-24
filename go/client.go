package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

var filters = []string{
	"grayscale",
	"invert",
	"blur",
	"gaussian",
	"sobel",
	"median",
	"bilateral",
	"oilpaint",
}

func main() {
	in := "input.jpg"

	imgBytes, err := os.ReadFile(in)
	if err != nil {
		panic("cannot read input.jpg (place it next to client.go)")
	}

	reader := bufio.NewReader(os.Stdin)

	serverAddr := askServer(reader)
	filterName := askFilter(reader)

	radius := 0
	if filterName == "blur" {
		radius = askInt(reader, "Choisis l'intensité du blur (radius >= 1): ", 1, 1000)
	}
	if filterName == "oilpaint" {
		radius = askInt(reader, "Choisis la taille du pinceau OilPaint (brushSize >= 3): ", 3, 999)
	}

	workers := askWorkers(reader)

	conn, err := net.DialTimeout("tcp", serverAddr, 10*time.Second)
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(120 * time.Second))

	if err := sendRequest(conn, filterName, radius, workers, imgBytes); err != nil {
		panic(err)
	}

	respImg, err := readResponse(conn)
	if err != nil {
		panic(err)
	}

	base := strings.TrimSuffix(filepath.Base(in), filepath.Ext(in))
	outName := fmt.Sprintf("%s_output_%s.jpg", base, filterName)
	if err := os.WriteFile(outName, respImg, 0644); err != nil {
		panic(err)
	}

	fmt.Printf("\n✅ Image reçue et sauvegardée: %s\n", outName)
}

func askServer(r *bufio.Reader) string {
	for {
		fmt.Print("Adresse du serveur (IP:PORT) [ex: 192.168.1.10:5000] : ")
		s, _ := r.ReadString('\n')
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		if strings.Count(s, ":") < 1 {
			fmt.Println("❌ Format invalide. Exemple: 192.168.1.10:5000")
			continue
		}
		return s
	}
}

func askFilter(r *bufio.Reader) string {
	fmt.Println("\nChoisis un filtre :")
	for i, f := range filters {
		fmt.Printf("  %d) %s\n", i+1, f)
	}
	for {
		fmt.Print("Ton choix (numéro) : ")
		s, _ := r.ReadString('\n')
		s = strings.TrimSpace(s)
		n, err := strconv.Atoi(s)
		if err != nil || n < 1 || n > len(filters) {
			fmt.Println("❌ Choix invalide. Donne un numéro de la liste.")
			continue
		}
		return filters[n-1]
	}
}

func askWorkers(r *bufio.Reader) int {
	fmt.Println("\nWorkers (nombre de goroutines côté serveur) :")
	fmt.Println("  0) Laisser le serveur choisir (recommandé)")
	fmt.Println("  2,4,8,...) Forcer une valeur")
	return askInt(r, "Ton choix [0..128] : ", 0, 128)
}

func askInt(r *bufio.Reader, prompt string, min int, max int) int {
	for {
		fmt.Print(prompt)
		s, _ := r.ReadString('\n')
		s = strings.TrimSpace(s)
		n, err := strconv.Atoi(s)
		if err != nil || n < min || n > max {
			fmt.Printf("❌ Valeur invalide. Entre %d et %d.\n", min, max)
			continue
		}
		return n
	}
}

func sendRequest(w io.Writer, filterName string, radius int, workers int, img []byte) error {
	nameBytes := []byte(filterName)

	if err := binary.Write(w, binary.BigEndian, uint32(len(nameBytes))); err != nil {
		return err
	}
	if _, err := w.Write(nameBytes); err != nil {
		return err
	}

	if err := binary.Write(w, binary.BigEndian, int32(radius)); err != nil {
		return err
	}
	if err := binary.Write(w, binary.BigEndian, int32(workers)); err != nil {
		return err
	}
	if err := binary.Write(w, binary.BigEndian, uint64(len(img))); err != nil {
		return err
	}
	_, err := w.Write(img)
	return err
}

func readResponse(conn net.Conn) ([]byte, error) {
	r := bufio.NewReader(conn)

	var status uint32
	if err := binary.Read(r, binary.BigEndian, &status); err != nil {
		return nil, err
	}

	if status != 0 {
		var msgLen uint32
		if err := binary.Read(r, binary.BigEndian, &msgLen); err != nil {
			return nil, err
		}
		msg := make([]byte, msgLen)
		if _, err := io.ReadFull(r, msg); err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("server error: %s", string(msg))
	}

	var size uint64
	if err := binary.Read(r, binary.BigEndian, &size); err != nil {
		return nil, err
	}
	if size == 0 || size > 200*1024*1024 {
		return nil, fmt.Errorf("invalid response size: %d", size)
	}

	buf := make([]byte, size)
	_, err := io.ReadFull(r, buf)
	return buf, err
}
