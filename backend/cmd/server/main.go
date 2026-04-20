package main

import (
	"log"
	"net/http"

	"github.com/Armatorix/ChessMgr/backend/internal/auth"
	"github.com/Armatorix/ChessMgr/backend/internal/config"
	"github.com/Armatorix/ChessMgr/backend/internal/db"
	"github.com/Armatorix/ChessMgr/backend/internal/handler"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	// Load .env if it exists (ignore error if not found)
	_ = godotenv.Load()

	cfg := config.Load()

	database, err := db.New(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("connecting to database: %v", err)
	}
	defer database.Close()

	if err := database.Migrate(); err != nil {
		log.Fatalf("running migrations: %v", err)
	}

	jwtManager := auth.NewJWTManager(cfg.JWTSecret)
	googleAuth := auth.NewGoogleAuth(cfg.GoogleClientID, cfg.GoogleClientSecret, cfg.GoogleRedirectURL, database.DB)
	lichessAuth := auth.NewLichessAuth(cfg.LichessClientID, cfg.LichessRedirectURL, database.DB)

	authHandler := handler.NewAuthHandler(database, jwtManager, googleAuth, lichessAuth, cfg.FrontendURL)
	userHandler := handler.NewUserHandler(database)
	tournamentHandler := handler.NewTournamentHandler(database)
	roundHandler := handler.NewRoundHandler(database)

	e := echo.New()
	e.HideBanner = true

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     cfg.CORSOrigins,
		AllowMethods:     []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodOptions},
		AllowHeaders:     []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
		AllowCredentials: true,
	}))

	jwtMiddleware := auth.JWTMiddleware(jwtManager)

	// Auth routes
	e.POST("/api/auth/register", authHandler.Register)
	e.POST("/api/auth/login", authHandler.Login)
	e.GET("/api/auth/google", authHandler.GoogleLogin)
	e.GET("/api/auth/google/callback", authHandler.GoogleCallback)
	e.GET("/api/auth/lichess", authHandler.LichessLogin)
	e.GET("/api/auth/lichess/callback", authHandler.LichessCallback)

	// User routes (protected)
	e.GET("/api/users/me", userHandler.GetMe, jwtMiddleware)
	e.PUT("/api/users/me", userHandler.UpdateMe, jwtMiddleware)

	// Tournament routes
	e.GET("/api/tournaments", tournamentHandler.ListTournaments)
	e.POST("/api/tournaments", tournamentHandler.CreateTournament, jwtMiddleware)
	e.GET("/api/tournaments/:id", tournamentHandler.GetTournament)
	e.PUT("/api/tournaments/:id", tournamentHandler.UpdateTournament, jwtMiddleware)
	e.DELETE("/api/tournaments/:id", tournamentHandler.DeleteTournament, jwtMiddleware)
	e.POST("/api/tournaments/:id/start", tournamentHandler.StartTournament, jwtMiddleware)
	e.GET("/api/tournaments/:id/players", tournamentHandler.ListPlayers)
	e.POST("/api/tournaments/:id/players", tournamentHandler.AddPlayer, jwtMiddleware)
	e.DELETE("/api/tournaments/:id/players/:playerId", tournamentHandler.RemovePlayer, jwtMiddleware)

	// Round routes
	e.GET("/api/tournaments/:id/rounds", roundHandler.ListRounds)
	e.POST("/api/tournaments/:id/rounds", roundHandler.CreateRound, jwtMiddleware)
	e.GET("/api/tournaments/:id/rounds/:roundId", roundHandler.GetRound)
	e.GET("/api/tournaments/:id/rounds/:roundId/pairings", roundHandler.ListPairings)
	e.PUT("/api/tournaments/:id/rounds/:roundId/pairings/:pairingId", roundHandler.UpdatePairing, jwtMiddleware)

	log.Printf("Starting server on port %s", cfg.ServerPort)
	if err := e.Start(":" + cfg.ServerPort); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server error: %v", err)
	}
}
