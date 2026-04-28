package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type LichessAuth struct {
	clientID    string
	redirectURL string
	db          *sql.DB
}

type LichessUserInfo struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}

func NewLichessAuth(clientID, redirectURL string, db *sql.DB) *LichessAuth {
	return &LichessAuth{
		clientID:    clientID,
		redirectURL: redirectURL,
		db:          db,
	}
}

func (l *LichessAuth) GetAuthURL() (string, string, error) {
	state, err := generateRandomState()
	if err != nil {
		return "", "", err
	}
	codeVerifier, err := generateCodeVerifier()
	if err != nil {
		return "", "", err
	}
	codeChallenge := generateCodeChallenge(codeVerifier)
	if err := l.saveStateWithVerifier(state, codeVerifier); err != nil {
		return "", "", err
	}
	params := url.Values{
		"response_type":         {"code"},
		"client_id":             {l.clientID},
		"redirect_uri":          {l.redirectURL},
		"scope":                 {"email:read"},
		"state":                 {state},
		"code_challenge_method": {"S256"},
		"code_challenge":        {codeChallenge},
	}
	authURL := "https://lichess.org/oauth?" + params.Encode()
	return authURL, state, nil
}

func (l *LichessAuth) ExchangeCode(ctx context.Context, code, state string) (*LichessUserInfo, error) {
	codeVerifier, err := l.getAndDeleteState(state)
	if err != nil {
		return nil, fmt.Errorf("invalid state: %w", err)
	}
	data := url.Values{
		"grant_type":    {"authorization_code"},
		"code":          {code},
		"redirect_uri":  {l.redirectURL},
		"client_id":     {l.clientID},
		"code_verifier": {codeVerifier},
	}
	req, err := http.NewRequestWithContext(ctx, "POST", "https://lichess.org/api/token",
		strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("exchanging code: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var tokenResp struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, err
	}
	return l.getUserInfo(ctx, tokenResp.AccessToken)
}

func (l *LichessAuth) getUserInfo(ctx context.Context, accessToken string) (*LichessUserInfo, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://lichess.org/api/account", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var userInfo LichessUserInfo
	if err := json.Unmarshal(body, &userInfo); err != nil {
		return nil, err
	}
	return &userInfo, nil
}

func (l *LichessAuth) saveStateWithVerifier(state, codeVerifier string) error {
	_, err := l.db.Exec(
		"INSERT INTO oauth_states (state, code_verifier, expires_at) VALUES ($1, $2, $3)",
		state, codeVerifier, time.Now().Add(10*time.Minute),
	)
	return err
}

func (l *LichessAuth) getAndDeleteState(state string) (string, error) {
	var codeVerifier string
	var expiresAt time.Time
	err := l.db.QueryRow(
		"DELETE FROM oauth_states WHERE state = $1 RETURNING code_verifier, expires_at",
		state,
	).Scan(&codeVerifier, &expiresAt)
	if err == sql.ErrNoRows {
		return "", fmt.Errorf("state not found")
	}
	if err != nil {
		return "", err
	}
	if time.Now().After(expiresAt) {
		return "", fmt.Errorf("state expired")
	}
	return codeVerifier, nil
}

func generateCodeVerifier() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func generateCodeChallenge(verifier string) string {
	h := sha256.New()
	h.Write([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(h.Sum(nil))
}
