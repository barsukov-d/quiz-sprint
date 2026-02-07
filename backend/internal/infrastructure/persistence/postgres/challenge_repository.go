package postgres

import (
	"database/sql"
	"errors"

	"github.com/barsukov/quiz-sprint/backend/internal/domain/quick_duel"
	"github.com/barsukov/quiz-sprint/backend/internal/domain/shared"
)

type ChallengeRepository struct {
	db *sql.DB
}

func NewChallengeRepository(db *sql.DB) *ChallengeRepository {
	return &ChallengeRepository{db: db}
}

func (r *ChallengeRepository) Save(challenge *quick_duel.DuelChallenge) error {
	query := `
		INSERT INTO duel_challenges (
			id, challenger_id, challenged_id, challenge_type, status,
			challenge_link, match_id, expires_at, created_at, responded_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (id) DO UPDATE SET
			challenged_id = EXCLUDED.challenged_id,
			status = EXCLUDED.status,
			match_id = EXCLUDED.match_id,
			responded_at = EXCLUDED.responded_at
	`

	var challengedID *string
	if challenge.ChallengedID() != nil {
		id := challenge.ChallengedID().String()
		challengedID = &id
	}

	var matchID *string
	if challenge.MatchID() != nil {
		id := challenge.MatchID().String()
		matchID = &id
	}

	var respondedAt *int64
	if challenge.RespondedAt() > 0 {
		ra := challenge.RespondedAt()
		respondedAt = &ra
	}

	_, err := r.db.Exec(query,
		challenge.ID().String(),
		challenge.ChallengerID().String(),
		challengedID,
		string(challenge.Type()),
		string(challenge.Status()),
		challenge.ChallengeLink(),
		matchID,
		challenge.ExpiresAt(),
		challenge.CreatedAt(),
		respondedAt,
	)

	return err
}

func (r *ChallengeRepository) FindByID(id quick_duel.ChallengeID) (*quick_duel.DuelChallenge, error) {
	query := `
		SELECT id, challenger_id, challenged_id, challenge_type, status,
			challenge_link, match_id, expires_at, created_at, responded_at
		FROM duel_challenges
		WHERE id = $1
	`

	return r.scanChallenge(r.db.QueryRow(query, id.String()))
}

func (r *ChallengeRepository) FindByLink(link string) (*quick_duel.DuelChallenge, error) {
	query := `
		SELECT id, challenger_id, challenged_id, challenge_type, status,
			challenge_link, match_id, expires_at, created_at, responded_at
		FROM duel_challenges
		WHERE challenge_link = $1
	`

	return r.scanChallenge(r.db.QueryRow(query, link))
}

func (r *ChallengeRepository) FindPendingForPlayer(playerID quick_duel.UserID) ([]*quick_duel.DuelChallenge, error) {
	query := `
		SELECT id, challenger_id, challenged_id, challenge_type, status,
			challenge_link, match_id, expires_at, created_at, responded_at
		FROM duel_challenges
		WHERE challenged_id = $1 AND status = 'pending'
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query, playerID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanChallenges(rows)
}

func (r *ChallengeRepository) FindPendingByChallenger(playerID quick_duel.UserID) ([]*quick_duel.DuelChallenge, error) {
	query := `
		SELECT id, challenger_id, challenged_id, challenge_type, status,
			challenge_link, match_id, expires_at, created_at, responded_at
		FROM duel_challenges
		WHERE challenger_id = $1 AND status = 'pending'
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query, playerID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanChallenges(rows)
}

func (r *ChallengeRepository) Delete(id quick_duel.ChallengeID) error {
	query := `DELETE FROM duel_challenges WHERE id = $1`
	_, err := r.db.Exec(query, id.String())
	return err
}

func (r *ChallengeRepository) DeleteExpired(currentTime int64) error {
	query := `
		UPDATE duel_challenges
		SET status = 'expired', responded_at = $1
		WHERE status = 'pending' AND expires_at <= $1
	`
	_, err := r.db.Exec(query, currentTime)
	return err
}

func (r *ChallengeRepository) scanChallenge(row *sql.Row) (*quick_duel.DuelChallenge, error) {
	var (
		id            string
		challengerID  string
		challengedID  sql.NullString
		challengeType string
		status        string
		challengeLink sql.NullString
		matchID       sql.NullString
		expiresAt     int64
		createdAt     int64
		respondedAt   sql.NullInt64
	)

	err := row.Scan(
		&id, &challengerID, &challengedID, &challengeType, &status,
		&challengeLink, &matchID, &expiresAt, &createdAt, &respondedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, quick_duel.ErrChallengeNotFound
	}
	if err != nil {
		return nil, err
	}

	return r.reconstructChallenge(
		id, challengerID, challengedID, challengeType, status,
		challengeLink, matchID, expiresAt, createdAt, respondedAt,
	)
}

func (r *ChallengeRepository) scanChallenges(rows *sql.Rows) ([]*quick_duel.DuelChallenge, error) {
	var challenges []*quick_duel.DuelChallenge

	for rows.Next() {
		var (
			id            string
			challengerID  string
			challengedID  sql.NullString
			challengeType string
			status        string
			challengeLink sql.NullString
			matchID       sql.NullString
			expiresAt     int64
			createdAt     int64
			respondedAt   sql.NullInt64
		)

		err := rows.Scan(
			&id, &challengerID, &challengedID, &challengeType, &status,
			&challengeLink, &matchID, &expiresAt, &createdAt, &respondedAt,
		)
		if err != nil {
			return nil, err
		}

		challenge, err := r.reconstructChallenge(
			id, challengerID, challengedID, challengeType, status,
			challengeLink, matchID, expiresAt, createdAt, respondedAt,
		)
		if err != nil {
			return nil, err
		}

		challenges = append(challenges, challenge)
	}

	return challenges, nil
}

func (r *ChallengeRepository) reconstructChallenge(
	id, challengerID string,
	challengedID sql.NullString,
	challengeType, status string,
	challengeLink, matchID sql.NullString,
	expiresAt, createdAt int64,
	respondedAt sql.NullInt64,
) (*quick_duel.DuelChallenge, error) {
	cid := quick_duel.NewChallengeIDFromString(id)
	challengerUID, _ := shared.NewUserID(challengerID)

	var challengedUID *quick_duel.UserID
	if challengedID.Valid {
		uid, _ := shared.NewUserID(challengedID.String)
		challengedUID = &uid
	}

	var mid *quick_duel.GameID
	if matchID.Valid {
		gid := quick_duel.NewGameIDFromString(matchID.String)
		mid = &gid
	}

	var ra int64
	if respondedAt.Valid {
		ra = respondedAt.Int64
	}

	link := ""
	if challengeLink.Valid {
		link = challengeLink.String
	}

	return quick_duel.ReconstructDuelChallenge(
		cid,
		challengerUID,
		challengedUID,
		quick_duel.ChallengeType(challengeType),
		quick_duel.ChallengeStatus(status),
		link,
		expiresAt,
		createdAt,
		ra,
		mid,
	), nil
}
