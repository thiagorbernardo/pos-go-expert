# Weather CEP Service

Sistema em Go que recebe um CEP, identifica a cidade e retorna o clima atual em múltiplas escalas de temperatura.

## Funcionalidades

- ✅ Validação de CEP (8 dígitos)
- ✅ Consulta de localização usando ViaCEP
- ✅ Consulta de clima atual via WeatherAPI
- ✅ Conversão automática de temperaturas (Celsius, Fahrenheit, Kelvin)
- ✅ Tratamento de erros adequado
- ✅ Health check endpoint
- ✅ Containerização com Docker
- ✅ Testes automatizados

## APIs Utilizadas

- **CEP**: ViaCEP (https://viacep.com.br/)
- **Clima**: WeatherAPI (https://www.weatherapi.com/)

## Endpoints

### GET /weather/{cep}
Consulta o clima para um CEP específico.

**Parâmetros:**
- `cep`: CEP válido (8 dígitos, com ou sem hífen)

**Respostas:**

**Sucesso (200):**
```json
{
  "temp_C": 28.5,
  "temp_F": 83.3,
  "temp_K": 301.65
}
```

**CEP inválido (422):**
```json
{
  "message": "invalid zipcode"
}
```

**CEP não encontrado (404):**
```json
{
  "message": "can not find zipcode"
}
```

### GET /health
Health check do serviço.

**Resposta (200):**
```json
{
  "status": "ok"
}
```

## 🧪 Teste da Aplicação

### **URL de Teste (Google Cloud Run):**
```
https://weather-cep-service-1060290311769.us-central1.run.app
```

### **Exemplos de Teste:**
```bash
# Health check
curl https://weather-cep-service-1060290311769.us-central1.run.app/health

# Consulta de clima por CEP (Curitiba)
curl https://weather-cep-service-1060290311769.us-central1.run.app/weather/80220-000

# CEP inválido
curl https://weather-cep-service-1060290311769.us-central1.run.app/weather/123
```

## Instalação e Execução

### Pré-requisitos
- Go 1.21+
- Docker e Docker Compose
- Chave da API WeatherAPI

### 1. Configuração
Edite o arquivo `docker-compose.yml` e substitua `sua_chave_aqui` pela sua chave da WeatherAPI:

```yaml
environment:
  - WEATHER_API_KEY=sua_chave_real_aqui
```

### 2. Execução Local
```bash
# Instalar dependências
go mod tidy

# Executar testes
go test -v

# Executar aplicação
go run main.go
```

### 3. Execução com Docker
```bash
# Build e execução
docker-compose up --build

# Apenas execução (após build)
docker-compose up

# Execução em background
docker-compose up -d
```

## Testes

```bash
# Executar todos os testes
go test -v

# Executar testes com coverage
go test -v -cover

# Executar testes específicos
go test -v -run TestValidCEP
```

## Deploy no Google Cloud Run

### 1. Configurar Google Cloud
```bash
# Instalar gcloud CLI
# Autenticar
gcloud auth login

# Configurar projeto
gcloud config set project SEU_PROJETO_ID
```

### 2. Build e Deploy
```bash
# Build da imagem
gcloud builds submit --tag gcr.io/SEU_PROJETO_ID/weather-cep-service

# Deploy no Cloud Run
gcloud run deploy weather-cep-service \
  --image gcr.io/SEU_PROJETO_ID/weather-cep-service \
  --platform managed \
  --region us-central1 \
  --allow-unauthenticated \
  --set-env-vars WEATHER_API_KEY=sua_chave_aqui
```

### 3. Configurar variáveis de ambiente
No console do Google Cloud Run, configure:
- `WEATHER_API_KEY`: Sua chave da WeatherAPI

## Fórmulas de Conversão

- **Fahrenheit**: F = C × 1.8 + 32
- **Kelvin**: K = C + 273.15

## Estrutura do Projeto

```
weather-cep-service/
├── main.go              # Servidor principal
├── main_test.go         # Testes automatizados
├── Dockerfile           # Containerização
├── docker-compose.yml   # Orquestração local
├── go.mod               # Dependências Go
├── go.sum               # Checksums das dependências
├── env.example          # Exemplo de variáveis de ambiente
└── README.md            # Este arquivo
```

## Contribuição

1. Fork o projeto
2. Crie uma branch para sua feature
3. Commit suas mudanças
4. Push para a branch
5. Abra um Pull Request

## Licença

Este projeto está sob a licença MIT.
