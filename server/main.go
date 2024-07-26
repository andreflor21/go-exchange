package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3"
)


type Exchange struct {
	Usdbrl struct {
		Code       string `json:"code"`
		Codein     string `json:"codein"`
		Name       string `json:"name"`
		High       string `json:"high"`
		Low        string `json:"low"`
		VarBid     string `json:"varBid"`
		PctChange  string `json:"pctChange"`
		Bid        string `json:"bid"`
		Ask        string `json:"ask"`
		Timestamp  string `json:"timestamp"`
		CreateDate string `json:"create_date"`
	} `json:"USDBRL"`
}

func main(){
	http.HandleFunc("/cotacao", ExchangeHandler);
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func ExchangeHandler(w http.ResponseWriter, r *http.Request){
	log.Printf("Recebendo requisição\n")
	rate, err := GetExchangeRate()
	if err != nil {
		log.Printf("Erro ao buscar cotação: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Printf("Salvando cotação...\n")
	err = saveToDatabase(rate)
	if err != nil {
		log.Printf("Erro ao salvar cotação no banco de dados: %v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Printf("Cotação salva\n")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rate.Usdbrl.Bid)
	log.Printf("Requisição finalizada\n")
}	

func GetExchangeRate() (*Exchange, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "GET", "https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)
	if err != nil {
		log.Printf("Erro ao criar a requisição na API: %v\n", err)
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("Erro ao fazer a requisição na API: %v\n", err)
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Erro ao ler o corpo da resposta: %v\n", err)
		return nil, err
	}
	var rate Exchange
	err = json.Unmarshal(body, &rate)
	if err != nil {
		log.Printf("Erro ao fazer o decode da resposta: %v\n", err)
		return nil, err
	}

	return &rate, nil

}

func saveToDatabase(rate *Exchange) error {
	db, err := sql.Open("sqlite3", "./exchange.db")
	if err != nil {
		log.Printf("Erro ao abrir o banco de dados\n")
		return err
	}
	defer db.Close()

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS exchange_rates (id INTEGER PRIMARY KEY AUTOINCREMENT, bid DECIMAL(10,4), date DATETIME)")
	if err != nil {
		log.Printf("Erro ao criar a tabela no banco de dados\n")
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	_,err = db.ExecContext(ctx, `INSERT INTO exchange_rates (bid, date) VALUES (?, ?)`, rate.Usdbrl.Bid, rate.Usdbrl.CreateDate)
	if err != nil {
		log.Printf("Erro ao inserir cotação no banco de dados\n")
		return err
	}
	return nil
}