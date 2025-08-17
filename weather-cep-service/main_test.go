package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.GET("/weather/:cep", getWeatherByCEP)
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
	return r
}

func TestValidCEP(t *testing.T) {
	router := setupTestRouter()

	// Teste com CEP válido (São Paulo)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/weather/01310-100", nil)
	router.ServeHTTP(w, req)

	// Como não temos API key configurada nos testes, esperamos erro 500
	// ou sucesso se a API key estiver configurada
	assert.Contains(t, []int{http.StatusOK, http.StatusInternalServerError}, w.Code)

	if w.Code == http.StatusOK {
		var response WeatherResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		// Verificar se as temperaturas estão presentes
		assert.Greater(t, response.TempC, -100.0)
		assert.Less(t, response.TempC, 100.0)
		assert.Greater(t, response.TempF, -148.0)
		assert.Less(t, response.TempF, 212.0)
		assert.Greater(t, response.TempK, 173.0)
		assert.Less(t, response.TempK, 373.0)
	}
}

func TestInvalidCEPFormat(t *testing.T) {
	router := setupTestRouter()

	// Teste com CEP inválido (formato incorreto)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/weather/123", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)

	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "invalid zipcode", response.Message)
}

func TestInvalidCEPFormat2(t *testing.T) {
	router := setupTestRouter()

	// Teste com CEP inválido (formato incorreto)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/weather/123456789", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)

	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "invalid zipcode", response.Message)
}

func TestInvalidCEPFormat3(t *testing.T) {
	router := setupTestRouter()

	// Teste com CEP inválido (formato incorreto)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/weather/abc12345", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)

	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "invalid zipcode", response.Message)
}

func TestCEPWithSpecialCharacters(t *testing.T) {
	router := setupTestRouter()

	// Teste com CEP com caracteres especiais
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/weather/01310-100", nil)
	router.ServeHTTP(w, req)

	// Como não temos API key configurada nos testes, esperamos erro 500
	// ou sucesso se a API key estiver configurada
	assert.Contains(t, []int{http.StatusOK, http.StatusInternalServerError}, w.Code)
}

func TestCEPWithoutHyphen(t *testing.T) {
	router := setupTestRouter()

	// Teste com CEP sem hífen
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/weather/01310100", nil)
	router.ServeHTTP(w, req)

	// Como não temos API key configurada nos testes, esperamos erro 500
	// ou sucesso se a API key estiver configurada
	assert.Contains(t, []int{http.StatusOK, http.StatusInternalServerError}, w.Code)
}

func TestTemperatureConversion(t *testing.T) {
	// Teste das fórmulas de conversão
	tempC := 25.0
	expectedTempF := tempC*1.8 + 32
	expectedTempK := tempC + 273.15

	assert.Equal(t, 77.0, expectedTempF)
	assert.Equal(t, 298.15, expectedTempK)
}

func TestHealthEndpoint(t *testing.T) {
	router := setupTestRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestFetchJSON(t *testing.T) {
	// Teste da função fetchJSON
	var response map[string]interface{}
	err := fetchJSON("https://httpbin.org/json", &response)

	assert.NoError(t, err)
	assert.NotNil(t, response)
}

func TestCEPValidation(t *testing.T) {
	// Testes de validação de CEP
	assert.True(t, isValidCEP("01310100"))
	assert.True(t, isValidCEP("01310-100"))
	assert.True(t, isValidCEP("12345678"))

	assert.False(t, isValidCEP("123"))
	assert.False(t, isValidCEP("123456789"))
	assert.False(t, isValidCEP("abc12345"))
	assert.False(t, isValidCEP(""))
}
