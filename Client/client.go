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
	ctx, cancel := context.WithTimeout(context.Background(), 400*time.Millisecond)
	defer cancel()

	cotacao, err := getCotacao(ctx)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			log.Print("Timeout: Timeout ao obter a cotação do servidor")
		} else {
			log.Fatalf("Erro: %v", err.Error())
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

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusRequestTimeout {
		return nil, fmt.Errorf("Timeout ao obter a cotação")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Erro ao obter a cotação: %s", resp.Status)
	}

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
