package handler

import (
	"database/sql"
	"net/http"
	"strings"
	"time"

	"github.com/Armatorix/ChessMgr/backend/internal/auth"
	"github.com/Armatorix/ChessMgr/backend/internal/db"
	"github.com/Armatorix/ChessMgr/backend/internal/model"
	"github.com/labstack/echo/v4"
)

type UserHandler struct {
	db *db.DB
}

func NewUserHandler(db *db.DB) *UserHandler {
	return &UserHandler{db: db}
}

func (h *UserHandler) GetMe(c echo.Context) error {
	userID := auth.GetUserID(c)
	var user model.User
	err := h.db.QueryRow(
		`SELECT id, email, password_hash, name, google_id, lichess_id, lichess_username, avatar_url, created_at, updated_at
		 FROM users WHERE id = $1`,
		userID,
	).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.Name,
		&user.GoogleID, &user.LichessID, &user.LichessUsername, &user.AvatarURL,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return echo.NewHTTPError(http.StatusNotFound, "user not found")
	}
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "database error")
	}
	return c.JSON(http.StatusOK, &user)
}

func (h *UserHandler) UpdateMe(c echo.Context) error {
	userID := auth.GetUserID(c)

	type updateRequest struct {
		Name      string `json:"name"`
		AvatarURL string `json:"avatar_url"`
	}
	var req updateRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	req.Name = strings.TrimSpace(req.Name)

	_, err := h.db.Exec(
		`UPDATE users SET name = COALESCE(NULLIF($1, ''), name), avatar_url = CASE WHEN $2 != '' THEN $2 ELSE avatar_url END, updated_at = $3 WHERE id = $4`,
		req.Name, req.AvatarURL, time.Now(), userID,
	)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to update user")
	}

	var user model.User
	err = h.db.QueryRow(
		`SELECT id, email, password_hash, name, google_id, lichess_id, lichess_username, avatar_url, created_at, updated_at
		 FROM users WHERE id = $1`,
		userID,
	).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.Name,
		&user.GoogleID, &user.LichessID, &user.LichessUsername, &user.AvatarURL,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "database error")
	}
	return c.JSON(http.StatusOK, &user)
}
