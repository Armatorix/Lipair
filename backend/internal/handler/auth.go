package handler

import (
	"database/sql"
	"net/http"
	"strings"
	"time"

	"github.com/Armatorix/ChessMgr/backend/internal/auth"
	"github.com/Armatorix/ChessMgr/backend/internal/db"
	"github.com/Armatorix/ChessMgr/backend/internal/model"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	db          *db.DB
	jwtManager  *auth.JWTManager
	googleAuth  *auth.GoogleAuth
	lichessAuth *auth.LichessAuth
	frontendURL string
}

func NewAuthHandler(db *db.DB, jwtManager *auth.JWTManager, googleAuth *auth.GoogleAuth, lichessAuth *auth.LichessAuth, frontendURL string) *AuthHandler {
	return &AuthHandler{
		db:          db,
		jwtManager:  jwtManager,
		googleAuth:  googleAuth,
		lichessAuth: lichessAuth,
		frontendURL: frontendURL,
	}
}

type registerRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type authResponse struct {
	Token string       `json:"token"`
	User  *model.User  `json:"user"`
}

func (h *AuthHandler) Register(c echo.Context) error {
	var req registerRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	req.Name = strings.TrimSpace(req.Name)
	req.Email = strings.TrimSpace(req.Email)
	if req.Name == "" || req.Email == "" || req.Password == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "name, email and password are required")
	}
	if len(req.Password) < 6 {
		return echo.NewHTTPError(http.StatusBadRequest, "password must be at least 6 characters")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to hash password")
	}

	user := &model.User{
		ID:        uuid.New().String(),
		Name:      req.Name,
		Email:     sql.NullString{String: req.Email, Valid: true},
		PasswordHash: sql.NullString{String: string(hash), Valid: true},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	_, err = h.db.Exec(
		`INSERT INTO users (id, email, password_hash, name, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		user.ID, user.Email.String, user.PasswordHash.String, user.Name, user.CreatedAt, user.UpdatedAt,
	)
	if err != nil {
		if strings.Contains(err.Error(), "unique") {
			return echo.NewHTTPError(http.StatusConflict, "email already registered")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to create user")
	}

	token, err := h.jwtManager.Generate(user.ID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to generate token")
	}

	return c.JSON(http.StatusCreated, authResponse{Token: token, User: user})
}

func (h *AuthHandler) Login(c echo.Context) error {
	var req loginRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	req.Email = strings.TrimSpace(req.Email)
	if req.Email == "" || req.Password == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "email and password are required")
	}

	var user model.User
	err := h.db.QueryRow(
		`SELECT id, email, password_hash, name, google_id, lichess_id, lichess_username, avatar_url, created_at, updated_at
		 FROM users WHERE email = $1`,
		req.Email,
	).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.Name,
		&user.GoogleID, &user.LichessID, &user.LichessUsername, &user.AvatarURL,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid credentials")
	}
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "database error")
	}
	if !user.PasswordHash.Valid {
		return echo.NewHTTPError(http.StatusUnauthorized, "no password set for this account")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash.String), []byte(req.Password)); err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid credentials")
	}

	token, err := h.jwtManager.Generate(user.ID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to generate token")
	}

	return c.JSON(http.StatusOK, authResponse{Token: token, User: &user})
}

func (h *AuthHandler) GoogleLogin(c echo.Context) error {
	url, _, err := h.googleAuth.GetAuthURL()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get auth URL")
	}
	return c.Redirect(http.StatusTemporaryRedirect, url)
}

func (h *AuthHandler) GoogleCallback(c echo.Context) error {
	code := c.QueryParam("code")
	state := c.QueryParam("state")
	if code == "" || state == "" {
		return c.Redirect(http.StatusTemporaryRedirect, h.frontendURL+"/login?error=missing_params")
	}

	userInfo, err := h.googleAuth.ExchangeCode(c.Request().Context(), code, state)
	if err != nil {
		return c.Redirect(http.StatusTemporaryRedirect, h.frontendURL+"/login?error=oauth_failed")
	}

	user, err := h.findOrCreateGoogleUser(userInfo)
	if err != nil {
		return c.Redirect(http.StatusTemporaryRedirect, h.frontendURL+"/login?error=user_creation_failed")
	}

	token, err := h.jwtManager.Generate(user.ID)
	if err != nil {
		return c.Redirect(http.StatusTemporaryRedirect, h.frontendURL+"/login?error=token_failed")
	}

	return c.Redirect(http.StatusTemporaryRedirect, h.frontendURL+"/auth/callback?token="+token)
}

func (h *AuthHandler) findOrCreateGoogleUser(info *auth.GoogleUserInfo) (*model.User, error) {
	var user model.User
	err := h.db.QueryRow(
		`SELECT id, email, password_hash, name, google_id, lichess_id, lichess_username, avatar_url, created_at, updated_at
		 FROM users WHERE google_id = $1`,
		info.ID,
	).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.Name,
		&user.GoogleID, &user.LichessID, &user.LichessUsername, &user.AvatarURL,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if err == nil {
		return &user, nil
	}
	if err != sql.ErrNoRows {
		return nil, err
	}

	// Try to find by email
	if info.Email != "" {
		err = h.db.QueryRow(
			`SELECT id, email, password_hash, name, google_id, lichess_id, lichess_username, avatar_url, created_at, updated_at
			 FROM users WHERE email = $1`,
			info.Email,
		).Scan(
			&user.ID, &user.Email, &user.PasswordHash, &user.Name,
			&user.GoogleID, &user.LichessID, &user.LichessUsername, &user.AvatarURL,
			&user.CreatedAt, &user.UpdatedAt,
		)
		if err == nil {
			// Link Google account
			_, err = h.db.Exec(
				`UPDATE users SET google_id = $1, avatar_url = $2, updated_at = NOW() WHERE id = $3`,
				info.ID, info.Picture, user.ID,
			)
			if err != nil {
				return nil, err
			}
			user.GoogleID = sql.NullString{String: info.ID, Valid: true}
			return &user, nil
		}
		if err != sql.ErrNoRows {
			return nil, err
		}
	}

	// Create new user
	newUser := &model.User{
		ID:        uuid.New().String(),
		Name:      info.Name,
		GoogleID:  sql.NullString{String: info.ID, Valid: true},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if info.Email != "" {
		newUser.Email = sql.NullString{String: info.Email, Valid: true}
	}
	if info.Picture != "" {
		newUser.AvatarURL = sql.NullString{String: info.Picture, Valid: true}
	}

	_, err = h.db.Exec(
		`INSERT INTO users (id, email, name, google_id, avatar_url, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		newUser.ID, nullStringVal(newUser.Email), newUser.Name,
		newUser.GoogleID.String, nullStringVal(newUser.AvatarURL),
		newUser.CreatedAt, newUser.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return newUser, nil
}

func (h *AuthHandler) LichessLogin(c echo.Context) error {
	url, _, err := h.lichessAuth.GetAuthURL()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get auth URL")
	}
	return c.Redirect(http.StatusTemporaryRedirect, url)
}

func (h *AuthHandler) LichessCallback(c echo.Context) error {
	code := c.QueryParam("code")
	state := c.QueryParam("state")
	if code == "" || state == "" {
		return c.Redirect(http.StatusTemporaryRedirect, h.frontendURL+"/login?error=missing_params")
	}

	userInfo, err := h.lichessAuth.ExchangeCode(c.Request().Context(), code, state)
	if err != nil {
		return c.Redirect(http.StatusTemporaryRedirect, h.frontendURL+"/login?error=oauth_failed")
	}

	user, err := h.findOrCreateLichessUser(userInfo)
	if err != nil {
		return c.Redirect(http.StatusTemporaryRedirect, h.frontendURL+"/login?error=user_creation_failed")
	}

	token, err := h.jwtManager.Generate(user.ID)
	if err != nil {
		return c.Redirect(http.StatusTemporaryRedirect, h.frontendURL+"/login?error=token_failed")
	}

	return c.Redirect(http.StatusTemporaryRedirect, h.frontendURL+"/auth/callback?token="+token)
}

func (h *AuthHandler) findOrCreateLichessUser(info *auth.LichessUserInfo) (*model.User, error) {
	var user model.User
	err := h.db.QueryRow(
		`SELECT id, email, password_hash, name, google_id, lichess_id, lichess_username, avatar_url, created_at, updated_at
		 FROM users WHERE lichess_id = $1`,
		info.ID,
	).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.Name,
		&user.GoogleID, &user.LichessID, &user.LichessUsername, &user.AvatarURL,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if err == nil {
		return &user, nil
	}
	if err != sql.ErrNoRows {
		return nil, err
	}

	// Create new user
	newUser := &model.User{
		ID:              uuid.New().String(),
		Name:            info.Username,
		LichessID:       sql.NullString{String: info.ID, Valid: true},
		LichessUsername: sql.NullString{String: info.Username, Valid: true},
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	_, err = h.db.Exec(
		`INSERT INTO users (id, name, lichess_id, lichess_username, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		newUser.ID, newUser.Name, newUser.LichessID.String, newUser.LichessUsername.String,
		newUser.CreatedAt, newUser.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return newUser, nil
}

func nullStringVal(ns sql.NullString) interface{} {
	if ns.Valid {
		return ns.String
	}
	return nil
}
