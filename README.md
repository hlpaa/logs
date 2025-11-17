# Servidor Go simples

Este é um servidor HTTP mínimo escrito em Go.

Como executar:

```bash
# rodar diretamente (requere Go instalado)
go run server.go
```

Endpoints:

- `/` — retorna uma mensagem de boas-vindas em texto (ex.: "Olá! Este é um servidor Go").
- `POST /logs` — adiciona um log. Aceita JSON estruturado como:
  ```json
  {
    "timestamp": "2025-11-15T20:15:00Z",
    "severity": "ERROR",
    "message": "Falha ao conectar"
  }
  ```
  Retorna 201 quando criado.
- `GET /logs` — retorna todos os logs em JSON (array de objetos).

Testar com curl:

```bash
# acessar home
curl http://localhost:8080/

# adicionar log (JSON estruturado)
curl -X POST -H "Content-Type: application/json" -d '{"timestamp":"2025-11-15T20:15:00Z","severity":"ERROR","message":"Falha ao conectar"}' http://localhost:8080/logs

# listar logs
curl http://localhost:8080/logs

# filtrar por severidade
curl "http://localhost:8080/logs?severity=ERROR"

# filtrar por intervalo de tempo
curl "http://localhost:8080/logs?from=2025-11-15T00:00:00Z&to=2025-11-15T23:59:59Z"

# filtrar por intervalo + severidae
curl "http://localhost:8080/logs?from=2025-11-15T20:00:00Z&to=2025-11-15T23:00:00Z&severity=ERROR"
```
