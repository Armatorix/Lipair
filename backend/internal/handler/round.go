package handler

import (
	"database/sql"
	"net/http"

	"github.com/Armatorix/ChessMgr/backend/internal/auth"
	"github.com/Armatorix/ChessMgr/backend/internal/db"
	"github.com/Armatorix/ChessMgr/backend/internal/model"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type RoundHandler struct {
	db *db.DB
}

func NewRoundHandler(db *db.DB) *RoundHandler {
	return &RoundHandler{db: db}
}

func (h *RoundHandler) ListRounds(c echo.Context) error {
	tournamentID := c.Param("id")

	rows, err := h.db.Query(
		`SELECT id, tournament_id, round_number, status, created_at FROM rounds WHERE tournament_id = $1 ORDER BY round_number`,
		tournamentID,
	)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "database error")
	}
	defer rows.Close()

	rounds := make([]model.Round, 0)
	for rows.Next() {
		var r model.Round
		if err := rows.Scan(&r.ID, &r.TournamentID, &r.RoundNumber, &r.Status, &r.CreatedAt); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "scan error")
		}
		rounds = append(rounds, r)
	}
	return c.JSON(http.StatusOK, rounds)
}

func (h *RoundHandler) CreateRound(c echo.Context) error {
	userID := auth.GetUserID(c)
	tournamentID := c.Param("id")

	var t struct {
		OrganizerID   string `db:"organizer_id"`
		PairingSystem string `db:"pairing_system"`
		Status        string `db:"status"`
	}
	err := h.db.QueryRow(
		`SELECT organizer_id, pairing_system, status FROM tournaments WHERE id = $1`, tournamentID,
	).Scan(&t.OrganizerID, &t.PairingSystem, &t.Status)
	if err == sql.ErrNoRows {
		return echo.NewHTTPError(http.StatusNotFound, "tournament not found")
	}
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "database error")
	}
	if t.OrganizerID != userID {
		return echo.NewHTTPError(http.StatusForbidden, "not the organizer")
	}
	if t.Status != "active" {
		return echo.NewHTTPError(http.StatusBadRequest, "tournament is not active")
	}

	var maxRound int
	err = h.db.QueryRow(
		`SELECT COALESCE(MAX(round_number), 0) FROM rounds WHERE tournament_id = $1`, tournamentID,
	).Scan(&maxRound)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "database error")
	}

	players, err := h.getTournamentPlayers(tournamentID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get players")
	}
	if len(players) < 2 {
		return echo.NewHTTPError(http.StatusBadRequest, "not enough players")
	}

	newRoundNum := maxRound + 1
	roundID := uuid.New().String()

	tx, err := h.db.Begin()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to begin transaction")
	}
	defer tx.Rollback()

	_, err = tx.Exec(
		`INSERT INTO rounds (id, tournament_id, round_number, status, created_at) VALUES ($1, $2, $3, 'active', NOW())`,
		roundID, tournamentID, newRoundNum,
	)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to create round")
	}

	pairings := generatePairings(t.PairingSystem, players, newRoundNum)
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

	round := model.Round{
		ID:           roundID,
		TournamentID: tournamentID,
		RoundNumber:  newRoundNum,
		Status:       "active",
	}
	return c.JSON(http.StatusCreated, &round)
}

func (h *RoundHandler) GetRound(c echo.Context) error {
	roundID := c.Param("roundId")
	var r model.Round
	err := h.db.QueryRow(
		`SELECT id, tournament_id, round_number, status, created_at FROM rounds WHERE id = $1`, roundID,
	).Scan(&r.ID, &r.TournamentID, &r.RoundNumber, &r.Status, &r.CreatedAt)
	if err == sql.ErrNoRows {
		return echo.NewHTTPError(http.StatusNotFound, "round not found")
	}
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "database error")
	}
	return c.JSON(http.StatusOK, &r)
}

func (h *RoundHandler) ListPairings(c echo.Context) error {
	roundID := c.Param("roundId")

	rows, err := h.db.Query(
		`SELECT p.id, p.round_id, p.white_player_id, p.black_player_id, p.result, p.board_number, p.created_at, p.updated_at,
		        COALESCE(wu.name, '') as white_player_name, COALESCE(bu.name, '') as black_player_name
		 FROM pairings p
		 LEFT JOIN users wu ON wu.id = p.white_player_id
		 LEFT JOIN users bu ON bu.id = p.black_player_id
		 WHERE p.round_id = $1
		 ORDER BY p.board_number NULLS LAST`,
		roundID,
	)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "database error")
	}
	defer rows.Close()

	pairings := make([]model.Pairing, 0)
	for rows.Next() {
		var p model.Pairing
		if err := rows.Scan(
			&p.ID, &p.RoundID, &p.WhitePlayerID, &p.BlackPlayerID,
			&p.Result, &p.BoardNumber, &p.CreatedAt, &p.UpdatedAt,
			&p.WhitePlayerName, &p.BlackPlayerName,
		); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "scan error")
		}
		pairings = append(pairings, p)
	}
	return c.JSON(http.StatusOK, pairings)
}

func (h *RoundHandler) UpdatePairing(c echo.Context) error {
	userID := auth.GetUserID(c)
	tournamentID := c.Param("id")
	pairingID := c.Param("pairingId")

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

	type updateRequest struct {
		Result string `json:"result"`
	}
	var req updateRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	validResults := map[string]bool{
		"white_wins": true, "black_wins": true, "draw": true, "bye": true, "pending": true,
	}
	if !validResults[req.Result] {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid result")
	}

	// Get current pairing
	var p model.Pairing
	err = h.db.QueryRow(
		`SELECT id, round_id, white_player_id, black_player_id, result, board_number, created_at, updated_at FROM pairings WHERE id = $1`,
		pairingID,
	).Scan(&p.ID, &p.RoundID, &p.WhitePlayerID, &p.BlackPlayerID, &p.Result, &p.BoardNumber, &p.CreatedAt, &p.UpdatedAt)
	if err == sql.ErrNoRows {
		return echo.NewHTTPError(http.StatusNotFound, "pairing not found")
	}
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "database error")
	}

	tx, err := h.db.Begin()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to begin transaction")
	}
	defer tx.Rollback()

	// Revert previous score if applicable
	if p.Result != "pending" {
		if err := revertScores(tx, p); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to revert scores")
		}
	}

	// Apply new result
	_, err = tx.Exec(
		`UPDATE pairings SET result = $1, updated_at = NOW() WHERE id = $2`,
		req.Result, pairingID,
	)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to update pairing")
	}

	p.Result = req.Result
	if req.Result != "pending" {
		if err := applyScores(tx, p, tournamentID); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to apply scores")
		}
	}

	if err := tx.Commit(); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to commit")
	}

	// Fetch updated pairing with player names
	err = h.db.QueryRow(
		`SELECT p.id, p.round_id, p.white_player_id, p.black_player_id, p.result, p.board_number, p.created_at, p.updated_at,
		        COALESCE(wu.name, '') as white_player_name, COALESCE(bu.name, '') as black_player_name
		 FROM pairings p
		 LEFT JOIN users wu ON wu.id = p.white_player_id
		 LEFT JOIN users bu ON bu.id = p.black_player_id
		 WHERE p.id = $1`,
		pairingID,
	).Scan(
		&p.ID, &p.RoundID, &p.WhitePlayerID, &p.BlackPlayerID,
		&p.Result, &p.BoardNumber, &p.CreatedAt, &p.UpdatedAt,
		&p.WhitePlayerName, &p.BlackPlayerName,
	)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "database error")
	}

	return c.JSON(http.StatusOK, &p)
}

func revertScores(tx *sql.Tx, p model.Pairing) error {
	switch p.Result {
	case "white_wins":
		if p.WhitePlayerID.Valid {
			if _, err := tx.Exec(`UPDATE tournament_players SET score = score - 1 WHERE player_id = $1`, p.WhitePlayerID.String); err != nil {
				return err
			}
		}
	case "black_wins":
		if p.BlackPlayerID.Valid {
			if _, err := tx.Exec(`UPDATE tournament_players SET score = score - 1 WHERE player_id = $1`, p.BlackPlayerID.String); err != nil {
				return err
			}
		}
	case "draw":
		if p.WhitePlayerID.Valid {
			if _, err := tx.Exec(`UPDATE tournament_players SET score = score - 0.5 WHERE player_id = $1`, p.WhitePlayerID.String); err != nil {
				return err
			}
		}
		if p.BlackPlayerID.Valid {
			if _, err := tx.Exec(`UPDATE tournament_players SET score = score - 0.5 WHERE player_id = $1`, p.BlackPlayerID.String); err != nil {
				return err
			}
		}
	case "bye":
		if p.WhitePlayerID.Valid {
			if _, err := tx.Exec(`UPDATE tournament_players SET score = score - 1 WHERE player_id = $1`, p.WhitePlayerID.String); err != nil {
				return err
			}
		}
	}
	return nil
}

func applyScores(tx *sql.Tx, p model.Pairing, tournamentID string) error {
	switch p.Result {
	case "white_wins":
		if p.WhitePlayerID.Valid {
			if _, err := tx.Exec(
				`UPDATE tournament_players SET score = score + 1 WHERE player_id = $1 AND tournament_id = $2`,
				p.WhitePlayerID.String, tournamentID,
			); err != nil {
				return err
			}
		}
	case "black_wins":
		if p.BlackPlayerID.Valid {
			if _, err := tx.Exec(
				`UPDATE tournament_players SET score = score + 1 WHERE player_id = $1 AND tournament_id = $2`,
				p.BlackPlayerID.String, tournamentID,
			); err != nil {
				return err
			}
		}
	case "draw":
		if p.WhitePlayerID.Valid {
			if _, err := tx.Exec(
				`UPDATE tournament_players SET score = score + 0.5 WHERE player_id = $1 AND tournament_id = $2`,
				p.WhitePlayerID.String, tournamentID,
			); err != nil {
				return err
			}
		}
		if p.BlackPlayerID.Valid {
			if _, err := tx.Exec(
				`UPDATE tournament_players SET score = score + 0.5 WHERE player_id = $1 AND tournament_id = $2`,
				p.BlackPlayerID.String, tournamentID,
			); err != nil {
				return err
			}
		}
	case "bye":
		if p.WhitePlayerID.Valid {
			if _, err := tx.Exec(
				`UPDATE tournament_players SET score = score + 1 WHERE player_id = $1 AND tournament_id = $2`,
				p.WhitePlayerID.String, tournamentID,
			); err != nil {
				return err
			}
		}
	}
	return nil
}

func (h *RoundHandler) getTournamentPlayers(tournamentID string) ([]model.TournamentPlayer, error) {
	rows, err := h.db.Query(
		`SELECT tp.id, tp.tournament_id, tp.player_id, tp.seed, tp.rating, tp.score, tp.joined_at,
		        u.name, u.email, u.avatar_url
		 FROM tournament_players tp
		 JOIN users u ON u.id = tp.player_id
		 WHERE tp.tournament_id = $1
		 ORDER BY tp.score DESC, tp.seed NULLS LAST`,
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
