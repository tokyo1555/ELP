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
	conn, _ := net.Dial("tcp", "127.0.0.1:8000")
	defer conn.Close()

	r := bufio.NewReader(conn)
	in := bufio.NewReader(os.Stdin)

	fmt.Print("Serveur:", must(r.ReadString('\n')))
	fmt.Print("Filtre: ")
	filter := must(in.ReadString('\n'))
	conn.Write([]byte(filter))
	filter = strings.TrimSpace(strings.ToLower(filter))

	if filter == "blur" {
		fmt.Print("Serveur:", must(r.ReadString('\n')))
		fmt.Print("Radius: ")
		conn.Write([]byte(must(in.ReadString('\n'))))
	}

	must(r.ReadString('\n')) // OK

	img, _ := os.ReadFile("input.jpg")
	conn.Write([]byte(fmt.Sprintf("SIZE=%d\n", len(img))))
	conn.Write(img)

	if filter == "blur" {
		fmt.Println("Durées:", must(r.ReadString('\n')))
	}

	sizeLineBytes, err := r.ReadBytes('\n')
	if err != nil {
		panic(err)
	}
	sizeLine := string(sizeLineBytes)

	if !strings.HasPrefix(sizeLine, "SIZE=") {
		panic("protocole cassé: SIZE attendu")
	}

	size, _ := strconv.Atoi(strings.TrimPrefix(strings.TrimSpace(sizeLine), "SIZE="))

	out := make([]byte, size)
	io.ReadFull(conn, out)

	os.WriteFile("output.jpg", out, 0644)
	fmt.Println("✅ Image reçue : output.jpg")
}

func must(s string, err error) string {
	if err != nil {
		panic(err)
	}
	return s
}
