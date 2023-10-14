package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type APICotacaoResponse struct {
	Valor float64 `json:"valor"`
}

func main() {
	cotacao, err := getUSDtoBRL()
	if err != nil {
		panic(err)
	}

	fmt.Printf("Cotação do dólar para real: %v\n", cotacao.Valor)
	populateFile(cotacao.Valor)
}

func getUSDtoBRL() (*APICotacaoResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*300)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost:8080/cotacao", nil)
	if err != nil {
		return nil, fmt.Errorf("Erro ao criar requisição: %v\n", err)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Erro ao fazer requisição: %v\n", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("Erro ao ler resposta: %v\n", err)
	}

	var data APICotacaoResponse
	err = json.Unmarshal(body, &data)
	if err != nil {
		return nil, fmt.Errorf("Erro ao converter resposta: %v\n", err)
	}

	return &data, nil
}

func populateFile(cotacao float64) {
	f, err := os.Create("cotacao.txt")
	if err != nil {
		panic(err)
	}

	_, err = f.Write([]byte(fmt.Sprintf("Dólar: %f", cotacao)))
	if err != nil {
		panic(err)
	}

	err = f.Close()
	if err != nil {
		return
	}
}
