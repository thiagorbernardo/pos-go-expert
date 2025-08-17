package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"time"

	"github.com/gin-gonic/gin"
)

type ViaCEPResponse struct {
	Cep         string `json:"cep"`
	Logradouro  string `json:"logradouro"`
	Complemento string `json:"complemento"`
	Bairro      string `json:"bairro"`
	Localidade  string `json:"localidade"`
	Uf          string `json:"uf"`
	Erro        bool   `json:"erro"`
}

type WeatherAPIResponse struct {
	Location struct {
		Name    string `json:"name"`
		Region  string `json:"region"`
		Country string `json:"country"`
	} `json:"location"`
	Current struct {
		TempC float64 `json:"temp_c"`
	} `json:"current"`
}

type WeatherResponse struct {
	TempC float64 `json:"temp_C"`
	TempF float64 `json:"temp_F"`
	TempK float64 `json:"temp_K"`
}

type ErrorResponse struct {
	Message string `json:"message"`
}

func main() {
	// Configurar o modo do Gin
	gin.SetMode(gin.ReleaseMode)

	// Criar o router
	r := gin.Default()

	// Middleware para logging de todas as requests
	r.Use(loggingMiddleware())

	// Rota para consultar o clima por CEP
	r.GET("/weather/:cep", getWeatherByCEP)

	// Rota de health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Obter a porta do ambiente ou usar padrão
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("🚀 Servidor iniciando na porta %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal("❌ Erro ao iniciar o servidor:", err)
	}
}

// Middleware para logging de requests
func loggingMiddleware() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		// Log colorido e detalhado
		var statusColor string
		switch {
		case param.StatusCode >= 200 && param.StatusCode < 300:
			statusColor = "✅" // Verde para sucesso
		case param.StatusCode >= 400 && param.StatusCode < 500:
			statusColor = "⚠️" // Amarelo para erro do cliente
		case param.StatusCode >= 500:
			statusColor = "❌" // Vermelho para erro do servidor
		default:
			statusColor = "ℹ️" // Azul para outros
		}

		return fmt.Sprintf("%s [%s] %s %s %d %s %s %s\n",
			statusColor,
			param.TimeStamp.Format("2006/01/02 - 15:04:05"),
			param.Method,
			param.Path,
			param.StatusCode,
			param.Latency,
			param.ClientIP,
			param.ErrorMessage,
		)
	})
}

func getWeatherByCEP(c *gin.Context) {
	startTime := time.Now()
	cep := c.Param("cep")

	log.Printf("🌍 Iniciando consulta de clima para CEP: %s", cep)

	// Validar formato do CEP (8 dígitos)
	if !isValidCEP(cep) {
		log.Printf("❌ CEP inválido: %s", cep)
		c.JSON(422, ErrorResponse{Message: "invalid zipcode"})
		return
	}

	log.Printf("✅ CEP validado: %s", cep)

	// Buscar informações do CEP usando ViaCEP
	location, err := getLocationByCEP(cep)
	if err != nil {
		log.Printf("❌ Erro ao buscar localização para CEP %s: %v", cep, err)
		c.JSON(404, ErrorResponse{Message: "can not find zipcode"})
		return
	}

	log.Printf("📍 Localização encontrada: %s", location)

	// Buscar informações do clima
	weather, err := getWeatherByLocation(location)
	if err != nil {
		log.Printf("❌ Erro ao buscar clima para localização %s: %v", location, err)
		c.JSON(500, ErrorResponse{Message: "error fetching weather data"})
		return
	}

	elapsed := time.Since(startTime)
	log.Printf("🌤️ Clima obtido com sucesso para %s em %v: TempC=%.1f°C, TempF=%.1f°F, TempK=%.1fK",
		location, elapsed, weather.TempC, weather.TempF, weather.TempK)

	// Retornar resposta de sucesso
	c.JSON(200, weather)
}

func isValidCEP(cep string) bool {
	// Remover caracteres não numéricos
	cep = regexp.MustCompile(`[^\d]`).ReplaceAllString(cep, "")

	// Verificar se tem exatamente 8 dígitos
	return len(cep) == 8
}

func getLocationByCEP(cep string) (string, error) {
	// Limpar CEP para busca
	cep = regexp.MustCompile(`[^\d]`).ReplaceAllString(cep, "")

	// URL da ViaCEP
	url := fmt.Sprintf("http://viacep.com.br/ws/%s/json/", cep)

	log.Printf("🔍 Consultando ViaCEP: %s", url)

	var response ViaCEPResponse
	if err := fetchJSON(url, &response); err != nil {
		log.Printf("❌ Erro na consulta ViaCEP: %v", err)
		return "", fmt.Errorf("erro ao buscar CEP: %w", err)
	}

	log.Printf("📡 Resposta ViaCEP: CEP=%s, Localidade=%s, UF=%s, Erro=%v",
		response.Cep, response.Localidade, response.Uf, response.Erro)

	// Verificar se a ViaCEP retornou erro
	if response.Erro {
		log.Printf("❌ ViaCEP retornou erro para CEP: %s", cep)
		return "", fmt.Errorf("CEP não encontrado")
	}

	// Verificar se os campos necessários estão preenchidos
	if response.Localidade == "" {
		log.Printf("❌ Campo obrigatório vazio: Localidade='%s'",
			response.Localidade)
		return "", fmt.Errorf("localidade não encontrada para o CEP")
	}

	return response.Localidade, nil
}

func fetchJSON(url string, target interface{}) error {
	startTime := time.Now()
	log.Printf("🌐 Fazendo requisição HTTP para: %s", url)

	resp, err := http.Get(url)
	if err != nil {
		log.Printf("❌ Erro na requisição HTTP: %v", err)
		return err
	}
	defer resp.Body.Close()

	elapsed := time.Since(startTime)
	log.Printf("📡 Resposta HTTP recebida: Status=%d, Tempo=%v", resp.StatusCode, elapsed)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("❌ Erro ao ler corpo da resposta: %v", err)
		return err
	}

	log.Printf("📄 Tamanho da resposta: %d bytes", len(body))

	if resp.StatusCode != 200 {
		log.Printf("❌ Status HTTP não OK: %d, Corpo: %s", resp.StatusCode, string(body))
		return fmt.Errorf("status HTTP não OK: %d", resp.StatusCode)
	}

	if err := json.Unmarshal(body, target); err != nil {
		log.Printf("❌ Erro ao fazer parse JSON: %v, Corpo: %s", err, string(body))
		return err
	}

	log.Printf("✅ JSON parseado com sucesso")
	return nil
}

func getWeatherByLocation(location string) (*WeatherResponse, error) {
	// Obter API key do ambiente
	apiKey := os.Getenv("WEATHER_API_KEY")
	if apiKey == "" {
		log.Printf("❌ WEATHER_API_KEY não configurada")
		return nil, fmt.Errorf("WEATHER_API_KEY não configurada")
	}

	// URL da WeatherAPI
	url := fmt.Sprintf("http://api.weatherapi.com/v1/current.json?key=%s&q=%s&aqi=no",
		apiKey, location)

	log.Printf("🌤️ Consultando WeatherAPI para: %s", location)

	resp, err := http.Get(url)
	if err != nil {
		log.Printf("❌ Erro na requisição HTTP para WeatherAPI: %v", err)
		return nil, fmt.Errorf("erro na requisição HTTP: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Printf("❌ WeatherAPI retornou status: %d", resp.StatusCode)
		return nil, fmt.Errorf("erro ao buscar clima, status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("❌ Erro ao ler resposta da WeatherAPI: %v", err)
		return nil, fmt.Errorf("erro ao ler resposta: %w", err)
	}

	log.Printf("📡 WeatherAPI resposta recebida: %d bytes", len(body))

	var weatherResp WeatherAPIResponse
	if err := json.Unmarshal(body, &weatherResp); err != nil {
		log.Printf("❌ Erro ao decodificar JSON da WeatherAPI: %v", err)
		return nil, fmt.Errorf("erro ao decodificar JSON: %w", err)
	}

	log.Printf("🌍 Dados da WeatherAPI: Cidade=%s, Região=%s, País=%s, TempC=%.1f",
		weatherResp.Location.Name, weatherResp.Location.Region,
		weatherResp.Location.Country, weatherResp.Current.TempC)

	// Converter temperaturas
	tempC := weatherResp.Current.TempC
	tempF := tempC*1.8 + 32
	tempK := tempC + 273.15

	log.Printf("🌡️ Temperaturas convertidas: C=%.1f°C, F=%.1f°F, K=%.1fK", tempC, tempF, tempK)

	return &WeatherResponse{
		TempC: tempC,
		TempF: tempF,
		TempK: tempK,
	}, nil
}
