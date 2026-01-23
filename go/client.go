package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
)

func main() {
	// ===============================
	// 0) Vérification des arguments
	// ===============================
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run client.go <server_ip:port>")
		fmt.Println("Exemple: go run client.go 127.0.0.1:8000")
		return
	}

	serverAddr := os.Args[1]

	// ===============================
	// 1) Connexion au serveur
	// ===============================
	conn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		fmt.Println("Erreur connexion:", err)
		return
	}
	defer conn.Close()

	reader := bufio.NewReader(conn)
	stdin := bufio.NewReader(os.Stdin)

	// ===============================
	// 2) Choix du filtre
	// ===============================
	fmt.Print("Serveur: ")
	fmt.Print(must(reader.ReadString('\n')))

	fmt.Print("Filtre: ")
	filter := must(stdin.ReadString('\n'))
	conn.Write([]byte(filter))
	filter = strings.TrimSpace(strings.ToLower(filter))

	// ===============================
	// 3) Radius si blur
	// ===============================
	if filter == "blur" {
		fmt.Print("Serveur: ")
		fmt.Print(must(reader.ReadString('\n')))

		fmt.Print("Radius: ")
		radius := must(stdin.ReadString('\n'))
		conn.Write([]byte(radius))
	}

	// ===============================
	// 4) OK serveur
	// ===============================
	must(reader.ReadString('\n')) // "OK"

	// ===============================
	// 5) Envoi image
	// ===============================
	imgBytes, err := os.ReadFile("input.jpg")
	if err != nil {
		fmt.Println("Erreur lecture input.jpg:", err)
		return
	}

	conn.Write([]byte(fmt.Sprintf("SIZE=%d\n", len(imgBytes))))
	conn.Write(imgBytes)
	fmt.Println("📤 Image envoyée")

	// ===============================
	// 6) Durées si blur
	// ===============================
	if filter == "blur" {
		fmt.Println("Durées:", must(reader.ReadString('\n')))
	}

	// ===============================
	// 7) Lecture SIZE image résultat
	// ===============================
	sizeLineBytes, err := reader.ReadBytes('\n')
	if err != nil {
		fmt.Println("Erreur lecture SIZE:", err)
		return
	}
	sizeLine := string(sizeLineBytes)

	if !strings.HasPrefix(sizeLine, "SIZE=") {
		fmt.Println("Protocole invalide, SIZE attendu")
		return
	}

	sizeStr := strings.TrimPrefix(strings.TrimSpace(sizeLine), "SIZE=")
	outSize, err := strconv.Atoi(sizeStr)
	if err != nil || outSize <= 0 {
		fmt.Println("Taille invalide:", sizeStr)
		return
	}

	// ===============================
	// 8) Lecture image binaire
	// ===============================
	outBytes := make([]byte, outSize)
	_, err = io.ReadFull(reader, outBytes)
	if err != nil {
		fmt.Println("Erreur lecture image:", err)
		return
	}

	// ===============================
	// 9) Sauvegarde image
	// ===============================
	err = os.WriteFile("output.jpg", outBytes, 0644)
	if err != nil {
		fmt.Println("Erreur écriture output.jpg:", err)
		return
	}

	fmt.Println("✅ Image filtrée reçue : output.jpg")
}

// ===================================================
// Fonction utilitaire : panic si erreur
// ===================================================
func must(s string, err error) string {
	if err != nil {
		panic(err)
	}
	return s
}
