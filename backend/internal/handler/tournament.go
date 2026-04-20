package handler

import (
	"database/sql"
	"math/rand"
	"net/http"
	"time"

	"github.com/Armatorix/ChessMgr/backend/internal/auth"
	"github.com/Armatorix/ChessMgr/backend/internal/db"
	"github.com/Armatorix/ChessMgr/backend/internal/model"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type TournamentHandler struct {
	db *db.DB
}

func NewTournamentHandler(db *db.DB) *TournamentHandler {
	return &TournamentHandler{db: db}
}

func (h *TournamentHandler) ListTournaments(c echo.Context) error {
	rows, err := h.db.Query(
		`SELECT id, name, description, organizer_id, pairing_system, status, rounds_count, time_control, start_date, end_date, settings, created_at, updated_at
		 FROM tournaments ORDER BY created_at DESC`,
	)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "database error")
	}
	defer rows.Close()

	tournaments := make([]model.Tournament, 0)
	for rows.Next() {
		var t model.Tournament
		if err := rows.Scan(
			&t.ID, &t.Name, &t.Description, &t.OrganizerID, &t.PairingSystem, &t.Status,
			&t.RoundsCount, &t.TimeControl, &t.StartDate, &t.EndDate, &t.Settings,
			&t.CreatedAt, &t.UpdatedAt,
		); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "scan error")
		}
		tournaments = append(tournaments, t)
	}
	return c.JSON(http.StatusOK, tournaments)
}

func (h *TournamentHandler) CreateTournament(c echo.Context) error {
	userID := auth.GetUserID(c)

	type createRequest struct {
		Name          string  `json:"name"`
		Description   string  `json:"description"`
		PairingSystem string  `json:"pairing_system"`
		RoundsCount   *int    `json:"rounds_count"`
		TimeControl   string  `json:"time_control"`
		StartDate     *string `json:"start_date"`
		EndDate       *string `json:"end_date"`
	}
	var req createRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if req.Name == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "name is required")
	}
	if req.PairingSystem == "" {
		req.PairingSystem = "round_robin"
	}

	validSystems := map[string]bool{"round_robin": true, "knockout": true, "swiss": true, "scheveningen": true, "manual": true}
	if !validSystems[req.PairingSystem] {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid pairing system")
	}

	t := model.Tournament{
		ID:            uuid.New().String(),
		Name:          req.Name,
		OrganizerID:   userID,
		PairingSystem: req.PairingSystem,
		Status:        "draft",
		Settings:      []byte("{}"),
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	if req.Description != "" {
		t.Description = sql.NullString{String: req.Description, Valid: true}
	}
	if req.RoundsCount != nil {
		t.RoundsCount = sql.NullInt64{Int64: int64(*req.RoundsCount), Valid: true}
	}
	if req.TimeControl != "" {
		t.TimeControl = sql.NullString{String: req.TimeControl, Valid: true}
	}
	if req.StartDate != nil && *req.StartDate != "" {
		parsed, err := time.Parse(time.RFC3339, *req.StartDate)
		if err == nil {
			t.StartDate = sql.NullTime{Time: parsed, Valid: true}
		}
	}
	if req.EndDate != nil && *req.EndDate != "" {
		parsed, err := time.Parse(time.RFC3339, *req.EndDate)
		if err == nil {
			t.EndDate = sql.NullTime{Time: parsed, Valid: true}
		}
	}

	_, err := h.db.Exec(
		`INSERT INTO tournaments (id, name, description, organizer_id, pairing_system, status, rounds_count, time_control, start_date, end_date, settings, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`,
		t.ID, t.Name, nullString(t.Description), t.OrganizerID, t.PairingSystem, t.Status,
		nullInt64(t.RoundsCount), nullString(t.TimeControl), nullTime(t.StartDate), nullTime(t.EndDate),
		t.Settings, t.CreatedAt, t.UpdatedAt,
	)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to create tournament")
	}

	return c.JSON(http.StatusCreated, &t)
}

func (h *TournamentHandler) GetTournament(c echo.Context) error {
	id := c.Param("id")
	var t model.Tournament
	err := h.db.QueryRow(
		`SELECT id, name, description, organizer_id, pairing_system, status, rounds_count, time_control, start_date, end_date, settings, created_at, updated_at
		 FROM tournaments WHERE id = $1`,
		id,
	).Scan(
		&t.ID, &t.Name, &t.Description, &t.OrganizerID, &t.PairingSystem, &t.Status,
		&t.RoundsCount, &t.TimeControl, &t.StartDate, &t.EndDate, &t.Settings,
		&t.CreatedAt, &t.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return echo.NewHTTPError(http.StatusNotFound, "tournament not found")
	}
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "database error")
	}
	return c.JSON(http.StatusOK, &t)
}

func (h *TournamentHandler) UpdateTournament(c echo.Context) error {
	userID := auth.GetUserID(c)
	id := c.Param("id")

	var t model.Tournament
	err := h.db.QueryRow(
		`SELECT id, organizer_id FROM tournaments WHERE id = $1`, id,
	).Scan(&t.ID, &t.OrganizerID)
	if err == sql.ErrNoRows {
		return echo.NewHTTPError(http.StatusNotFound, "tournament not found")
	}
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "database error")
	}
	if t.OrganizerID != userID {
		return echo.NewHTTPError(http.StatusForbidden, "not the organizer")
	}

	type updateRequest struct {
		Name          *string `json:"name"`
		Description   *string `json:"description"`
		PairingSystem *string `json:"pairing_system"`
		RoundsCount   *int    `json:"rounds_count"`
		TimeControl   *string `json:"time_control"`
		StartDate     *string `json:"start_date"`
		EndDate       *string `json:"end_date"`
		Status        *string `json:"status"`
	}
	var req updateRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	_, err = h.db.Exec(
		`UPDATE tournaments SET
			name = COALESCE($1, name),
			description = CASE WHEN $2::text IS NOT NULL THEN $2::text ELSE description END,
			pairing_system = COALESCE($3, pairing_system),
			rounds_count = CASE WHEN $4::int IS NOT NULL THEN $4::int ELSE rounds_count END,
			time_control = CASE WHEN $5::text IS NOT NULL THEN $5::text ELSE time_control END,
			status = COALESCE($6, status),
			updated_at = NOW()
		 WHERE id = $7`,
		req.Name, req.Description, req.PairingSystem, req.RoundsCount, req.TimeControl, req.Status, id,
	)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to update tournament")
	}

	var updated model.Tournament
	err = h.db.QueryRow(
		`SELECT id, name, description, organizer_id, pairing_system, status, rounds_count, time_control, start_date, end_date, settings, created_at, updated_at
		 FROM tournaments WHERE id = $1`,
		id,
	).Scan(
		&updated.ID, &updated.Name, &updated.Description, &updated.OrganizerID, &updated.PairingSystem, &updated.Status,
		&updated.RoundsCount, &updated.TimeControl, &updated.StartDate, &updated.EndDate, &updated.Settings,
		&updated.CreatedAt, &updated.UpdatedAt,
	)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "database error")
	}
	return c.JSON(http.StatusOK, &updated)
}

func (h *TournamentHandler) DeleteTournament(c echo.Context) error {
	userID := auth.GetUserID(c)
	id := c.Param("id")

	var organizerID string
	err := h.db.QueryRow(`SELECT organizer_id FROM tournaments WHERE id = $1`, id).Scan(&organizerID)
	if err == sql.ErrNoRows {
		return echo.NewHTTPError(http.StatusNotFound, "tournament not found")
	}
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "database error")
	}
	if organizerID != userID {
		return echo.NewHTTPError(http.StatusForbidden, "not the organizer")
	}

	_, err = h.db.Exec(`DELETE FROM tournaments WHERE id = $1`, id)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to delete tournament")
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *TournamentHandler) StartTournament(c echo.Context) error {
	userID := auth.GetUserID(c)
	id := c.Param("id")

	var t model.Tournament
	err := h.db.QueryRow(
		`SELECT id, name, description, organizer_id, pairing_system, status, rounds_count, time_control, start_date, end_date, settings, created_at, updated_at
		 FROM tournaments WHERE id = $1`,
		id,
	).Scan(
		&t.ID, &t.Name, &t.Description, &t.OrganizerID, &t.PairingSystem, &t.Status,
		&t.RoundsCount, &t.TimeControl, &t.StartDate, &t.EndDate, &t.Settings,
		&t.CreatedAt, &t.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return echo.NewHTTPError(http.StatusNotFound, "tournament not found")
	}
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "database error")
	}
	if t.OrganizerID != userID {
		return echo.NewHTTPError(http.StatusForbidden, "not the organizer")
	}
	if t.Status != "draft" && t.Status != "registration" {
		return echo.NewHTTPError(http.StatusBadRequest, "tournament cannot be started from current status")
	}

	players, err := h.getPlayers(id)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get players")
	}
	if len(players) < 2 {
		return echo.NewHTTPError(http.StatusBadRequest, "at least 2 players required to start tournament")
	}

	tx, err := h.db.Begin()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to begin transaction")
	}
	defer tx.Rollback()

	_, err = tx.Exec(`UPDATE tournaments SET status = 'active', updated_at = NOW() WHERE id = $1`, id)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to update tournament status")
	}

	roundID := uuid.New().String()
	_, err = tx.Exec(
		`INSERT INTO rounds (id, tournament_id, round_number, status, created_at) VALUES ($1, $2, 1, 'active', NOW())`,
		roundID, id,
	)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to create first round")
	}

	pairings := generatePairings(t.PairingSystem, players, 1)
	for _, p := range pairings {
		p.ID = uuid.New().String()
		p.RoundID = roundID
		_, err = tx.Exec(
			`INSERT INTO pairings (id, round_id, white_player_id, black_player_id, result, board_number, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())`,
			p.ID, p.RoundID, nullString(p.WhitePlayerID), nullString(p.BlackPlayerID),
			p.Result, p.BoardNumber,
		)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to create pairing")
		}
	}

	if err := tx.Commit(); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to commit transaction")
	}

	t.Status = "active"
	return c.JSON(http.StatusOK, &t)
}

func (h *TournamentHandler) ListPlayers(c echo.Context) error {
	id := c.Param("id")
	players, err := h.getPlayers(id)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "database error")
	}
	return c.JSON(http.StatusOK, players)
}

func (h *TournamentHandler) getPlayers(tournamentID string) ([]model.TournamentPlayer, error) {
	rows, err := h.db.Query(
		`SELECT tp.id, tp.tournament_id, tp.player_id, tp.seed, tp.rating, tp.score, tp.joined_at,
		        u.name, u.email, u.avatar_url
		 FROM tournament_players tp
		 JOIN users u ON u.id = tp.player_id
		 WHERE tp.tournament_id = $1
		 ORDER BY tp.seed NULLS LAST, tp.joined_at`,
		tournamentID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	players := make([]model.TournamentPlayer, 0)
	for rows.Next() {
		var p model.TournamentPlayer
		if err := rows.Scan(
			&p.ID, &p.TournamentID, &p.PlayerID, &p.Seed, &p.Rating, &p.Score, &p.JoinedAt,
			&p.Name, &p.Email, &p.AvatarURL,
		); err != nil {
			return nil, err
		}
		players = append(players, p)
	}
	return players, nil
}

func (h *TournamentHandler) AddPlayer(c echo.Context) error {
	userID := auth.GetUserID(c)
	tournamentID := c.Param("id")

	var organizerID string
	err := h.db.QueryRow(`SELECT organizer_id FROM tournaments WHERE id = $1`, tournamentID).Scan(&organizerID)
	if err == sql.ErrNoRows {
		return echo.NewHTTPError(http.StatusNotFound, "tournament not found")
	}
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "database error")
	}
	if organizerID != userID {
		return echo.NewHTTPError(http.StatusForbidden, "not the organizer")
	}

	type addPlayerRequest struct {
		PlayerID string `json:"player_id"`
		Name     string `json:"name"`
		Rating   *int   `json:"rating"`
		Seed     *int   `json:"seed"`
	}
	var req addPlayerRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	var playerID string
	if req.PlayerID != "" {
		// Verify user exists
		err := h.db.QueryRow(`SELECT id FROM users WHERE id = $1`, req.PlayerID).Scan(&playerID)
		if err == sql.ErrNoRows {
			return echo.NewHTTPError(http.StatusNotFound, "player not found")
		}
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "database error")
		}
	} else if req.Name != "" {
		// Create a guest user for the player
		guestID := uuid.New().String()
		_, err := h.db.Exec(
			`INSERT INTO users (id, name, created_at, updated_at) VALUES ($1, $2, NOW(), NOW())`,
			guestID, req.Name,
		)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to create guest player")
		}
		playerID = guestID
	} else {
		return echo.NewHTTPError(http.StatusBadRequest, "player_id or name is required")
	}

	tp := model.TournamentPlayer{
		ID:           uuid.New().String(),
		TournamentID: tournamentID,
		PlayerID:     playerID,
		Seed:         req.Seed,
		Rating:       req.Rating,
		Score:        0,
		JoinedAt:     time.Now(),
	}

	_, err = h.db.Exec(
		`INSERT INTO tournament_players (id, tournament_id, player_id, seed, rating, score, joined_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		tp.ID, tp.TournamentID, tp.PlayerID, tp.Seed, tp.Rating, tp.Score, tp.JoinedAt,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return echo.NewHTTPError(http.StatusConflict, "player already in tournament")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to add player")
	}

	// Fetch with joined user data
	err = h.db.QueryRow(
		`SELECT tp.id, tp.tournament_id, tp.player_id, tp.seed, tp.rating, tp.score, tp.joined_at,
		        u.name, u.email, u.avatar_url
		 FROM tournament_players tp
		 JOIN users u ON u.id = tp.player_id
		 WHERE tp.id = $1`,
		tp.ID,
	).Scan(
		&tp.ID, &tp.TournamentID, &tp.PlayerID, &tp.Seed, &tp.Rating, &tp.Score, &tp.JoinedAt,
		&tp.Name, &tp.Email, &tp.AvatarURL,
	)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to fetch player")
	}

	return c.JSON(http.StatusCreated, &tp)
}

func (h *TournamentHandler) RemovePlayer(c echo.Context) error {
	userID := auth.GetUserID(c)
	tournamentID := c.Param("id")
	playerEntryID := c.Param("playerId")

	var organizerID string
	err := h.db.QueryRow(`SELECT organizer_id FROM tournaments WHERE id = $1`, tournamentID).Scan(&organizerID)
	if err == sql.ErrNoRows {
		return echo.NewHTTPError(http.StatusNotFound, "tournament not found")
	}
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "database error")
	}
	if organizerID != userID {
		return echo.NewHTTPError(http.StatusForbidden, "not the organizer")
	}

	result, err := h.db.Exec(
		`DELETE FROM tournament_players WHERE id = $1 AND tournament_id = $2`,
		playerEntryID, tournamentID,
	)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to remove player")
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return echo.NewHTTPError(http.StatusNotFound, "player not found in tournament")
	}
	return c.NoContent(http.StatusNoContent)
}

func generatePairings(system string, players []model.TournamentPlayer, roundNum int) []model.Pairing {
	switch system {
	case "round_robin":
		return generateRoundRobinPairings(players, roundNum)
	case "swiss":
		return generateSwissPairings(players)
	case "knockout":
		return generateKnockoutPairings(players)
	default:
		return generateRoundRobinPairings(players, roundNum)
	}
}

func generateRoundRobinPairings(players []model.TournamentPlayer, roundNum int) []model.Pairing {
	n := len(players)
	if n%2 == 1 {
		players = append(players, model.TournamentPlayer{PlayerID: "bye"})
		n++
	}
	rotation := make([]model.TournamentPlayer, n-1)
	copy(rotation, players[1:])
	for i := 0; i < roundNum-1; i++ {
		last := rotation[len(rotation)-1]
		copy(rotation[1:], rotation[:len(rotation)-1])
		rotation[0] = last
	}
	table := append([]model.TournamentPlayer{players[0]}, rotation...)
	var pairings []model.Pairing
	for i := 0; i < n/2; i++ {
		white := table[i]
		black := table[n-1-i]
		if white.PlayerID == "bye" || black.PlayerID == "bye" {
			continue
		}
		board := i + 1
		p := model.Pairing{
			RoundID:     "",
			Result:      "pending",
			BoardNumber: &board,
		}
		if roundNum%2 == 0 {
			p.WhitePlayerID = sql.NullString{String: black.PlayerID, Valid: true}
			p.BlackPlayerID = sql.NullString{String: white.PlayerID, Valid: true}
		} else {
			p.WhitePlayerID = sql.NullString{String: white.PlayerID, Valid: true}
			p.BlackPlayerID = sql.NullString{String: black.PlayerID, Valid: true}
		}
		pairings = append(pairings, p)
	}
	return pairings
}

func generateSwissPairings(players []model.TournamentPlayer) []model.Pairing {
	// Group players by score
	scoreGroups := make(map[float64][]model.TournamentPlayer)
	for _, p := range players {
		scoreGroups[p.Score] = append(scoreGroups[p.Score], p)
	}
	// Collect unique scores and sort descending
	scores := make([]float64, 0, len(scoreGroups))
	for s := range scoreGroups {
		scores = append(scores, s)
	}
	for i := 0; i < len(scores)-1; i++ {
		for j := i + 1; j < len(scores); j++ {
			if scores[j] > scores[i] {
				scores[i], scores[j] = scores[j], scores[i]
			}
		}
	}
	// Shuffle within each score group and build sorted list
	var sorted []model.TournamentPlayer
	for _, s := range scores {
		group := scoreGroups[s]
		rand.Shuffle(len(group), func(i, j int) {
			group[i], group[j] = group[j], group[i]
		})
		sorted = append(sorted, group...)
	}

	var pairings []model.Pairing
	board := 1
	for i := 0; i+1 < len(sorted); i += 2 {
		b := board
		p := model.Pairing{
			WhitePlayerID: sql.NullString{String: sorted[i].PlayerID, Valid: true},
			BlackPlayerID: sql.NullString{String: sorted[i+1].PlayerID, Valid: true},
			Result:        "pending",
			BoardNumber:   &b,
		}
		pairings = append(pairings, p)
		board++
	}
	return pairings
}

func generateKnockoutPairings(players []model.TournamentPlayer) []model.Pairing {
	var pairings []model.Pairing
	board := 1
	for i := 0; i+1 < len(players); i += 2 {
		b := board
		p := model.Pairing{
			WhitePlayerID: sql.NullString{String: players[i].PlayerID, Valid: true},
			BlackPlayerID: sql.NullString{String: players[i+1].PlayerID, Valid: true},
			Result:        "pending",
			BoardNumber:   &b,
		}
		pairings = append(pairings, p)
		board++
	}
	return pairings
}

func nullString(ns sql.NullString) interface{} {
	if ns.Valid {
		return ns.String
	}
	return nil
}

func nullInt64(ni sql.NullInt64) interface{} {
	if ni.Valid {
		return ni.Int64
	}
	return nil
}

func nullTime(nt sql.NullTime) interface{} {
	if nt.Valid {
		return nt.Time
	}
	return nil
}

func isUniqueViolation(err error) bool {
	return err != nil && (containsStr(err.Error(), "unique") || containsStr(err.Error(), "duplicate"))
}

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
