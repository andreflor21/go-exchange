package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {

	ctx, cancel := context.WithTimeout(context.Background(),300*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", "http://127.0.0.1:8080/cotacao", nil)
	if err != nil {
		log.Fatalf("Erro ao criar a requisição: %v", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatalf("Erro ao fazer a requisição: %v", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Erro ao ler a resposta da requisição: %v", err)
	}

	var bid string
	err = json.Unmarshal(body, &bid)
	if err != nil {
		log.Fatalf("Erro ao fazer o parse do JSON: %v", err)
	}

	err = os.WriteFile("cotacao.txt", []byte("Dólar: " + bid), 0644)
	if err != nil {
		log.Fatalf("Erro ao escrever no arquivo: %v", err)
	}
	file, err := os.Open("cotacao.txt")
	if err != nil {
		log.Fatalf("Erro ao abrir o arquivo cotacao.txt: %v", err)
	}
	reader := bufio.NewReader(file)
	buffer := make([]byte, 50)
	for {
		n, err := reader.Read(buffer)
		if err != nil {
			break
		}
		fmt.Println(string(buffer[:n]))
	}
}