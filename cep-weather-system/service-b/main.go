package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/zipkin"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	"go.opentelemetry.io/otel/trace"
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
		Name string `json:"name"`
	} `json:"location"`
	Current struct {
		TempC float64 `json:"temp_c"`
	} `json:"current"`
}

type WeatherResponse struct {
	City  string  `json:"city"`
	TempC float64 `json:"temp_C"`
	TempF float64 `json:"temp_F"`
	TempK float64 `json:"temp_K"`
}

type ErrorResponse struct {
	Message string `json:"message"`
}

var tracer trace.Tracer

func main() {
	if err := initTracer(); err != nil {
		log.Fatal(err)
	}
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.Use(loggingMiddleware())
	r.GET("/weather/:cep", getWeatherByCEP)
	r.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })
	port := os.Getenv("PORT")
	if port == "" {
		port = os.Getenv("SERVICE_B_PORT")
	}
	if port == "" {
		port = "8080"
	}
	log.Printf("service-b listening on :%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}

func initTracer() error {
	zipkinURL := os.Getenv("ZIPKIN_URL")
	if zipkinURL == "" {
		zipkinURL = "http://localhost:9411/api/v2/spans"
	}
	exporter, err := zipkin.New(zipkinURL)
	if err != nil {
		return err
	}
	res, err := resource.New(context.Background(), resource.WithAttributes(
		attribute.String("service.name", "weather-orchestrator"),
		attribute.String("service.version", "1.0.0"),
	))
	if err != nil {
		return err
	}
	tp := sdktrace.NewTracerProvider(sdktrace.WithBatcher(exporter), sdktrace.WithResource(res))
	otel.SetTracerProvider(tp)
	tracer = otel.Tracer("weather-orchestrator")
	return nil
}

func loggingMiddleware() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("%s %s %d %s %s\n", param.Method, param.Path, param.StatusCode, param.Latency, param.ClientIP)
	})
}

func getWeatherByCEP(c *gin.Context) {
	ctx, span := tracer.Start(c.Request.Context(), "getWeatherByCEP")
	defer span.End()
	start := time.Now()
	cep := c.Param("cep")
	if !isValidCEP(cep) {
		c.JSON(422, ErrorResponse{Message: "invalid zipcode"})
		return
	}
	city, err := getLocationByCEP(ctx, cep)
	if err != nil {
		c.JSON(404, ErrorResponse{Message: "can not find zipcode"})
		return
	}
	weather, err := getWeatherByLocation(ctx, city)
	if err != nil {
		c.JSON(500, ErrorResponse{Message: "error fetching weather data"})
		return
	}
	span.SetAttributes(attribute.Int64("response_time_ms", time.Since(start).Milliseconds()))
	c.JSON(200, weather)
}

func isValidCEP(cep string) bool {
	cep = regexp.MustCompile(`[^\d]`).ReplaceAllString(cep, "")
	return len(cep) == 8
}

func getLocationByCEP(ctx context.Context, cep string) (string, error) {
	ctx, span := tracer.Start(ctx, "getLocationByCEP")
	defer span.End()
	cep = regexp.MustCompile(`[^\d]`).ReplaceAllString(cep, "")
	url := fmt.Sprintf("http://viacep.com.br/ws/%s/json/", cep)
	var resp ViaCEPResponse
	if err := fetchJSON(ctx, url, &resp); err != nil {
		return "", err
	}
	if resp.Erro || resp.Localidade == "" {
		return "", fmt.Errorf("not found")
	}
	return resp.Localidade, nil
}

func fetchJSON(ctx context.Context, url string, target any) error {
	ctx, span := tracer.Start(ctx, "fetchJSON")
	defer span.End()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		b, _ := io.ReadAll(res.Body)
		return fmt.Errorf("status %d: %s", res.StatusCode, string(b))
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(body, target)
}

func getWeatherByLocation(ctx context.Context, location string) (*WeatherResponse, error) {
	ctx, span := tracer.Start(ctx, "getWeatherByLocation")
	defer span.End()
	apiKey := os.Getenv("WEATHER_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("WEATHER_API_KEY not set")
	}
	url := fmt.Sprintf("http://api.weatherapi.com/v1/current.json?key=%s&q=%s&aqi=no", apiKey, location)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("weather api status %d", res.StatusCode)
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	var w WeatherAPIResponse
	if err := json.Unmarshal(body, &w); err != nil {
		return nil, err
	}
	tempC := w.Current.TempC
	tempF := tempC*1.8 + 32
	tempK := tempC + 273.15
	return &WeatherResponse{City: w.Location.Name, TempC: tempC, TempF: tempF, TempK: tempK}, nil
}
