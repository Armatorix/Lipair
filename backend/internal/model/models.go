package model

import (
	"database/sql"
	"time"
)

type User struct {
	ID              string         `json:"id" db:"id"`
	Email           sql.NullString `json:"email,omitempty" db:"email"`
	PasswordHash    sql.NullString `json:"-" db:"password_hash"`
	Name            string         `json:"name" db:"name"`
	GoogleID        sql.NullString `json:"google_id,omitempty" db:"google_id"`
	LichessID       sql.NullString `json:"lichess_id,omitempty" db:"lichess_id"`
	LichessUsername sql.NullString `json:"lichess_username,omitempty" db:"lichess_username"`
	AvatarURL       sql.NullString `json:"avatar_url,omitempty" db:"avatar_url"`
	CreatedAt       time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at" db:"updated_at"`
}

type Tournament struct {
	ID            string         `json:"id" db:"id"`
	Name          string         `json:"name" db:"name"`
	Description   sql.NullString `json:"description,omitempty" db:"description"`
	OrganizerID   string         `json:"organizer_id" db:"organizer_id"`
	PairingSystem string         `json:"pairing_system" db:"pairing_system"`
	Status        string         `json:"status" db:"status"`
	RoundsCount   sql.NullInt64  `json:"rounds_count,omitempty" db:"rounds_count"`
	TimeControl   sql.NullString `json:"time_control,omitempty" db:"time_control"`
	StartDate     sql.NullTime   `json:"start_date,omitempty" db:"start_date"`
	EndDate       sql.NullTime   `json:"end_date,omitempty" db:"end_date"`
	Settings      []byte         `json:"settings" db:"settings"`
	CreatedAt     time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at" db:"updated_at"`
}

type TournamentPlayer struct {
	ID           string    `json:"id" db:"id"`
	TournamentID string    `json:"tournament_id" db:"tournament_id"`
	PlayerID     string    `json:"player_id" db:"player_id"`
	Seed         *int      `json:"seed,omitempty" db:"seed"`
	Rating       *int      `json:"rating,omitempty" db:"rating"`
	Score        float64   `json:"score" db:"score"`
	JoinedAt     time.Time `json:"joined_at" db:"joined_at"`
	// Joined fields
	Name      string         `json:"name,omitempty"`
	Email     sql.NullString `json:"email,omitempty"`
	AvatarURL sql.NullString `json:"avatar_url,omitempty"`
}

type Round struct {
	ID           string    `json:"id" db:"id"`
	TournamentID string    `json:"tournament_id" db:"tournament_id"`
	RoundNumber  int       `json:"round_number" db:"round_number"`
	Status       string    `json:"status" db:"status"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

type Pairing struct {
	ID            string         `json:"id" db:"id"`
	RoundID       string         `json:"round_id" db:"round_id"`
	WhitePlayerID sql.NullString `json:"white_player_id,omitempty" db:"white_player_id"`
	BlackPlayerID sql.NullString `json:"black_player_id,omitempty" db:"black_player_id"`
	Result        string         `json:"result" db:"result"`
	BoardNumber   *int           `json:"board_number,omitempty" db:"board_number"`
	CreatedAt     time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at" db:"updated_at"`
	// Joined fields
	WhitePlayerName string `json:"white_player_name,omitempty"`
	BlackPlayerName string `json:"black_player_name,omitempty"`
}

type OAuthState struct {
	State        string    `db:"state"`
	CodeVerifier string    `db:"code_verifier"`
	ExpiresAt    time.Time `db:"expires_at"`
}

// Claims for JWT
type Claims struct {
	UserID string `json:"user_id"`
}
