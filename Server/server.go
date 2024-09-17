package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
)

type Cotacao struct {
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
}

func main() {

	http.HandleFunc("/cotacao", handlecotacao)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handlecotacao(w http.ResponseWriter, r *http.Request) {

	ctx, cancel := context.WithTimeout(r.Context(), 200*time.Millisecond)
	defer cancel()

	cotacao, err := getCotacao(ctx)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			log.Printf("Timeout: Timeout ao obter a cotacao")
		} else {
			log.Printf("Erro ao obter a cotacao: %v", err)
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	ctxDB, cancelDB := context.WithTimeout(context.Background(), 10*time.Millisecond)

	defer cancelDB()

	err = saveCotacao(ctxDB, cotacao)
	if err != nil {
		if ctxDB.Err() == context.DeadlineExceeded {
			log.Printf("Timeout: Timeout ao salvar a cotação no banco de dados")
		} else {
			log.Printf("Erro ao saalvar a cotação no banco de dados: %v", err)
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cotacao)

}

func getCotacao(ctx context.Context) (*Cotacao, error) {

	var result map[string]map[string]string

	req, err := http.NewRequestWithContext(ctx, "GET", "https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	cotacao := &Cotacao{

		Code:       result["USDBRL"]["code"],
		Codein:     result["USDBRL"]["Codein"],
		Name:       result["USDBRL"]["name"],
		High:       result["USDBRL"]["high"],
		Low:        result["USDBRL"]["low"],
		VarBid:     result["USDBRL"]["varbid"],
		PctChange:  result["USDBRL"]["pctchange"],
		Bid:        result["USDBRL"]["bid"],
		Ask:        result["USDBRL"]["ask"],
		Timestamp:  result["USDBRL"]["timestamp"],
		CreateDate: result["USDBRL"]["bid"],
	}

	log.Printf("Cotacao: %+v\n", cotacao)

	return cotacao, nil

}

func saveCotacao(ctx context.Context, cotacao *Cotacao) error {

	db, err := sql.Open("mysql", "root:root@tcp(localhost:3306)/goexperts")
	if err != nil {
		return err
	}
	defer db.Close()

	// Preparar a instrução para criar a tabela
	stmtCreate, err := db.PrepareContext(ctx, "CREATE TABLE IF NOT EXISTS cotacoes (id CHAR(36) PRIMARY KEY, bid TEXT, timestamp DATETIME DEFAULT CURRENT_TIMESTAMP)")
	if err != nil {
		return err
	}
	defer stmtCreate.Close()

	// Executar a instrução de criação da tabela
	_, err = stmtCreate.ExecContext(ctx)
	if err != nil {
		return err
	}

	// Preparar a instrução para inserir dados
	stmtInsert, err := db.PrepareContext(ctx, "INSERT INTO cotacoes (id, bid) VALUES (?, ?)")
	if err != nil {
		return err
	}
	defer stmtInsert.Close()

	// Executar a instrução de inserção
	_, err = stmtInsert.ExecContext(ctx, uuid.New().String(), cotacao.Bid)
	if err != nil {
		return err
	}

	return nil

}
