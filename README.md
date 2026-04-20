# ChessMgr

A chess tournament manager platform built with Go (Echo), React, and PostgreSQL.

## Features

- **Authentication**: Email/password, Google OAuth, Lichess OAuth (PKCE)
- **Tournament formats**:
  - Round Robin
  - Knockout (single elimination)
  - Swiss system
  - Scheveningen system
  - Manual pairings
- **Tournament management**: Create, configure, start tournaments
- **Player management**: Add/remove players with optional ratings and seeds
- **Round & pairing tracking**: Auto-generate pairings, record results, track scores

## Tech Stack

| Layer       | Technology                                       |
|-------------|--------------------------------------------------|
| Backend     | Go 1.22 + Echo v4                                |
| Database    | PostgreSQL 16                                    |
| Frontend    | React 18 + TypeScript + Vite                     |
| Styling     | TailwindCSS + HeadlessUI                         |
| Auth        | JWT + Google OAuth2 + Lichess OAuth2 (PKCE)      |
| API spec    | OpenAPI 3.0                                      |
| Codegen     | oapi-codegen (backend), openapi-typescript-codegen (frontend) |
| Deployment  | Docker Compose (dev) / Docker Swarm (prod)       |

## Project Structure

```
.
├── api/                  # OpenAPI specification
│   └── openapi.yaml
├── backend/              # Go backend
│   ├── cmd/server/       # Entry point
│   ├── internal/
│   │   ├── auth/         # JWT, Google OAuth, Lichess OAuth
│   │   ├── config/       # Configuration (env vars + Docker secrets)
│   │   ├── db/           # Database connection + migrations
│   │   ├── handler/      # HTTP handlers
│   │   └── model/        # Data models
│   ├── Dockerfile
│   └── go.mod
├── frontend/             # React frontend
│   ├── src/
│   │   ├── api/          # API client (axios)
│   │   ├── components/   # Shared UI components
│   │   ├── hooks/        # React hooks
│   │   ├── pages/        # Page components
│   │   └── store/        # Zustand state management
│   ├── Dockerfile
│   └── package.json
├── docker-compose.yml      # Local development
└── docker-compose.prod.yml # Docker Swarm production
```

## Quick Start (Local Development)

### Prerequisites
- Docker & Docker Compose

### 1. Clone and configure

```bash
git clone https://github.com/Armatorix/ChessMgr.git
cd ChessMgr
cp .env.example .env
# Edit .env with your OAuth credentials (optional for basic testing)
```

### 2. Start all services

```bash
docker compose up --build
```

- Frontend: http://localhost:3000
- Backend API: http://localhost:8080
- PostgreSQL: localhost:5432

## Development Without Docker

### Backend

```bash
cd backend
# Start PostgreSQL separately (or use docker compose up postgres)
go run ./cmd/server
```

### Frontend

```bash
cd frontend
npm install
npm run dev
```

## OAuth Setup

### Google OAuth
1. Go to [Google Cloud Console](https://console.developers.google.com)
2. Create a project and enable the Google+ API
3. Create OAuth 2.0 credentials
4. Set authorized redirect URI: `http://localhost:8080/api/auth/google/callback`
5. Add `GOOGLE_CLIENT_ID` and `GOOGLE_CLIENT_SECRET` to your `.env`

### Lichess OAuth
1. Go to [Lichess OAuth Apps](https://lichess.org/account/oauth/app)
2. Create a new application
3. Set callback URL: `http://localhost:8080/api/auth/lichess/callback`
4. Add `LICHESS_CLIENT_ID` to your `.env` (Lichess uses PKCE, no secret needed)

## Production Deployment (Docker Swarm)

### 1. Initialize swarm
```bash
docker swarm init
```

### 2. Create secrets
```bash
echo "your-strong-jwt-secret" | docker secret create jwt_secret -
echo "your-postgres-password" | docker secret create postgres_password -
echo "your-google-client-secret" | docker secret create google_client_secret -
# Lichess uses PKCE, no secret needed
```

### 3. Deploy the stack
```bash
export FRONTEND_URL=https://chessmgr.example.com
export CORS_ORIGINS=https://chessmgr.example.com
export GOOGLE_CLIENT_ID=your-google-client-id
export GOOGLE_REDIRECT_URL=https://chessmgr.example.com/api/auth/google/callback
export LICHESS_CLIENT_ID=your-lichess-client-id
export LICHESS_REDIRECT_URL=https://chessmgr.example.com/api/auth/lichess/callback

docker stack deploy -c docker-compose.prod.yml chessmgr
```

## API Reference

The OpenAPI spec is at `api/openapi.yaml`. Key endpoints:

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/auth/register` | Register with email/password |
| POST | `/api/auth/login` | Login with email/password |
| GET | `/api/auth/google` | Start Google OAuth flow |
| GET | `/api/auth/lichess` | Start Lichess OAuth flow |
| GET | `/api/users/me` | Get current user profile |
| GET | `/api/tournaments` | List all tournaments |
| POST | `/api/tournaments` | Create a tournament |
| POST | `/api/tournaments/:id/start` | Start tournament (generates round 1) |
| POST | `/api/tournaments/:id/players` | Add player to tournament |
| POST | `/api/tournaments/:id/rounds` | Generate next round |
| PUT | `/api/tournaments/:id/rounds/:rId/pairings/:pId` | Update pairing result |

## Pairing Systems

| System | Description |
|--------|-------------|
| `round_robin` | Every player plays every other player (Berger table rotation) |
| `knockout` | Single elimination bracket seeded by rating |
| `swiss` | Pair players with similar scores each round |
| `scheveningen` | Two teams play each other in all combinations |
| `manual` | Organizer sets pairings manually |

## Generating API Clients

### Backend (oapi-codegen)
```bash
cd backend
go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest
oapi-codegen --config oapi-codegen.yaml ../api/openapi.yaml
```

### Frontend (openapi-typescript-codegen)
```bash
cd frontend
npx openapi-typescript-codegen --input ../api/openapi.yaml --output src/api/generated --client axios
```


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