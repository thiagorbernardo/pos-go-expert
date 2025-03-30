package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Cotacao struct {
	USDBRL struct {
		Bid string `json:"bid"`
	} `json:"USDBRL"`
}

func main() {
	db, err := sql.Open("sqlite3", "./data/cotacoes.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS cotacoes (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			valor TEXT NOT NULL,
			data DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/cotacao", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
		defer cancel()

		client := &http.Client{}

		req, err := http.NewRequestWithContext(ctx, "GET", "https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)
		if err != nil {
			log.Printf("Erro ao criar requisição: %v", err)
			http.Error(w, "Erro interno do servidor", http.StatusInternalServerError)
			return
		}

		resp, err := client.Do(req)
		if err != nil {
			log.Printf("Erro ao fazer requisição: %v", err)
			http.Error(w, "Erro ao buscar cotação", http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		var cotacao Cotacao
		if err := json.NewDecoder(resp.Body).Decode(&cotacao); err != nil {
			log.Printf("Erro ao decodificar resposta: %v", err)
			http.Error(w, "Erro ao processar resposta", http.StatusInternalServerError)
			return
		}

		ctxDB, cancelDB := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancelDB()

		_, err = db.ExecContext(ctxDB, "INSERT INTO cotacoes (valor) VALUES (?)", cotacao.USDBRL.Bid)
		if err != nil {
			log.Printf("Erro ao salvar no banco: %v", err)
			http.Error(w, "Erro ao salvar cotação", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"bid": cotacao.USDBRL.Bid})
	})

	fmt.Println("Servidor rodando na porta 8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
} 