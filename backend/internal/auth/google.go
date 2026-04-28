package auth

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type GoogleAuth struct {
	config *oauth2.Config
	db     *sql.DB
}

type GoogleUserInfo struct {
	ID      string `json:"id"`
	Email   string `json:"email"`
	Name    string `json:"name"`
	Picture string `json:"picture"`
}

func NewGoogleAuth(clientID, clientSecret, redirectURL string, db *sql.DB) *GoogleAuth {
	return &GoogleAuth{
		config: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  redirectURL,
			Scopes:       []string{"openid", "email", "profile"},
			Endpoint:     google.Endpoint,
		},
		db: db,
	}
}

func (g *GoogleAuth) GetAuthURL() (string, string, error) {
	state, err := generateRandomState()
	if err != nil {
		return "", "", err
	}
	if err := g.saveState(state, ""); err != nil {
		return "", "", err
	}
	url := g.config.AuthCodeURL(state, oauth2.AccessTypeOnline)
	return url, state, nil
}

func (g *GoogleAuth) ExchangeCode(ctx context.Context, code, state string) (*GoogleUserInfo, error) {
	if err := g.validateState(state); err != nil {
		return nil, fmt.Errorf("invalid state: %w", err)
	}
	token, err := g.config.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("exchanging code: %w", err)
	}
	client := g.config.Client(ctx, token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return nil, fmt.Errorf("getting user info: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var userInfo GoogleUserInfo
	if err := json.Unmarshal(body, &userInfo); err != nil {
		return nil, err
	}
	return &userInfo, nil
}

func (g *GoogleAuth) saveState(state, codeVerifier string) error {
	_, err := g.db.Exec(
		"INSERT INTO oauth_states (state, code_verifier, expires_at) VALUES ($1, $2, $3)",
		state, codeVerifier, time.Now().Add(10*time.Minute),
	)
	return err
}

func (g *GoogleAuth) validateState(state string) error {
	var expiresAt time.Time
	err := g.db.QueryRow(
		"DELETE FROM oauth_states WHERE state = $1 RETURNING expires_at", state,
	).Scan(&expiresAt)
	if err == sql.ErrNoRows {
		return fmt.Errorf("state not found")
	}
	if err != nil {
		return err
	}
	if time.Now().After(expiresAt) {
		return fmt.Errorf("state expired")
	}
	return nil
}

func generateRandomState() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
