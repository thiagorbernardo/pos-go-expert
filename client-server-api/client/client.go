package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

type Response struct {
	Bid string `json:"bid"`
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	client := &http.Client{}

	req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:8080/cotacao", nil)
	if err != nil {
		log.Printf("Erro ao criar requisição: %v", err)
		return
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Erro ao fazer requisição: %v", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Erro ao ler resposta: %v", err)
		return
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("Erro: servidor retornou status %d - %s", resp.StatusCode, string(body))
		return
	}

	var response Response
	if err := json.Unmarshal(body, &response); err != nil {
		log.Printf("Erro ao decodificar resposta: %v\nResposta recebida: %s", err, string(body))
		return
	}

	file, err := os.OpenFile("cotacao.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("Erro ao abrir arquivo: %v", err)
		return
	}
	defer file.Close()

	now := time.Now().Format("2006-01-02 15:04:05")
	content := fmt.Sprintf("[%s] Dólar: %s\n", now, response.Bid)

	if _, err := file.WriteString(content); err != nil {
		log.Printf("Erro ao escrever no arquivo: %v", err)
		return
	}

	fmt.Printf("Cotação salva com sucesso: %s\n", response.Bid)
} 