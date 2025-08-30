# Sistema de Clima por CEP

Sistema distribuído em Go que recebe um CEP (código postal brasileiro), identifica a cidade e retorna o clima atual com temperaturas em Celsius, Fahrenheit e Kelvin. Implementa OpenTelemetry (OTEL) e Zipkin para rastreamento distribuído.

## Visão Geral

Este sistema é composto por dois serviços que trabalham em conjunto:
- **Serviço A**: Recebe CEPs via POST, valida a entrada e encaminha para o Serviço B
- **Serviço B**: Resolve CEP para cidade, busca dados meteorológicos e retorna temperaturas formatadas
- **Zipkin**: Coletor e interface de rastreamento distribuído

## Início Rápido

### 1. Configurar Chave da API de Clima
```bash
export WEATHER_API_KEY="sua_chave_api_aqui"
```

### 2. Iniciar Serviços
```bash
docker-compose up --build
```

### 3. Acessar Serviços
- **Serviço A**: http://localhost:8081
- **Serviço B**: http://localhost:8080  
- **Zipkin UI**: http://localhost:9411

## Como Usar

### Endpoint Principal - POST `/cep`

O sistema recebe CEPs via POST e retorna dados meteorológicos da cidade correspondente.

**Exemplo de Uso:**
```bash
curl -X POST http://localhost:8081/cep \
  -H "Content-Type: application/json" \
  -d '{"cep": "80220000"}'
```

**Resposta de Sucesso (200):**
```json
{
  "city": "Curitiba",
  "temp_C": 22.5,
  "temp_F": 72.5,
  "temp_K": 295.65
}
```

## Exemplos de Teste

### CEPs Válidos
```bash
# Curitiba
curl -X POST http://localhost:8081/cep \
  -H "Content-Type: application/json" \
  -d '{"cep": "80220000"}'
```

### Casos de Erro
```bash
# CEP Inválido (muito curto)
curl -X POST http://localhost:8081/cep \
  -H "Content-Type: application/json" \
  -d '{"cep": "123"}'

# CEP Inválido (com letras)
curl -X POST http://localhost:8081/cep \
  -H "Content-Type: application/json" \
  -d '{"cep": "12345abc"}'

# Campo CEP ausente
curl -X POST http://localhost:8081/cep \
  -H "Content-Type: application/json" \
  -d '{"city": "teste"}'

# JSON inválido
curl -X POST http://localhost:8081/cep \
  -H "Content-Type: application/json" \
  -d '{"cep": "80220000"'
```

### Verificação de Saúde
```bash
# Saúde do Serviço A
curl http://localhost:8081/health

# Saúde do Serviço B
curl http://localhost:8080/health
```

## Configuração

### Variáveis de Ambiente
Configure estas variáveis antes de executar `docker-compose up`:

```bash
# Obrigatório: Obtenha em https://www.weatherapi.com/
export WEATHER_API_KEY="sua_chave_api_aqui"

# Opcional: Sobrescrever padrões
export SERVICE_A_PORT=8081
export SERVICE_B_PORT=8080
export ZIPKIN_URL=http://zipkin:9411/api/v2/spans
```

## Funcionalidades

- ✅ **Validação de CEP**: Validação numérica de 8 dígitos
- ✅ **Integração com API de Clima**: Dados de temperatura atual
- ✅ **Conversão de Temperatura**: Celsius, Fahrenheit, Kelvin
- ✅ **Rastreamento Distribuído**: OpenTelemetry + Zipkin
- ✅ **Tratamento de Erros**: Códigos de status HTTP apropriados
- ✅ **Suporte Docker**: Implantação e teste fáceis

## APIs Externas Utilizadas

- **ViaCEP**: Resolução de código postal brasileiro (https://viacep.com.br/)
- **WeatherAPI**: Dados meteorológicos atuais (https://www.weatherapi.com/)

## Fórmulas de Conversão de Temperatura

- **Fahrenheit**: `F = C × 1,8 + 32`
- **Kelvin**: `K = C + 273,15`
