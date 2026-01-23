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
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run client.go <server_ip:port>")
		fmt.Println("Exemple: go run client.go 192.168.1.10:8000")
		return
	}

	serverAddr := os.Args[1]
	fmt.Println("Connexion au serveur", serverAddr)

	conn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		fmt.Println("Erreur connexion:", err)
		return
	}
	defer conn.Close()

	reader := bufio.NewReader(conn)
	stdin := bufio.NewReader(os.Stdin)

	// 1) Choix du filtre
	fmt.Print("Serveur: ")
	serverLine := must(reader.ReadString('\n'))
	fmt.Print(serverLine)

	fmt.Print("Filtre: ")
	filter := must(stdin.ReadString('\n'))
	filter = strings.TrimSpace(strings.ToLower(filter))
	conn.Write([]byte(filter + "\n")) // <-- Important, ajouter le \n

	// 2) Radius si blur
	if filter == "blur" {
		fmt.Print("Serveur: ")
		serverLine = must(reader.ReadString('\n'))
		fmt.Print(serverLine)

		fmt.Print("Radius: ")
		radius := must(stdin.ReadString('\n'))
		radius = strings.TrimSpace(radius)
		conn.Write([]byte(radius + "\n")) // <-- Important, ajouter le \n
	}

	// 3) OK serveur
	_ = must(reader.ReadString('\n')) // "OK"

	// 4) Envoi image
	imgBytes, err := os.ReadFile("input.jpg")
	if err != nil {
		fmt.Println("Erreur lecture input.jpg:", err)
		return
	}

	conn.Write([]byte(fmt.Sprintf("SIZE=%d\n", len(imgBytes))))
	conn.Write(imgBytes)
	fmt.Println("📤 Image envoyée")

	// 5) Durées si blur
	if filter == "blur" {
		durations := must(reader.ReadString('\n'))
		fmt.Println("Durées:", durations)
	}

	// 6) Lecture SIZE image résultat
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

	// 7) Lecture image binaire
	outBytes := make([]byte, outSize)
	_, err = io.ReadFull(reader, outBytes)
	if err != nil {
		fmt.Println("Erreur lecture image:", err)
		return
	}

	// 8) Sauvegarde image
	err = os.WriteFile("output.jpg", outBytes, 0644)
	if err != nil {
		fmt.Println("Erreur écriture output.jpg:", err)
		return
	}

	fmt.Println("✅ Image filtrée reçue : output.jpg")
}

// ============================
// Fonction utilitaire
// ============================
func must(s string, err error) string {
	if err != nil {
		panic(err)
	}
	return s
}
