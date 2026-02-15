package quick_duel

// ========================================
// Common DTOs
// ========================================

// DuelPlayerDTO represents a player in a duel
type DuelPlayerDTO struct {
	ID         string  `json:"id"`
	Username   string  `json:"username"`
	Avatar     string  `json:"avatar,omitempty"`
	MMR        int     `json:"mmr"`
	League     string  `json:"league"`
	Division   int     `json:"division"`
	LeagueIcon string  `json:"leagueIcon"`
	Score      int     `json:"score"`
	Connected  bool    `json:"connected"`
}

// DuelGameDTO represents a duel game
type DuelGameDTO struct {
	ID           string          `json:"id"`
	Status       string          `json:"status"`
	Player1      DuelPlayerDTO   `json:"player1"`
	Player2      DuelPlayerDTO   `json:"player2"`
	CurrentRound int             `json:"currentRound"`
	TotalRounds  int             `json:"totalRounds"`
	StartedAt    int64           `json:"startedAt"`
	FinishedAt   int64           `json:"finishedAt,omitempty"`
	WinnerID     *string         `json:"winnerId,omitempty"`
	IsFriendGame bool            `json:"isFriendGame"`
}

// DuelQuestionDTO represents a question in a duel (no IsCorrect field!)
type DuelQuestionDTO struct {
	ID           string            `json:"id"`
	QuestionNum  int               `json:"questionNumber"`
	Text         string            `json:"text"`
	Answers      []DuelAnswerDTO   `json:"answers"`
	TimeLimit    int               `json:"timeLimit"`
	ServerTime   int64             `json:"serverTime"`
}

// DuelAnswerDTO represents an answer option
type DuelAnswerDTO struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}

// PlayerRatingDTO represents a player's competitive ranking
type PlayerRatingDTO struct {
	PlayerID     string  `json:"playerId"`
	MMR          int     `json:"mmr"`
	League       string  `json:"league"`
	Division     int     `json:"division"`
	LeagueLabel  string  `json:"leagueLabel"`
	LeagueIcon   string  `json:"leagueIcon"`
	PeakMMR      int     `json:"peakMmr"`
	PeakLeague   string  `json:"peakLeague"`
	SeasonWins   int     `json:"seasonWins"`
	SeasonLosses int     `json:"seasonLosses"`
	WinRate      float64 `json:"winRate"`
	GamesAtRank  int     `json:"gamesAtRank"`
	CanDemote    bool    `json:"canDemote"`
}

// LeaderboardEntryDTO represents a leaderboard entry
type LeaderboardEntryDTO struct {
	Rank       int     `json:"rank"`
	PlayerID   string  `json:"playerId"`
	Username   string  `json:"username"`
	Avatar     string  `json:"avatar,omitempty"`
	MMR        int     `json:"mmr"`
	League     string  `json:"league"`
	LeagueIcon string  `json:"leagueIcon"`
	Wins       int     `json:"wins"`
	Losses     int     `json:"losses"`
	WinRate    float64 `json:"winRate"`
}

// ChallengeDTO represents a duel challenge
type ChallengeDTO struct {
	ID            string  `json:"id"`
	ChallengerID  string  `json:"challengerId"`
	ChallengedID  *string `json:"challengedId,omitempty"`
	Type          string  `json:"type"`
	Status        string  `json:"status"`
	ChallengeLink string  `json:"challengeLink,omitempty"`
	ExpiresAt     int64   `json:"expiresAt"`
	ExpiresIn     int     `json:"expiresIn"`
	CreatedAt     int64   `json:"createdAt"`
}

// GameHistoryEntryDTO represents a game in history
type GameHistoryEntryDTO struct {
	GameID        string `json:"gameId"`
	Opponent      string `json:"opponent"`
	OpponentMMR   int    `json:"opponentMmr"`
	Result        string `json:"result"`
	PlayerScore   int    `json:"playerScore"`
	OpponentScore int    `json:"opponentScore"`
	MMRChange     int    `json:"mmrChange"`
	IsFriendGame  bool   `json:"isFriendGame"`
	CompletedAt   int64  `json:"completedAt"`
}

// ReferralDTO represents a referral
type ReferralDTO struct {
	ID                string   `json:"id"`
	InviteeID         string   `json:"inviteeId"`
	InviteeUsername   string   `json:"inviteeUsername"`
	MilestonesReached []string `json:"milestonesReached"`
	PendingRewards    []string `json:"pendingRewards"`
	CreatedAt         int64    `json:"createdAt"`
}

// ReferralRewardDTO represents a referral reward
type ReferralRewardDTO struct {
	Tickets int    `json:"tickets"`
	Coins   int    `json:"coins"`
	Badge   string `json:"badge,omitempty"`
	Avatar  string `json:"avatar,omitempty"`
	Title   string `json:"title,omitempty"`
}

// FriendDTO represents an online friend
type FriendDTO struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Avatar   string `json:"avatar,omitempty"`
	MMR      int    `json:"mmr"`
	League   string `json:"league"`
	IsOnline bool   `json:"isOnline"`
	InGame   bool   `json:"inGame"`
}

// ========================================
// GetDuelStatus Use Case
// ========================================

type GetDuelStatusInput struct {
	PlayerID string `json:"playerId"`
}

type GetDuelStatusOutput struct {
	HasActiveDuel     bool              `json:"hasActiveDuel"`
	ActiveGameID      *string           `json:"activeGameId,omitempty"`
	Player            PlayerRatingDTO   `json:"player"`
	Tickets           int               `json:"tickets"`
	FriendsOnline     []FriendDTO       `json:"friendsOnline"`
	PendingChallenges []ChallengeDTO    `json:"pendingChallenges"`
	SeasonID          string            `json:"seasonId"`
	SeasonEndsAt      int64             `json:"seasonEndsAt"`
}

// ========================================
// JoinQueue Use Case
// ========================================

type JoinQueueInput struct {
	PlayerID string `json:"playerId"`
}

type JoinQueueOutput struct {
	QueueID       string `json:"queueId"`
	Status        string `json:"status"`
	EstimatedWait int    `json:"estimatedWait"`
	MMRRange      string `json:"mmrRange"`
	Position      int    `json:"position"`
}

// ========================================
// LeaveQueue Use Case
// ========================================

type LeaveQueueInput struct {
	PlayerID string `json:"playerId"`
}

type LeaveQueueOutput struct {
	Success        bool `json:"success"`
	TicketRefunded bool `json:"ticketRefunded"`
	NewTicketCount int  `json:"newTicketCount"`
}

// ========================================
// SendChallenge Use Case
// ========================================

type SendChallengeInput struct {
	PlayerID string `json:"playerId"`
	FriendID string `json:"friendId"`
}

type SendChallengeOutput struct {
	ChallengeID    string `json:"challengeId"`
	Status         string `json:"status"`
	ExpiresIn      int    `json:"expiresIn"`
	TicketConsumed bool   `json:"ticketConsumed"`
}

// ========================================
// RespondChallenge Use Case
// ========================================

type RespondChallengeInput struct {
	PlayerID    string `json:"playerId"`
	ChallengeID string `json:"challengeId"`
	Action      string `json:"action"` // "accept" or "decline"
}

type RespondChallengeOutput struct {
	Success        bool    `json:"success"`
	GameID         *string `json:"gameId,omitempty"`
	TicketConsumed bool    `json:"ticketConsumed"`
	StartsIn       *int    `json:"startsIn,omitempty"`
}

// ========================================
// AcceptByLinkCode Use Case
// ========================================

type AcceptByLinkCodeInput struct {
	PlayerID string `json:"playerId"`
	LinkCode string `json:"linkCode"` // e.g., "duel_abc12345"
}

type AcceptByLinkCodeOutput struct {
	Success        bool    `json:"success"`
	GameID         *string `json:"gameId,omitempty"`
	TicketConsumed bool    `json:"ticketConsumed"`
	StartsIn       *int    `json:"startsIn,omitempty"`
	ChallengerID   string  `json:"challengerId"`
}

// ========================================
// CreateChallengeLink Use Case
// ========================================

type CreateChallengeLinkInput struct {
	PlayerID string `json:"playerId"`
}

type CreateChallengeLinkOutput struct {
	ChallengeLink string `json:"challengeLink"`
	ExpiresAt     int64  `json:"expiresAt"`
	ShareText     string `json:"shareText"`
}

// ========================================
// GetGameResult Use Case
// ========================================

type GetGameResultInput struct {
	PlayerID string `json:"playerId"`
	GameID   string `json:"gameId"`
}

type GameQuestionResultDTO struct {
	QuestionNum       int    `json:"questionNumber"`
	QuestionText      string `json:"questionText"`
	PlayerAnswerID    string `json:"playerAnswerId"`
	PlayerCorrect     bool   `json:"playerCorrect"`
	PlayerTime        int64  `json:"playerTime"`
	OpponentAnswerID  string `json:"opponentAnswerId"`
	OpponentCorrect   bool   `json:"opponentCorrect"`
	OpponentTime      int64  `json:"opponentTime"`
	CorrectAnswerID   string `json:"correctAnswerId"`
}

type GetGameResultOutput struct {
	GameID           string                  `json:"gameId"`
	Result           string                  `json:"result"` // "win", "loss", "draw"
	PlayerScore      int                     `json:"playerScore"`
	OpponentScore    int                     `json:"opponentScore"`
	MMRChange        int                     `json:"mmrChange"`
	NewMMR           int                     `json:"newMmr"`
	RankChange       *string                 `json:"rankChange,omitempty"` // "promoted", "demoted", null
	NewLeague        string                  `json:"newLeague"`
	NewDivision      int                     `json:"newDivision"`
	Opponent         DuelPlayerDTO           `json:"opponent"`
	Questions        []GameQuestionResultDTO `json:"questions"`
	CanRematch       bool                    `json:"canRematch"`
	RematchExpiresIn *int                    `json:"rematchExpiresIn,omitempty"`
}

// ========================================
// RequestRematch Use Case
// ========================================

type RequestRematchInput struct {
	PlayerID string `json:"playerId"`
	GameID   string `json:"gameId"`
}

type RequestRematchOutput struct {
	RematchID string  `json:"rematchId"`
	ExpiresIn int     `json:"expiresIn"`
	Status    string  `json:"status"`
	GameID    *string `json:"gameId,omitempty"`
}

// ========================================
// GetGameHistory Use Case
// ========================================

type GetGameHistoryInput struct {
	PlayerID string `json:"playerId"`
	Limit    int    `json:"limit"`
	Offset   int    `json:"offset"`
	Filter   string `json:"filter"` // "all", "friends", "wins", "losses"
}

type GetGameHistoryOutput struct {
	Games   []GameHistoryEntryDTO `json:"games"`
	Total   int                   `json:"total"`
	HasMore bool                  `json:"hasMore"`
}

// ========================================
// GetLeaderboard Use Case
// ========================================

type GetLeaderboardInput struct {
	PlayerID string `json:"playerId"`
	Type     string `json:"type"`  // "seasonal", "friends", "referrals"
	Limit    int    `json:"limit"`
}

type GetLeaderboardOutput struct {
	Type       string                `json:"type"`
	SeasonID   string                `json:"seasonId,omitempty"`
	EndsAt     int64                 `json:"endsAt,omitempty"`
	Entries    []LeaderboardEntryDTO `json:"entries"`
	PlayerRank int                   `json:"playerRank"`
}

// ========================================
// GetReferrals Use Case
// ========================================

type GetReferralsInput struct {
	PlayerID string `json:"playerId"`
}

type GetReferralsOutput struct {
	ReferralLink          string        `json:"referralLink"`
	TotalReferrals        int           `json:"totalReferrals"`
	ActiveReferrals       int           `json:"activeReferrals"`
	PendingRewards        []string      `json:"pendingRewards"`
	ReferralLeaderboardRank int         `json:"referralLeaderboardRank"`
	Referrals             []ReferralDTO `json:"referrals"`
}

// ========================================
// ClaimReferralReward Use Case
// ========================================

type ClaimReferralRewardInput struct {
	PlayerID  string `json:"playerId"`
	FriendID  string `json:"friendId"`
	Milestone string `json:"milestone"`
}

type ClaimReferralRewardOutput struct {
	Success          bool              `json:"success"`
	Rewards          ReferralRewardDTO `json:"rewards"`
	NewTicketBalance int               `json:"newTicketBalance"`
	NewCoinBalance   int               `json:"newCoinBalance"`
}

// ========================================
// StartGame Use Case
// ========================================

type StartGameInput struct {
	Player1ID       string `json:"player1Id"`
	Player2ID       string `json:"player2Id"`
	Player1Username string `json:"player1Username"`
	Player2Username string `json:"player2Username"`
	IsFriendGame    bool   `json:"isFriendGame"`
}

type StartGameOutput struct {
	GameID    string `json:"gameId"`
	Player1ID string `json:"player1Id"`
	Player2ID string `json:"player2Id"`
	StartsAt  int64  `json:"startsAt"`
}

// RoundQuestionOutput represents a question for a round
type RoundQuestionOutput struct {
	QuestionID   string              `json:"questionId"`
	QuestionText string              `json:"questionText"`
	Answers      []map[string]string `json:"answers"`
}

// ========================================
// SubmitDuelAnswer Use Case
// ========================================

type SubmitDuelAnswerInput struct {
	PlayerID   string `json:"playerId"`
	GameID     string `json:"gameId"`
	QuestionID string `json:"questionId"`
	AnswerID   string `json:"answerId"`
	TimeTaken  int    `json:"timeTaken"` // milliseconds
}

type SubmitDuelAnswerOutput struct {
	IsCorrect        bool   `json:"isCorrect"`
	CorrectAnswerID  string `json:"correctAnswerId"`
	PointsEarned     int    `json:"pointsEarned"`
	Player1Score     int    `json:"player1Score"`
	Player2Score     int    `json:"player2Score"`
	RoundComplete    bool   `json:"roundComplete"`
	GameComplete     bool   `json:"gameComplete"`
	WinnerID         string `json:"winnerId,omitempty"`
	Player1MMRChange int    `json:"player1MmrChange,omitempty"`
	Player2MMRChange int    `json:"player2MmrChange,omitempty"`
	Player1NewMMR    int    `json:"player1NewMmr,omitempty"`
	Player2NewMMR    int    `json:"player2NewMmr,omitempty"`
}
