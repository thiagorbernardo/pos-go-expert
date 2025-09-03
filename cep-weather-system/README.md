# Sistema de Clima por CEP

Sistema distribuído em Go que recebe um CEP (código postal brasileiro), identifica a cidade e retorna o clima atual com temperaturas em Celsius, Fahrenheit e Kelvin. Implementa OpenTelemetry (OTEL) e Zipkin para rastreamento distribuído.

## Como Executar

### 1. Clone e entre no diretório:
```bash
cd cep-weather-system
```

### 2. Configure a chave da API de clima:
```bash
export WEATHER_API_KEY="sua_chave_api_aqui"
```

### 3. Suba os serviços:
```bash
docker-compose up --build
```

## Como Testar

### Teste de CEP Válido
```bash
# Teste com CEP de Curitiba (80220000) - Retorna sucesso
curl -X POST http://localhost:8080/cep \
  -H "Content-Type: application/json" \
  -d '{"cep": "80220000"}'
```

### Teste de CEP Inválido
```bash
# CEP com menos de 8 dígitos
curl -X POST http://localhost:8080/cep \
  -H "Content-Type: application/json" \
  -d '{"cep": "123"}'

# CEP com caracteres inválidos
curl -X POST http://localhost:8080/cep \
  -H "Content-Type: application/json" \
  -d '{"cep": "abc12345"}'
```

### Teste de CEP Não Encontrado
```bash
# CEP que não existe
curl -X POST http://localhost:8080/cep \
  -H "Content-Type: application/json" \
  -d '{"cep": "99999999"}'
```

## Respostas Esperadas

### Sucesso (200):
```json
{
  "city": "Curitiba",
  "temp_C": 28.5,
  "temp_F": 83.3,
  "temp_K": 301.65
}
```

### CEP Inválido (422):
```json
{
  "message": "invalid zipcode"
}
```

### CEP Não Encontrado (404):
```json
{
  "message": "can not find zipcode"
}
```

## Configuração

- **Porta**: 8080
- **WEATHER_API_KEY**: Sua chave da WeatherAPI

## Fórmulas de Conversão

- **Fahrenheit**: `F = C × 1,8 + 32`
- **Kelvin**: `K = C + 273,15`