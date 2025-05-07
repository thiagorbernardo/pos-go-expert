package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type BrasilAPIResponse struct {
	Cep          string `json:"cep"`
	State        string `json:"state"`
	City         string `json:"city"`
	Neighborhood string `json:"neighborhood"`
	Street       string `json:"street"`
}

type ViaCEPResponse struct {
	Cep         string `json:"cep"`
	Logradouro  string `json:"logradouro"`
	Complemento string `json:"complemento"`
	Bairro      string `json:"bairro"`
	Localidade  string `json:"localidade"`
	Uf          string `json:"uf"`
}

type APIResponse struct {
	API     string
	Address interface{}
	Err     error
}

func main() {
	cep := "11850000" // Exemplo de CEP
	result := fetchFastestCEP(cep)
	if result.Err != nil {
		fmt.Printf("Erro: %v\n", result.Err)
		return
	}

	fmt.Printf("Resultado mais rápido obtido da API: %s\n", result.API)
	switch addr := result.Address.(type) {
	case BrasilAPIResponse:
		fmt.Printf("CEP: %s\nEstado: %s\nCidade: %s\nBairro: %s\nRua: %s\n",
			addr.Cep, addr.State, addr.City, addr.Neighborhood, addr.Street)
	case ViaCEPResponse:
		fmt.Printf("CEP: %s\nEstado: %s\nCidade: %s\nBairro: %s\nRua: %s\n",
			addr.Cep, addr.Uf, addr.Localidade, addr.Bairro, addr.Logradouro)
	}
}

func fetchFastestCEP(cep string) APIResponse {
	ch := make(chan APIResponse)
	timeout := time.After(1 * time.Second)

	// Goroutine para BrasilAPI
	go func() {
		url := fmt.Sprintf("https://brasilapi.com.br/api/cep/v1/%s", cep)
		var response BrasilAPIResponse
		err := fetchJSON(url, &response)
		ch <- APIResponse{API: "BrasilAPI", Address: response, Err: err}
	}()

	// Goroutine para ViaCEP
	go func() {
		url := fmt.Sprintf("http://viacep.com.br/ws/%s/json/", cep)
		var response ViaCEPResponse
		err := fetchJSON(url, &response)
		ch <- APIResponse{API: "ViaCEP", Address: response, Err: err}
	}()

	// Aguarda o resultado mais rápido ou timeout
	select {
	case result := <-ch:
		if result.Err != nil {
			// Se a primeira resposta falhar, tenta a segunda
			select {
			case result = <-ch:
				return result
			case <-timeout:
				return APIResponse{Err: fmt.Errorf("timeout ao buscar CEP")}
			}
		}
		return result
	case <-timeout:
		return APIResponse{Err: fmt.Errorf("timeout ao buscar CEP")}
	}
}

func fetchJSON(url string, target interface{}) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	return json.Unmarshal(body, target)
} 