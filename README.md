# ChessMgr

A complete chess tournament management platform supporting multiple pairing systems, OAuth authentication, and real-time score tracking.

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Backend | Go 1.22, Echo v4, PostgreSQL |
| Frontend | React 18, TypeScript, Vite, Tailwind CSS |
| Auth | JWT, Google OAuth2, Lichess OAuth2 (PKCE) |
| Database | PostgreSQL 16 with custom migrations |
| Deployment | Docker, Docker Compose, Docker Swarm |

## Features

- **Pairing Systems**: Round Robin (Berger), Swiss, Knockout, Scheveningen, Manual
- **Authentication**: Email/password, Google OAuth, Lichess OAuth (PKCE)
- **Tournament Management**: Create, start, manage rounds and pairings
- **Score Tracking**: Automatic score updates (1pt win, 0.5pt draw, bye)
- **REST API**: Full OpenAPI 3.0 specification in `api/openapi.yaml`

## Quick Start (Development)

### Prerequisites
- Docker & Docker Compose

```bash
# Clone the repo
git clone https://github.com/Armatorix/ChessMgr.git
cd ChessMgr

# Copy and configure environment (optional for OAuth)
cp .env.example .env

# Start all services
docker compose up --build
```

- Frontend: http://localhost:3000
- Backend API: http://localhost:8080/api

### Local Development without Docker

**Backend:**
```bash
cd backend
go mod download
DATABASE_URL="postgres://chessmgr:chessmgr@localhost:5432/chessmgr?sslmode=disable" \
  JWT_SECRET="local-dev-secret" \
  go run ./cmd/server
```

**Frontend:**
```bash
cd frontend
npm install
npm run dev
```

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `SERVER_PORT` | `8080` | Backend port |
| `DATABASE_URL` | postgres://... | PostgreSQL connection string |
| `JWT_SECRET` | `change-me-in-production` | JWT signing secret |
| `FRONTEND_URL` | `http://localhost:3000` | Frontend URL for OAuth redirects |
| `CORS_ORIGINS` | `http://localhost:3000` | Comma-separated allowed origins |
| `GOOGLE_CLIENT_ID` | `` | Google OAuth client ID |
| `GOOGLE_CLIENT_SECRET` | `` | Google OAuth client secret |
| `LICHESS_CLIENT_ID` | `` | Lichess OAuth client ID |

## API Reference

Full spec: [`api/openapi.yaml`](api/openapi.yaml)

### Key Endpoints

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| POST | `/api/auth/register` | No | Email/password registration |
| POST | `/api/auth/login` | No | Email/password login |
| GET | `/api/auth/google` | No | Start Google OAuth flow |
| GET | `/api/auth/lichess` | No | Start Lichess OAuth flow |
| GET | `/api/users/me` | JWT | Get current user |
| GET | `/api/tournaments` | No | List all tournaments |
| POST | `/api/tournaments` | JWT | Create tournament |
| POST | `/api/tournaments/:id/start` | JWT | Start tournament (creates round 1) |
| POST | `/api/tournaments/:id/players` | JWT | Add player |
| POST | `/api/tournaments/:id/rounds` | JWT | Create next round |
| PUT | `/api/tournaments/:id/rounds/:rid/pairings/:pid` | JWT | Update pairing result |

## Production Deployment (Docker Swarm)

```bash
# Create secrets
echo "your-jwt-secret" | docker secret create jwt_secret -
echo "your-google-secret" | docker secret create google_client_secret -
echo "your-lichess-secret" | docker secret create lichess_client_secret -

# Deploy stack
docker stack deploy -c docker-compose.prod.yml chessmgr
```

## Project Structure

```
ChessMgr/
├── api/
│   └── openapi.yaml          # OpenAPI 3.0 specification
├── backend/
│   ├── cmd/server/main.go    # Entry point
│   ├── internal/
│   │   ├── auth/             # JWT, Google, Lichess auth
│   │   ├── config/           # Configuration loading
│   │   ├── db/               # Database + migrations
│   │   ├── handler/          # HTTP handlers
│   │   └── model/            # Data models
│   └── Dockerfile
├── frontend/
│   ├── src/
│   │   ├── api/              # API client
│   │   ├── components/       # Shared components
│   │   ├── hooks/            # Custom hooks
│   │   ├── pages/            # Page components
│   │   └── store/            # Zustand state
│   └── Dockerfile
├── docker-compose.yml        # Development
└── docker-compose.prod.yml   # Production (Swarm)
```