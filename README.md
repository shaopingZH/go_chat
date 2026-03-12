# Go-Chat

Go-Chat is a real-time chat backend built with Go + Gin + WebSocket + MySQL + Redis.

## Features

- User register/login with JWT
- User profile endpoint
- Group create/join/list endpoints
- WebSocket private/group chat
- Message persistence and history query

## Quick Start

1. Copy environment variables:

   ```bash
   cp .env.example .env
   ```

2. Prepare MySQL and Redis, then run SQL schema:

   ```sql
   sql/mysql/V001__init_schema.sql
   ```

3. Start service:

   ```bash
   go run ./cmd/server
   ```

Server default address: `:8080`

## Run With Docker

1. Copy Docker environment variables to a dedicated local Docker env file and replace placeholder secrets before first use:

   ```bash
   cp .env.docker.example .env.docker
   ```

2. Build and start all services (app + MySQL + Redis):

   ```bash
   docker compose --env-file .env.docker up -d --build
   ```

   If port `8080` is already occupied on host:

   ```bash
   GO_CHAT_APP_PORT=18080 docker compose --env-file .env.docker up -d --build
   ```

3. Check service status:

   ```bash
   docker compose --env-file .env.docker ps
   ```

4. Watch backend logs:

   ```bash
   docker compose --env-file .env.docker logs -f app
   ```

5. Stop services:

   ```bash
   docker compose --env-file .env.docker down
   ```

6. Stop and remove database/cache volumes:

   ```bash
   docker compose --env-file .env.docker down -v
   ```

Default exposed address: `http://127.0.0.1:8080`

Default Docker DB credentials (for local development only):

- MySQL user: `gochat`
- MySQL password: `gochat123`
- MySQL database: `go_chat`

## Frontend (Vue 3)

The frontend project is in `web/`, built with Vite + Vue 3 + TypeScript + Tailwind + Pinia + Vue Router, with a Telegram Web-like chat UI.

1. Install dependencies:

   ```bash
   cd web
   npm install
   ```

2. Start dev server:

   ```bash
   npm run dev
   ```

Dev server default address: `http://127.0.0.1:4173` (if occupied, Vite auto-selects next port)

3. Build production bundle:

   ```bash
   npm run build
   ```

The Vite config proxies `/api` and `/ws` to `http://localhost:8080` by default.

You can override frontend dev host/port via env vars:

- `VITE_DEV_HOST` (default `127.0.0.1`)
- `VITE_DEV_PORT` (default `4173`)

## Environment Variables

- `SERVER_ADDR` (default `:8080`)
- `MYSQL_DSN`
- `REDIS_ADDR`
- `REDIS_PASSWORD`
- `REDIS_DB`
- `JWT_SECRET`
- `JWT_EXPIRE_HOURS`

## HTTP APIs

- `POST /api/auth/register`
- `POST /api/auth/login`
- `GET /api/users/me`
- `POST /api/groups`
- `POST /api/groups/:id/join`
- `GET /api/groups`
- `GET /api/messages?target_id={id}&type=private|group&last_msg_id={id}&limit={n}`
- `GET /api/conversations/private?limit={n}`

Protected APIs require `Authorization: Bearer <token>`.

## WebSocket

- URL: `/ws?token=<jwt_token>`

Send:

```json
{
  "type": "chat",
  "payload": {
    "target_id": 2,
    "chat_type": "private",
    "content": "hello"
  }
}
```

Receive:

```json
{
  "type": "message",
  "payload": {
    "id": 1,
    "sender": { "id": 1, "username": "alice", "avatar": "" },
    "target_id": 2,
    "chat_type": "private",
    "content": "hello",
    "created_at": "2026-02-10T12:00:00Z"
  }
}
```
