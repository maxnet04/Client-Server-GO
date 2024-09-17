package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

type Cotacao struct {
	Bid string `json:"bid"`
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	cotacao, err := getCotacao(ctx)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			log.Print("Timeout: Timeout ao obter a cotação do servidor")
		} else {
			log.Fatalf("Erro ao obter a cotação: %v", err)
		}
	}

	err = saveCotacaoToFile(cotacao)
	if err != nil {
		log.Fatalf("Erro ao salvar a cotação no arquivo: %v", err)
	}
}

func getCotacao(ctx context.Context) (*Cotacao, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:8080/cotacao", nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var cotacao Cotacao
	if err := json.NewDecoder(resp.Body).Decode(&cotacao); err != nil {
		return nil, err
	}

	return &cotacao, nil
}

func saveCotacaoToFile(cotacao *Cotacao) error {
	content := fmt.Sprintf("Dólar: %s\n", cotacao.Bid)
	log.Print(content)
	return os.WriteFile("cotacao.txt", []byte(content), 0644)

}
