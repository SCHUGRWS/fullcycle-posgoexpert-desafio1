package main

import (
	"context"
	"encoding/json"
	"fmt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"io"
	"log"
	"net/http"
	"time"
)

var DB *gorm.DB

type Cotacao struct {
	Valor float64 `json:"valor"`
	gorm.Model
}

type EconomiaAPIUsdBrlResponse struct {
	Usdbrl struct {
		Code       string  `json:"code"`
		Codein     string  `json:"codein"`
		Name       string  `json:"name"`
		High       string  `json:"high"`
		Low        string  `json:"low"`
		VarBid     string  `json:"varBid"`
		PctChange  string  `json:"pctChange"`
		Bid        float64 `json:"bid,string"`
		Ask        string  `json:"ask"`
		Timestamp  string  `json:"timestamp"`
		CreateDate string  `json:"create_date"`
	} `json:"USDBRL"`
}

func main() {
	err := InitializeDatabase()
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/cotacao", getCotacaoUsdBrl)
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		return
	}
}

func InitializeDatabase() error {
	db, err := gorm.Open(sqlite.Open("desafio.db?cache=shared"), &gorm.Config{})
	if err != nil {
		return err
	}

	err = db.AutoMigrate(&Cotacao{})
	if err != nil {
		return err
	}

	DB = db

	return nil
}

func getCotacaoUsdBrl(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if r.URL.Path != "/cotacao" {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	select {
	default:
		data, err := getUSDtoBRL()
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
		}

		cotacao, err := createCotacao(data)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		err = json.NewEncoder(w).Encode(cotacao)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			_, err := w.Write([]byte("Erro ao converter resposta"))
			if err != nil {
				return
			}
		}
	case <-ctx.Done():
		log.Println("Request cancelada pelo cliente")
	}
}

func createCotacao(data *EconomiaAPIUsdBrlResponse) (*Cotacao, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*10)
	defer cancel()

	cotacao := Cotacao{
		Valor: data.Usdbrl.Bid,
	}

	err := DB.WithContext(ctx).Create(&cotacao).Error
	if err != nil {
		return nil, fmt.Errorf("Erro ao salvar cotacao: %v\n", err)
	}

	return &cotacao, nil
}

func getUSDtoBRL() (*EconomiaAPIUsdBrlResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*200)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)
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

	var data EconomiaAPIUsdBrlResponse
	err = json.Unmarshal(body, &data)
	if err != nil {
		return nil, fmt.Errorf("Erro ao converter resposta: %v\n", err)
	}

	return &data, nil
}
