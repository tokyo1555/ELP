// client.go
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

type filterInfo struct {
	Name string
	Desc string
}

var filters = []filterInfo{
	{"grayscale", "Convertit l'image en niveaux de gris."},
	{"invert", "Inverse les couleurs (effet négatif)."},
	{"blur", "Flou simple (box blur). Plus le rayon est grand, plus c'est flou."},
	{"gaussian", "Flou gaussien 5x5 (plus doux que blur)."},
	{"sobel", "Détection de contours (edges) en noir et blanc."},
	{"median", "Filtre médian 3x3 (réduit le bruit type 'sel et poivre')."},
	{"oilpaint", "Effet peinture à l'huile (couleur dominante dans un pinceau)."},
	{"pixelate", "Effet mosaïque (gros pixels)."},
}

func main() {

	//choix du fichier d'entrée
	candidats := []string{"input.jpg", "input.jpeg", "input.png"}
	inPath := ""

	for _, p := range candidats {
		if _, err := os.Stat(p); err == nil {
			inPath = p
			break
		}
	}
	if inPath == "" {
		panic("Aucun fichier trouvé: input.jpg / input.jpeg / input.png")
	}
	if len(os.Args) >= 2 {
		inPath = os.Args[1]
	}

	imgBytes, err := os.ReadFile(inPath)
	if err != nil {
		panic("Impossible de lire l'image. Mets un fichier (ex: input.png) ou lance: go run client.go monimage.png")
	}

	reader := bufio.NewReader(os.Stdin)

	//paramètres client
	serverAddr := askServer(reader)
	filterName := askFilter(reader)

	radius := 0
	switch filterName {
	case "blur":
		radius = askInt(reader, "Choisis l'intensité du flou (radius >= 1) : ", 1, 999)
	case "oilpaint":
		radius = askInt(reader, "Choisis la taille du pinceau (brushSize >= 3) : ", 3, 999)
	case "pixelate":
		radius = askInt(reader, "Choisis la taille des blocs mosaïque (block >= 2) : ", 2, 999)
	}

	workers := askWorkers(reader)

	//connexion + requête
	conn, err := net.DialTimeout("tcp", serverAddr, 10*time.Second)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	if err := sendRequest(conn, filterName, radius, workers, imgBytes); err != nil {
		panic(err)
	}

	//réponse + sauvegarde
	respImg, err := readResponse(conn)
	if err != nil {
		panic(err)
	}

	ext := filepath.Ext(inPath) // on garde la meme extension que l'entrée
	if ext == "" {
		ext = ".png" // fallback si le fichier n'a pas d'extension
	}

	base := strings.TrimSuffix(filepath.Base(inPath), ext)
	outName := fmt.Sprintf("%s_output_%s%s", base, filterName, ext)

	if err := os.WriteFile(outName, respImg, 0644); err != nil {
		panic(err)
	}

	fmt.Printf("\nImage reçue et sauvegardée : %s\n", outName)
}

// Saisie utilisateur
func askServer(r *bufio.Reader) string {
	for {
		fmt.Println("Conseil : place ton image dans ce dossier (en le nommant input) puis lance client.go avec le nom du fichier.")
		fmt.Print("Entre l'adresse du serveur (IP:PORT) [ex: 192.168.1.10:5000] : ")
		s, _ := r.ReadString('\n')
		s = strings.TrimSpace(s)

		if s == "" {
			continue
		}
		if strings.Count(s, ":") < 1 {
			fmt.Println(" Format invalide. Exemple : 192.168.1.10:5000")
			continue
		}
		return s
	}
}

func askFilter(r *bufio.Reader) string {
	fmt.Println("\nChoisis un filtre :")
	for i, f := range filters {
		fmt.Printf("  %d) %-9s  %s\n", i+1, f.Name, f.Desc)
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
		return filters[n-1].Name
	}
}

func askWorkers(r *bufio.Reader) int {
	fmt.Println("\nWorkers (nombre de goroutines côté serveur) :")
	fmt.Println("  0) Laisser le serveur choisir (recommandé)")
	fmt.Println("  2, 4, 8, ...) Forcer une valeur")
	return askInt(r, "Ton choix [0..64] : ", 0, 64)
}

func askInt(r *bufio.Reader, prompt string, min int, max int) int {
	for {
		fmt.Print(prompt)
		s, _ := r.ReadString('\n')
		s = strings.TrimSpace(s)

		n, err := strconv.Atoi(s)
		if err != nil || n < min || n > max {
			fmt.Printf("Valeur invalide. Entre %d et %d.\n", min, max)
			continue
		}
		return n
	}
}

// Protocole binaire (client)
func sendRequest(w io.Writer, filterName string, radius int, workers int, img []byte) error {
	nameBytes := []byte(filterName)

	// [u32 nameLen][name][i32 radius][i32 workers][u64 imgSize][imgBytes]
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

	// [u32 status] status=0 OK, status=1 erreur
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
		return nil, fmt.Errorf("erreur serveur: %s", string(msg))
	}

	var size uint64
	if err := binary.Read(r, binary.BigEndian, &size); err != nil {
		return nil, err
	}
	if size == 0 || size > 200*1024*1024 {
		return nil, fmt.Errorf("taille de réponse invalide: %d", size)
	}

	buf := make([]byte, size)
	_, err := io.ReadFull(r, buf)
	return buf, err
}
