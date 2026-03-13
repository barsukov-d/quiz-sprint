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
			challenge_link, link_code, match_id, expires_at, created_at, responded_at,
			telegram_message_id
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		ON CONFLICT (id) DO UPDATE SET
			challenged_id = EXCLUDED.challenged_id,
			status = EXCLUDED.status,
			match_id = EXCLUDED.match_id,
			responded_at = EXCLUDED.responded_at,
			telegram_message_id = EXCLUDED.telegram_message_id
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

	var linkCode *string
	if challenge.LinkCode() != "" {
		lc := challenge.LinkCode()
		linkCode = &lc
	}

	_, err := r.db.Exec(query,
		challenge.ID().String(),
		challenge.ChallengerID().String(),
		challengedID,
		string(challenge.Type()),
		string(challenge.Status()),
		challenge.ChallengeLink(),
		linkCode,
		matchID,
		challenge.ExpiresAt(),
		challenge.CreatedAt(),
		respondedAt,
		challenge.TelegramMessageID(),
	)

	return err
}

func (r *ChallengeRepository) FindByID(id quick_duel.ChallengeID) (*quick_duel.DuelChallenge, error) {
	query := `
		SELECT id, challenger_id, challenged_id, challenge_type, status,
			challenge_link, link_code, match_id, expires_at, created_at, responded_at,
			telegram_message_id
		FROM duel_challenges
		WHERE id = $1
	`

	return r.scanChallenge(r.db.QueryRow(query, id.String()))
}

func (r *ChallengeRepository) FindByLink(link string) (*quick_duel.DuelChallenge, error) {
	query := `
		SELECT id, challenger_id, challenged_id, challenge_type, status,
			challenge_link, link_code, match_id, expires_at, created_at, responded_at,
			telegram_message_id
		FROM duel_challenges
		WHERE challenge_link = $1
	`

	return r.scanChallenge(r.db.QueryRow(query, link))
}

func (r *ChallengeRepository) FindByLinkCode(code string) (*quick_duel.DuelChallenge, error) {
	query := `
		SELECT id, challenger_id, challenged_id, challenge_type, status,
			challenge_link, link_code, match_id, expires_at, created_at, responded_at,
			telegram_message_id
		FROM duel_challenges
		WHERE link_code = $1
	`

	return r.scanChallenge(r.db.QueryRow(query, code))
}

func (r *ChallengeRepository) FindPendingForPlayer(playerID quick_duel.UserID) ([]*quick_duel.DuelChallenge, error) {
	query := `
		SELECT id, challenger_id, challenged_id, challenge_type, status,
			challenge_link, link_code, match_id, expires_at, created_at, responded_at,
			telegram_message_id
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
			challenge_link, link_code, match_id, expires_at, created_at, responded_at,
			telegram_message_id
		FROM duel_challenges
		WHERE challenger_id = $1 AND status IN ('pending', 'accepted_waiting_inviter')
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query, playerID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanChallenges(rows)
}

func (r *ChallengeRepository) FindAcceptedWaitingForPlayer(playerID quick_duel.UserID) ([]*quick_duel.DuelChallenge, error) {
	query := `
		SELECT id, challenger_id, challenged_id, challenge_type, status,
			challenge_link, link_code, match_id, expires_at, created_at, responded_at,
			telegram_message_id
		FROM duel_challenges
		WHERE challenged_id = $1 AND status = 'accepted_waiting_inviter'
		ORDER BY responded_at DESC
	`
	rows, err := r.db.Query(query, playerID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return r.scanChallenges(rows)
}

// FindPendingExpiredWithMessageID returns pending challenges that have expired
// AND have a telegram_message_id set (need their message edited before bulk expire).
func (r *ChallengeRepository) FindPendingExpiredWithMessageID(currentTime int64) ([]*quick_duel.DuelChallenge, error) {
	query := `
		SELECT id, challenger_id, challenged_id, challenge_type, status,
			challenge_link, link_code, match_id, expires_at, created_at, responded_at,
			telegram_message_id
		FROM duel_challenges
		WHERE status = 'pending'
		  AND expires_at <= $1
		  AND telegram_message_id IS NOT NULL
		  AND telegram_message_id > 0
	`
	rows, err := r.db.Query(query, currentTime)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return r.scanChallenges(rows)
}

// FindExpiredForPlayer returns expired challenges visible to the player (as inviter or invitee).
func (r *ChallengeRepository) FindExpiredForPlayer(playerID quick_duel.UserID) ([]*quick_duel.DuelChallenge, error) {
	query := `
		SELECT id, challenger_id, challenged_id, challenge_type, status,
			challenge_link, link_code, match_id, expires_at, created_at, responded_at,
			telegram_message_id
		FROM duel_challenges
		WHERE status = 'expired'
		  AND (challenger_id = $1 OR challenged_id = $1)
		ORDER BY responded_at DESC
		LIMIT 20
	`
	rows, err := r.db.Query(query, playerID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return r.scanChallenges(rows)
}

// DeleteHardExpired removes expired/declined challenges older than olderThan unix timestamp.
func (r *ChallengeRepository) DeleteHardExpired(olderThan int64) error {
	query := `
		DELETE FROM duel_challenges
		WHERE status IN ('expired', 'declined')
		  AND responded_at IS NOT NULL
		  AND responded_at < $1
	`
	_, err := r.db.Exec(query, olderThan)
	return err
}

func (r *ChallengeRepository) FindPendingExpired(currentTime int64) ([]*quick_duel.DuelChallenge, error) {
	query := `
		SELECT id, challenger_id, challenged_id, challenge_type, status,
			challenge_link, link_code, match_id, expires_at, created_at, responded_at,
			telegram_message_id
		FROM duel_challenges
		WHERE status = 'pending' AND expires_at <= $1
		ORDER BY created_at ASC
		LIMIT 50
	`
	rows, err := r.db.Query(query, currentTime)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return r.scanChallenges(rows)
}

func (r *ChallengeRepository) FindWaitingExpired(currentTime int64) ([]*quick_duel.DuelChallenge, error) {
	query := `
		SELECT id, challenger_id, challenged_id, challenge_type, status,
			challenge_link, link_code, match_id, expires_at, created_at, responded_at,
			telegram_message_id
		FROM duel_challenges
		WHERE status = 'accepted_waiting_inviter'
		AND responded_at + 1800 <= $1
		ORDER BY responded_at ASC
		LIMIT 50
	`
	rows, err := r.db.Query(query, currentTime)
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
		WHERE (status = 'pending' AND expires_at <= $1)
		   OR (status = 'accepted_waiting_inviter' AND responded_at + 1800 <= $1)
	`
	_, err := r.db.Exec(query, currentTime)
	return err
}

func (r *ChallengeRepository) scanChallenge(row *sql.Row) (*quick_duel.DuelChallenge, error) {
	var (
		id                string
		challengerID      string
		challengedID      sql.NullString
		challengeType     string
		status            string
		challengeLink     sql.NullString
		linkCode          sql.NullString
		matchID           sql.NullString
		expiresAt         int64
		createdAt         int64
		respondedAt       sql.NullInt64
		telegramMessageID sql.NullInt64
	)

	err := row.Scan(
		&id, &challengerID, &challengedID, &challengeType, &status,
		&challengeLink, &linkCode, &matchID, &expiresAt, &createdAt, &respondedAt,
		&telegramMessageID,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, quick_duel.ErrChallengeNotFound
	}
	if err != nil {
		return nil, err
	}

	return r.reconstructChallenge(
		id, challengerID, challengedID, challengeType, status,
		challengeLink, linkCode, matchID, expiresAt, createdAt, respondedAt,
		telegramMessageID,
	)
}

func (r *ChallengeRepository) scanChallenges(rows *sql.Rows) ([]*quick_duel.DuelChallenge, error) {
	var challenges []*quick_duel.DuelChallenge

	for rows.Next() {
		var (
			id                string
			challengerID      string
			challengedID      sql.NullString
			challengeType     string
			status            string
			challengeLink     sql.NullString
			linkCode          sql.NullString
			matchID           sql.NullString
			expiresAt         int64
			createdAt         int64
			respondedAt       sql.NullInt64
			telegramMessageID sql.NullInt64
		)

		err := rows.Scan(
			&id, &challengerID, &challengedID, &challengeType, &status,
			&challengeLink, &linkCode, &matchID, &expiresAt, &createdAt, &respondedAt,
			&telegramMessageID,
		)
		if err != nil {
			return nil, err
		}

		challenge, err := r.reconstructChallenge(
			id, challengerID, challengedID, challengeType, status,
			challengeLink, linkCode, matchID, expiresAt, createdAt, respondedAt,
			telegramMessageID,
		)
		if err != nil {
			return nil, err
		}

		challenges = append(challenges, challenge)
	}

	return challenges, nil
}

func (r *ChallengeRepository) FindByIDForUpdate(tx *sql.Tx, id quick_duel.ChallengeID) (*quick_duel.DuelChallenge, error) {
	query := `
		SELECT id, challenger_id, challenged_id, challenge_type, status,
			challenge_link, link_code, match_id, expires_at, created_at, responded_at,
			telegram_message_id
		FROM duel_challenges
		WHERE id = $1
		FOR UPDATE
	`

	return r.scanChallenge(tx.QueryRow(query, id.String()))
}

func (r *ChallengeRepository) FindByLinkCodeForUpdate(tx *sql.Tx, code string) (*quick_duel.DuelChallenge, error) {
	query := `
		SELECT id, challenger_id, challenged_id, challenge_type, status,
			challenge_link, link_code, match_id, expires_at, created_at, responded_at,
			telegram_message_id
		FROM duel_challenges
		WHERE link_code = $1 AND status = 'pending'
		FOR UPDATE
		LIMIT 1
	`

	return r.scanChallenge(tx.QueryRow(query, code))
}

func (r *ChallengeRepository) SaveInTx(tx *sql.Tx, challenge *quick_duel.DuelChallenge) error {
	query := `
		INSERT INTO duel_challenges (
			id, challenger_id, challenged_id, challenge_type, status,
			challenge_link, link_code, match_id, expires_at, created_at, responded_at,
			telegram_message_id
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		ON CONFLICT (id) DO UPDATE SET
			challenged_id = EXCLUDED.challenged_id,
			status = EXCLUDED.status,
			match_id = EXCLUDED.match_id,
			responded_at = EXCLUDED.responded_at,
			telegram_message_id = EXCLUDED.telegram_message_id
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

	var linkCode *string
	if challenge.LinkCode() != "" {
		lc := challenge.LinkCode()
		linkCode = &lc
	}

	_, err := tx.Exec(query,
		challenge.ID().String(),
		challenge.ChallengerID().String(),
		challengedID,
		string(challenge.Type()),
		string(challenge.Status()),
		challenge.ChallengeLink(),
		linkCode,
		matchID,
		challenge.ExpiresAt(),
		challenge.CreatedAt(),
		respondedAt,
		challenge.TelegramMessageID(),
	)

	return err
}

func (r *ChallengeRepository) reconstructChallenge(
	id, challengerID string,
	challengedID sql.NullString,
	challengeType, status string,
	challengeLink, linkCode, matchID sql.NullString,
	expiresAt, createdAt int64,
	respondedAt sql.NullInt64,
	telegramMessageID sql.NullInt64,
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

	lc := ""
	if linkCode.Valid {
		lc = linkCode.String
	}

	var tgMsgID int64
	if telegramMessageID.Valid {
		tgMsgID = telegramMessageID.Int64
	}

	return quick_duel.ReconstructDuelChallenge(
		cid,
		challengerUID,
		challengedUID,
		quick_duel.ChallengeType(challengeType),
		quick_duel.ChallengeStatus(status),
		link,
		lc,
		expiresAt,
		createdAt,
		ra,
		mid,
		"",
		tgMsgID,
	), nil
}
