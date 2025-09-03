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

type CEPRequest struct {
	CEP string `json:"cep" binding:"required"`
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
		log.Fatal("failed to init tracer:", err)
	}

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.Use(loggingMiddleware())
	r.POST("/cep", handleCEPRequest)
	r.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })

	port := os.Getenv("PORT")
	if port == "" {
		port = os.Getenv("SERVICE_A_PORT")
		if port == "" {
			port = "8080"
		}
	}
	log.Printf("service-a listening on :%s", port)
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
		attribute.String("service.name", "cep-input-service"),
		attribute.String("service.version", "1.0.0"),
	))
	if err != nil {
		return err
	}
	tp := sdktrace.NewTracerProvider(sdktrace.WithBatcher(exporter), sdktrace.WithResource(res))
	otel.SetTracerProvider(tp)
	tracer = otel.Tracer("cep-input-service")
	return nil
}

func loggingMiddleware() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("%s %s %d %s %s\n", param.Method, param.Path, param.StatusCode, param.Latency, param.ClientIP)
	})
}

func handleCEPRequest(c *gin.Context) {
	ctx, span := tracer.Start(c.Request.Context(), "handleCEPRequest")
	defer span.End()

	start := time.Now()
	var req CEPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, ErrorResponse{Message: "invalid json format"})
		return
	}

	span.SetAttributes(attribute.String("cep", req.CEP))

	if !isValidCEP(req.CEP) {
		c.JSON(422, ErrorResponse{Message: "invalid zipcode"})
		return
	}

	weather, err := forwardToServiceB(ctx, req.CEP)
	if err != nil {
		// propagate known messages if they match spec
		switch err.Error() {
		case "invalid zipcode":
			c.JSON(422, ErrorResponse{Message: "invalid zipcode"})
		case "can not find zipcode":
			c.JSON(404, ErrorResponse{Message: "can not find zipcode"})
		default:
			c.JSON(500, ErrorResponse{Message: "internal server error"})
		}
		return
	}

	span.SetAttributes(attribute.Int64("response_time_ms", time.Since(start).Milliseconds()))
	c.JSON(200, weather)
}

func isValidCEP(cep string) bool {
	cep = regexp.MustCompile(`[^\d]`).ReplaceAllString(cep, "")
	return len(cep) == 8
}

func forwardToServiceB(ctx context.Context, cep string) (*WeatherResponse, error) {
	ctx, span := tracer.Start(ctx, "forwardToServiceB")
	defer span.End()
	base := os.Getenv("SERVICE_B_URL")
	if base == "" {
		base = "http://localhost:8081"
	}
	url := fmt.Sprintf("%s/weather/%s", base, cep)
	span.SetAttributes(attribute.String("service_b_url", url))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		var er ErrorResponse
		if json.Unmarshal(body, &er) == nil && er.Message != "" {
			return nil, fmt.Errorf(er.Message)
		}
		return nil, fmt.Errorf("service b error %d", resp.StatusCode)
	}
	var w WeatherResponse
	if err := json.Unmarshal(body, &w); err != nil {
		return nil, err
	}
	return &w, nil
}
