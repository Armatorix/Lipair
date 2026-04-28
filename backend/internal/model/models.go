package model

import (
	"database/sql"
	"encoding/json"
	"time"
)

// User is the internal DB model with sql.Null* fields for scanning.
type User struct {
	ID              string         `db:"id"`
	Email           sql.NullString `db:"email"`
	PasswordHash    sql.NullString `db:"password_hash"`
	Name            string         `db:"name"`
	GoogleID        sql.NullString `db:"google_id"`
	LichessID       sql.NullString `db:"lichess_id"`
	LichessUsername sql.NullString `db:"lichess_username"`
	AvatarURL       sql.NullString `db:"avatar_url"`
	CreatedAt       time.Time      `db:"created_at"`
	UpdatedAt       time.Time      `db:"updated_at"`
}

// MarshalJSON produces clean JSON: optional string fields are either a string or omitted.
func (u User) MarshalJSON() ([]byte, error) {
	type Alias struct {
		ID              string    `json:"id"`
		Email           *string   `json:"email,omitempty"`
		Name            string    `json:"name"`
		GoogleID        *string   `json:"google_id,omitempty"`
		LichessID       *string   `json:"lichess_id,omitempty"`
		LichessUsername *string   `json:"lichess_username,omitempty"`
		AvatarURL       *string   `json:"avatar_url,omitempty"`
		CreatedAt       time.Time `json:"created_at"`
		UpdatedAt       time.Time `json:"updated_at"`
	}
	a := Alias{
		ID:        u.ID,
		Name:      u.Name,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
	if u.Email.Valid {
		a.Email = &u.Email.String
	}
	if u.GoogleID.Valid {
		a.GoogleID = &u.GoogleID.String
	}
	if u.LichessID.Valid {
		a.LichessID = &u.LichessID.String
	}
	if u.LichessUsername.Valid {
		a.LichessUsername = &u.LichessUsername.String
	}
	if u.AvatarURL.Valid {
		a.AvatarURL = &u.AvatarURL.String
	}
	return json.Marshal(a)
}

// Tournament is the internal DB model.
type Tournament struct {
	ID            string          `db:"id"`
	Name          string          `db:"name"`
	Description   sql.NullString  `db:"description"`
	OrganizerID   string          `db:"organizer_id"`
	SportType     string          `db:"sport_type"`
	PairingSystem string          `db:"pairing_system"`
	Status        string          `db:"status"`
	RoundsCount   sql.NullInt64   `db:"rounds_count"`
	TimeControl   sql.NullString  `db:"time_control"`
	StartDate     sql.NullTime    `db:"start_date"`
	EndDate       sql.NullTime    `db:"end_date"`
	Settings      json.RawMessage `db:"settings"`
	CreatedAt     time.Time       `db:"created_at"`
	UpdatedAt     time.Time       `db:"updated_at"`
}

// MarshalJSON produces clean JSON for a Tournament.
func (t Tournament) MarshalJSON() ([]byte, error) {
	type Alias struct {
		ID            string          `json:"id"`
		Name          string          `json:"name"`
		Description   *string         `json:"description,omitempty"`
		OrganizerID   string          `json:"organizer_id"`
		SportType     string          `json:"sport_type"`
		PairingSystem string          `json:"pairing_system"`
		Status        string          `json:"status"`
		RoundsCount   *int64          `json:"rounds_count,omitempty"`
		TimeControl   *string         `json:"time_control,omitempty"`
		StartDate     *time.Time      `json:"start_date,omitempty"`
		EndDate       *time.Time      `json:"end_date,omitempty"`
		Settings      json.RawMessage `json:"settings"`
		CreatedAt     time.Time       `json:"created_at"`
		UpdatedAt     time.Time       `json:"updated_at"`
	}
	settings := t.Settings
	if len(settings) == 0 {
		settings = json.RawMessage(`{}`)
	}
	a := Alias{
		ID:            t.ID,
		Name:          t.Name,
		OrganizerID:   t.OrganizerID,
		SportType:     t.SportType,
		PairingSystem: t.PairingSystem,
		Status:        t.Status,
		Settings:      settings,
		CreatedAt:     t.CreatedAt,
		UpdatedAt:     t.UpdatedAt,
	}
	if t.Description.Valid {
		a.Description = &t.Description.String
	}
	if t.RoundsCount.Valid {
		a.RoundsCount = &t.RoundsCount.Int64
	}
	if t.TimeControl.Valid {
		a.TimeControl = &t.TimeControl.String
	}
	if t.StartDate.Valid {
		a.StartDate = &t.StartDate.Time
	}
	if t.EndDate.Valid {
		a.EndDate = &t.EndDate.Time
	}
	return json.Marshal(a)
}

// TournamentPlayer joins tournament_players with user fields.
type TournamentPlayer struct {
	ID           string         `db:"id"`
	TournamentID string         `db:"tournament_id"`
	PlayerID     string         `db:"player_id"`
	Seed         *int           `db:"seed"`
	Rating       *int           `db:"rating"`
	Score        float64        `db:"score"`
	JoinedAt     time.Time      `db:"joined_at"`
	// Joined fields
	Name      string         `db:"name"`
	Email     sql.NullString `db:"email"`
	AvatarURL sql.NullString `db:"avatar_url"`
}

// MarshalJSON produces clean JSON for a TournamentPlayer.
func (p TournamentPlayer) MarshalJSON() ([]byte, error) {
	type Alias struct {
		ID           string    `json:"id"`
		TournamentID string    `json:"tournament_id"`
		PlayerID     string    `json:"player_id"`
		Seed         *int      `json:"seed,omitempty"`
		Rating       *int      `json:"rating,omitempty"`
		Score        float64   `json:"score"`
		JoinedAt     time.Time `json:"joined_at"`
		Name         string    `json:"name,omitempty"`
		Email        *string   `json:"email,omitempty"`
		AvatarURL    *string   `json:"avatar_url,omitempty"`
	}
	a := Alias{
		ID:           p.ID,
		TournamentID: p.TournamentID,
		PlayerID:     p.PlayerID,
		Seed:         p.Seed,
		Rating:       p.Rating,
		Score:        p.Score,
		JoinedAt:     p.JoinedAt,
		Name:         p.Name,
	}
	if p.Email.Valid {
		a.Email = &p.Email.String
	}
	if p.AvatarURL.Valid {
		a.AvatarURL = &p.AvatarURL.String
	}
	return json.Marshal(a)
}

// Round represents a tournament round.
type Round struct {
	ID           string    `json:"id" db:"id"`
	TournamentID string    `json:"tournament_id" db:"tournament_id"`
	RoundNumber  int       `json:"round_number" db:"round_number"`
	Status       string    `json:"status" db:"status"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

// Pairing represents a single game in a round.
type Pairing struct {
	ID            string         `db:"id"`
	RoundID       string         `db:"round_id"`
	WhitePlayerID sql.NullString `db:"white_player_id"`
	BlackPlayerID sql.NullString `db:"black_player_id"`
	Result        string         `db:"result"`
	BoardNumber   *int           `db:"board_number"`
	CreatedAt     time.Time      `db:"created_at"`
	UpdatedAt     time.Time      `db:"updated_at"`
	// Joined fields
	WhitePlayerName string `db:"white_player_name"`
	BlackPlayerName string `db:"black_player_name"`
}

// MarshalJSON produces clean JSON for a Pairing.
func (p Pairing) MarshalJSON() ([]byte, error) {
	type Alias struct {
		ID              string    `json:"id"`
		RoundID         string    `json:"round_id"`
		WhitePlayerID   *string   `json:"white_player_id,omitempty"`
		BlackPlayerID   *string   `json:"black_player_id,omitempty"`
		Result          string    `json:"result"`
		BoardNumber     *int      `json:"board_number,omitempty"`
		CreatedAt       time.Time `json:"created_at"`
		UpdatedAt       time.Time `json:"updated_at"`
		WhitePlayerName string    `json:"white_player_name,omitempty"`
		BlackPlayerName string    `json:"black_player_name,omitempty"`
	}
	a := Alias{
		ID:              p.ID,
		RoundID:         p.RoundID,
		Result:          p.Result,
		BoardNumber:     p.BoardNumber,
		CreatedAt:       p.CreatedAt,
		UpdatedAt:       p.UpdatedAt,
		WhitePlayerName: p.WhitePlayerName,
		BlackPlayerName: p.BlackPlayerName,
	}
	if p.WhitePlayerID.Valid {
		a.WhitePlayerID = &p.WhitePlayerID.String
	}
	if p.BlackPlayerID.Valid {
		a.BlackPlayerID = &p.BlackPlayerID.String
	}
	return json.Marshal(a)
}

// OAuthState stores PKCE state and verifier.
type OAuthState struct {
	State        string    `db:"state"`
	CodeVerifier string    `db:"code_verifier"`
	ExpiresAt    time.Time `db:"expires_at"`
}

// Claims for JWT
type Claims struct {
	UserID string `json:"user_id"`
}
