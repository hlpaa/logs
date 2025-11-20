# Servidor Go simples

Este é um servidor HTTP mínimo escrito em Go para gerenciamento de logs por usuário.

Como executar:

```bash
# rodar diretamente (requere Go instalado)
go run server.go
```

Endpoints:

- `POST /event` — adiciona um evento no log do usuário. Aceita JSON estruturado como:
  ```json
  {
    "user_id": "usuario123",
    "timestamp": "2025-11-15T20:15:00Z",
    "severity": "ERROR",
    "message": "Falha ao conectar"
  }
  ```
  Retorna 201 quando criado. Todos os campos são obrigatórios.

- `GET /logs/{user_id}` — retorna todos os logs de um usuário específico em JSON.

Testar com curl:

```bash
# adicionar evento (JSON estruturado)
curl -X POST -H "Content-Type: application/json" -d '{"user_id":"usuario123","timestamp":"2025-11-15T20:15:00Z","severity":"ERROR","message":"Falha ao conectar"}' http://localhost:8080/event

# listar logs de um usuário
curl http://localhost:8080/logs/usuario123

# filtrar por severidade
curl "http://localhost:8080/logs/usuario123?severity=ERROR"

# filtrar por intervalo de tempo
curl "http://localhost:8080/logs/usuario123?from=2025-11-15T00:00:00Z&to=2025-11-15T23:59:59Z"

# filtrar por intervalo + severidade
curl "http://localhost:8080/logs/usuario123?from=2025-11-15T20:00:00Z&to=2025-11-15T23:00:00Z&severity=ERROR"
```

## Estrutura dos dados

O servidor organiza os logs por usuário, onde cada usuário pode ter múltiplos eventos. A resposta de `/logs/{user_id}` retorna:

```json
{
  "user_id": "usuario123",
  "events": [
    {
      "timestamp": "2025-11-15T20:15:00Z",
      "severity": "ERROR",
      "message": "Falha ao conectar"
    }
  ]
}
```

## Filtros disponíveis

- `severity`: filtra eventos por nível de severidade (ex: ERROR, INFO, WARNING)
- `from`: filtra eventos a partir de um timestamp (formato RFC3339)
- `to`: filtra eventos até um timestamp (formato RFC3339)