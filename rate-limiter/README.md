# Rate Limiter (Go)

Rate limiter configurável em Go para servidores HTTP com persistência Redis.

## Como Executar

### 1. Clone e entre no diretório:
```bash
cd rate-limiter
```

### 2. Suba os serviços:
```bash
docker compose up --build
```

### 3. Verifique se está rodando:
```bash
curl http://localhost:8080/
# Deve retornar: {"status":"ok"}
```

## Como Testar

### Teste Básico
```bash
# Teste simples
curl http://localhost:8080/
```

### Teste de Rate Limiting por IP
```bash
# Envie 3 requisições rápidas (limite: 2/s)
for i in {1..3}; do 
  echo "Requisição $i:"
  curl -i http://localhost:8080/
  echo -e "\n---"
done
```

**Resultado esperado:**
- Requisições 1-2: HTTP 200
- Requisição 3: HTTP 429

### Teste de Rate Limiting por Token
```bash
# Teste com token abc123 (limite: 5/s)
for i in {1..6}; do 
  echo "Requisição $i (com token):"
  curl -i -H "API_KEY: abc123" http://localhost:8080/
  echo -e "\n---"
done
```

**Resultado esperado:**
- Requisições 1-5: HTTP 200
- Requisição 6: HTTP 429

### Teste de Diferentes Tokens
```bash
# Teste token premium (10 req/s)
for i in {1..11}; do 
  curl -i -H "API_KEY: premium" http://localhost:8080/
done

# Teste token free (3 req/s)
for i in {1..4}; do 
  curl -i -H "API_KEY: free" http://localhost:8080/
done
```

## Configuração

Todas as configurações estão no `docker-compose.yml`:

- **Limite padrão**: 2 req/s
- **Tempo de bloqueio**: 10 segundos
- **Token overrides**: 
  - `abc123`: 5 req/s, bloqueio 10s
  - `premium`: 10 req/s, bloqueio 20s
  - `free`: 3 req/s, bloqueio 15s

## Troubleshooting

### Não está limitando?
Edite o `docker-compose.yml` e reduza os limites:
```yaml
- RATE_LIMIT_RPS=1                      # 1 req/s
- RATE_LIMIT_BLOCK_SECONDS=5            # 5s de bloqueio
```

### Erro de conexão?
```bash
# Verifique logs
docker compose logs app
docker compose logs redis

# Verifique status
docker compose ps
```

## Estrutura do Projeto

```
rate-limiter/
├── cmd/server/          # Servidor HTTP
├── internal/            # Lógica da aplicação
├── Dockerfile           # Container da aplicação
└── docker-compose.yml   # Stack completo (app + Redis)
```


