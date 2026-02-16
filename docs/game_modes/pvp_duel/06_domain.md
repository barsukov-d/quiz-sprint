# PvP Duel - Domain Model

## Bounded Context
`pvp_duel` - Real-time 1v1 competitive quiz mode with ranking and social features.

---

## Aggregates

### DuelGame (Root)

**Purpose:** Represents a single 1v1 quiz duel between two players.

```go
type DuelGame struct {
    id              GameID
    status          GameStatus          // waiting, countdown, in_progress, completed, cancelled

    player1         *DuelPlayer
    player2         *DuelPlayer

    questions       []DuelQuestion      // 7 questions
    currentQuestion int                 // 0-6

    isFriendGame    bool
    challengeID     *ChallengeID        // If from challenge

    startedAt       int64
    completedAt     *int64

    events          []Event
}
```

**Invariants:**
- ✅ Exactly 2 players
- ✅ Exactly 7 questions
- ✅ Both players answer same questions
- ✅ Winner determined by score, then time
- ✅ MMR change calculated after completion

**Factory:**
```go
func NewDuelGame(
    player1 *Player,
    player2 *Player,
    questions []Question,
    isFriendGame bool,
    startedAt int64,
) (*DuelGame, error)
```

**Key Methods:**
```go
func (dg *DuelGame) SubmitAnswer(
    playerID PlayerID,
    questionID QuestionID,
    answerID AnswerID,
    timeTaken int64,
    submittedAt int64,
) (*AnswerResult, error)

func (dg *DuelGame) GetCurrentQuestion() *DuelQuestion
func (dg *DuelGame) IsComplete() bool
func (dg *DuelGame) GetWinner() (*PlayerID, WinReason)
func (dg *DuelGame) CalculateMMRChanges() (player1Delta, player2Delta int)
func (dg *DuelGame) Forfeit(playerID PlayerID) error
```

**Repository:**
```go
type DuelGameRepository interface {
    GetByID(id GameID) (*DuelGame, error)
    GetActiveByPlayer(playerID PlayerID) (*DuelGame, error)
    Save(dg *DuelGame) error
}
```

---

### DuelChallenge

**Purpose:** Friend challenge request.

```go
type DuelChallenge struct {
    id            ChallengeID
    challengerID  PlayerID
    challengedID  *PlayerID       // nil if link-based
    status        ChallengeStatus // pending, accepted, declined, expired

    challengeLink *string         // For link-based challenges
    expiresAt     int64

    createdAt     int64
    respondedAt   *int64
}
```

**Factory:**
```go
func NewDirectChallenge(
    challengerID PlayerID,
    challengedID PlayerID,
    expiresAt int64,
) *DuelChallenge

func NewLinkChallenge(
    challengerID PlayerID,
    expiresAt int64,
) *DuelChallenge
```

---

### PlayerRating

**Purpose:** Player's MMR and league standing.

```go
type PlayerRating struct {
    playerID     PlayerID
    mmr          int
    league       League
    division     int          // 1-4 (I-IV)

    peakMMR      int          // Season peak
    peakLeague   League
    peakDivision int

    gamesAtRank  int          // For demotion protection

    seasonID     SeasonID
}
```

**Methods:**
```go
func (pr *PlayerRating) ApplyMatchResult(opponentMMR int, won bool) int
func (pr *PlayerRating) GetLeagueLabel() string  // "🥇 Gold III"
func (pr *PlayerRating) CanDemote() bool
func (pr *PlayerRating) CheckPromotion(newMMR int) *PromotionEvent
```

---

### Referral

**Purpose:** Track friend invitations and rewards.

```go
type Referral struct {
    id           ReferralID
    inviterID    PlayerID
    inviteeID    PlayerID

    milestones   map[Milestone]bool
    rewards      map[Milestone]RewardClaim

    createdAt    int64
}

type Milestone string
const (
    MilestoneRegistered     Milestone = "registered"
    MilestonePlayed5Duels   Milestone = "played_5_duels"
    MilestoneReachedSilver  Milestone = "reached_silver"
    MilestoneReachedGold    Milestone = "reached_gold"
    MilestoneReachedPlatinum Milestone = "reached_platinum"
)
```

---

### PlayerTickets

**Purpose:** Track PvP tickets for duel entry.

**Note:** Tickets are stored in `user.Inventory` (User Context), not in PvP Duel context.

```go
// Query tickets via User Context
type TicketService interface {
    GetTicketBalance(playerID PlayerID) (int, error)
    ConsumeTicket(playerID PlayerID) error
    RefundTicket(playerID PlayerID) error
    AddTickets(playerID PlayerID, amount int, source TicketSource) error
}

type TicketSource string
const (
    TicketSourceDailyChallenge TicketSource = "daily_challenge"
    TicketSourceDailyMission   TicketSource = "daily_mission"
    TicketSourceWeeklyMission  TicketSource = "weekly_mission"
    TicketSourceReferral       TicketSource = "referral"
    TicketSourceSeasonal       TicketSource = "seasonal"
    TicketSourcePurchase       TicketSource = "purchase"
    TicketSourceFriendDuel     TicketSource = "friend_duel_bonus"
)
```

**Ticket acquisition from Daily Challenge:**
| Daily Challenge Result | Tickets Earned |
|----------------------|----------------|
| 0-4 correct (Wooden Chest) | 1 🎟️ |
| 5-7 correct (Silver Chest) | 2-3 🎟️ |
| 8-10 correct (Golden Chest) | 4-5 🎟️ |

---

## Value Objects

### DuelPlayer

```go
type DuelPlayer struct {
    id          PlayerID
    username    string
    avatar      string
    mmrBefore   int
    mmrAfter    *int

    answers     []PlayerAnswer
    totalTime   int64
    score       int
}

func (dp *DuelPlayer) AddAnswer(answer PlayerAnswer) {
    dp.answers = append(dp.answers, answer)
    dp.totalTime += answer.TimeTaken
    if answer.IsCorrect {
        dp.score++
    }
}
```

---

### PlayerAnswer

```go
type PlayerAnswer struct {
    questionID  QuestionID
    answerID    *AnswerID     // nil if timeout
    isCorrect   bool
    timeTaken   int64         // milliseconds
    answeredAt  int64
}
```

---

### League

```go
type League string

const (
    LeagueBronze   League = "bronze"
    LeagueSilver   League = "silver"
    LeagueGold     League = "gold"
    LeaguePlatinum League = "platinum"
    LeagueDiamond  League = "diamond"
    LeagueLegend   League = "legend"
)

func GetLeagueFromMMR(mmr int) (League, int) {
    switch {
    case mmr >= 3000:
        return LeagueLegend, 1
    case mmr >= 2500:
        return LeagueDiamond, 4 - (mmr-2500)/125
    case mmr >= 2000:
        return LeaguePlatinum, 4 - (mmr-2000)/125
    case mmr >= 1500:
        return LeagueGold, 4 - (mmr-1500)/125
    case mmr >= 1000:
        return LeagueSilver, 4 - (mmr-1000)/125
    default:
        return LeagueBronze, 4 - mmr/250
    }
}

func (l League) Icon() string {
    icons := map[League]string{
        LeagueBronze:   "🥉",
        LeagueSilver:   "🥈",
        LeagueGold:     "🥇",
        LeaguePlatinum: "💍",
        LeagueDiamond:  "💎",
        LeagueLegend:   "👑",
    }
    return icons[l]
}
```

---

### GameStatus

```go
type GameStatus string

const (
    GameStatusWaiting    GameStatus = "waiting"
    GameStatusCountdown  GameStatus = "countdown"
    GameStatusInProgress GameStatus = "in_progress"
    GameStatusCompleted  GameStatus = "completed"
    GameStatusCancelled  GameStatus = "cancelled"
)
```

---

### WinReason

```go
type WinReason string

const (
    WinReasonScore   WinReason = "score"
    WinReasonTime    WinReason = "time"
    WinReasonForfeit WinReason = "forfeit"
)
```

---

## Domain Services

### MMRCalculator

```go
type MMRCalculator struct {
    kFactor int  // Default 32
}

func (c *MMRCalculator) Calculate(winnerMMR, loserMMR int) (winnerDelta, loserDelta int) {
    expectedWinner := 1.0 / (1.0 + math.Pow(10, float64(loserMMR-winnerMMR)/400))

    winnerDelta = int(math.Round(float64(c.kFactor) * (1.0 - expectedWinner)))
    loserDelta = -int(math.Round(float64(c.kFactor) * expectedWinner))

    // Minimum change
    if winnerDelta < 10 {
        winnerDelta = 10
    }
    if loserDelta > -10 {
        loserDelta = -10
    }

    return winnerDelta, loserDelta
}
```

---

### MatchmakingService

```go
type MatchmakingService struct {
    queue      *MatchmakingQueue
    repository DuelGameRepository
}

func (s *MatchmakingService) JoinQueue(player *Player) (*QueueEntry, error) {
    // Check not already in queue/match
    if s.queue.Contains(player.ID) {
        return nil, ErrAlreadyInQueue
    }

    if _, err := s.repository.GetActiveByPlayer(player.ID); err == nil {
        return nil, ErrAlreadyInGame
    }

    entry := &QueueEntry{
        PlayerID:  player.ID,
        MMR:       player.Rating.MMR,
        JoinedAt:  time.Now().Unix(),
    }

    s.queue.Add(entry)

    return entry, nil
}

func (s *MatchmakingService) FindOpponent(entry *QueueEntry) (*DuelGame, error) {
    elapsed := time.Since(time.Unix(entry.JoinedAt, 0))
    mmrRange := s.calculateMMRRange(elapsed)

    opponent := s.queue.FindInRange(
        entry.MMR - mmrRange,
        entry.MMR + mmrRange,
        entry.PlayerID,
    )

    if opponent == nil {
        return nil, ErrNoOpponentFound
    }

    // Create game
    game := NewDuelGame(...)

    return game, nil
}
```

---

### ChallengeService

```go
type ChallengeService struct {
    repository ChallengeRepository
    notifier   NotificationService
}

func (s *ChallengeService) CreateChallenge(
    challengerID PlayerID,
    challengedID PlayerID,
) (*DuelChallenge, error) {
    // Check friend not busy
    if s.IsBusy(challengedID) {
        return nil, ErrFriendBusy
    }

    challenge := NewDirectChallenge(
        challengerID,
        challengedID,
        time.Now().Add(60 * time.Second).Unix(),
    )

    if err := s.repository.Save(challenge); err != nil {
        return nil, err
    }

    // Send notification
    s.notifier.SendChallengeNotification(challengedID, challenge)

    return challenge, nil
}
```

---

### ReferralService

```go
type ReferralService struct {
    repository ReferralRepository
    rewards    RewardService
}

func (s *ReferralService) CheckMilestones(inviteeID PlayerID) error {
    referral := s.repository.GetByInvitee(inviteeID)
    if referral == nil {
        return nil  // Not a referred user
    }

    invitee := s.getPlayer(inviteeID)

    // Check each milestone
    if invitee.DuelsPlayed >= 5 && !referral.milestones[MilestonePlayed5Duels] {
        referral.milestones[MilestonePlayed5Duels] = true
        s.notifyInviter(referral.inviterID, MilestonePlayed5Duels)
    }

    if invitee.Rating.League >= LeagueSilver && !referral.milestones[MilestoneReachedSilver] {
        referral.milestones[MilestoneReachedSilver] = true
        s.notifyInviter(referral.inviterID, MilestoneReachedSilver)
    }

    // ... similar for other milestones

    return s.repository.Save(referral)
}
```

---

## Domain Events

```go
type DuelGameCreatedEvent struct {
    GameID        GameID
    Player1ID     PlayerID
    Player2ID     PlayerID
    IsFriendGame  bool
    Timestamp     int64
}

type PlayerAnsweredEvent struct {
    GameID        GameID
    PlayerID      PlayerID
    QuestionIndex int
    IsCorrect     bool
    TimeTaken     int64
    Timestamp     int64
}

type DuelGameFinishedEvent struct {
    GameID         GameID
    WinnerID       PlayerID
    LoserID        PlayerID
    WinnerScore    int
    LoserScore     int
    WinReason      WinReason
    WinnerMMRDelta int
    LoserMMRDelta  int
    IsFriendGame   bool
    Timestamp      int64
}

type PlayerPromotedEvent struct {
    PlayerID     PlayerID
    OldLeague    League
    OldDivision  int
    NewLeague    League
    NewDivision  int
    Timestamp    int64
}

type PlayerDemotedEvent struct {
    PlayerID     PlayerID
    OldLeague    League
    OldDivision  int
    NewLeague    League
    NewDivision  int
    Timestamp    int64
}

type ReferralCreatedEvent struct {
    ReferralID ReferralID
    InviterID  PlayerID
    InviteeID  PlayerID
    Timestamp  int64
}

type ReferralMilestoneEvent struct {
    ReferralID ReferralID
    InviterID  PlayerID
    InviteeID  PlayerID
    Milestone  Milestone
    Timestamp  int64
}

type SeasonResetEvent struct {
    SeasonID       SeasonID
    TopPlayers     []PlayerID
    RewardsQueued  int
    Timestamp      int64
}
```

---

## Database Schema

### duel_games
```sql
CREATE TABLE duel_games (
    id VARCHAR(36) PRIMARY KEY,
    status VARCHAR(20) NOT NULL,

    player1_id VARCHAR(36) NOT NULL,
    player2_id VARCHAR(36) NOT NULL,
    winner_id VARCHAR(36),

    player1_score INT DEFAULT 0,
    player2_score INT DEFAULT 0,
    player1_total_time INT DEFAULT 0,
    player2_total_time INT DEFAULT 0,

    player1_mmr_before INT,
    player1_mmr_after INT,
    player2_mmr_before INT,
    player2_mmr_after INT,

    win_reason VARCHAR(20),
    is_friend_game BOOLEAN DEFAULT FALSE,
    challenge_id VARCHAR(36),

    questions_data JSONB NOT NULL,
    answers_data JSONB,

    started_at TIMESTAMP,
    completed_at TIMESTAMP,

    INDEX idx_player1 (player1_id, completed_at DESC),
    INDEX idx_player2 (player2_id, completed_at DESC),
    INDEX idx_status (status)
);
```

### player_ratings
```sql
CREATE TABLE player_ratings (
    player_id VARCHAR(36) PRIMARY KEY,
    mmr INT DEFAULT 1000,
    league VARCHAR(20) DEFAULT 'silver',
    division INT DEFAULT 4,

    peak_mmr INT DEFAULT 1000,
    peak_league VARCHAR(20) DEFAULT 'silver',
    peak_division INT DEFAULT 4,

    games_at_rank INT DEFAULT 0,

    season_id VARCHAR(20),
    season_wins INT DEFAULT 0,
    season_losses INT DEFAULT 0,

    updated_at TIMESTAMP
);
```

### duel_challenges
```sql
CREATE TABLE duel_challenges (
    id VARCHAR(36) PRIMARY KEY,
    challenger_id VARCHAR(36) NOT NULL,
    challenged_id VARCHAR(36),

    status VARCHAR(20) DEFAULT 'pending',
    challenge_link VARCHAR(100),

    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL,
    responded_at TIMESTAMP,

    INDEX idx_challenged (challenged_id, status),
    INDEX idx_expires (expires_at)
);
```

### referrals
```sql
CREATE TABLE referrals (
    id VARCHAR(36) PRIMARY KEY,
    inviter_id VARCHAR(36) NOT NULL,
    invitee_id VARCHAR(36) NOT NULL,

    milestone_registered BOOLEAN DEFAULT TRUE,
    milestone_played_5 BOOLEAN DEFAULT FALSE,
    milestone_silver BOOLEAN DEFAULT FALSE,
    milestone_gold BOOLEAN DEFAULT FALSE,
    milestone_platinum BOOLEAN DEFAULT FALSE,

    inviter_rewards_claimed JSONB,
    invitee_rewards_claimed JSONB,

    created_at TIMESTAMP NOT NULL,

    UNIQUE (inviter_id, invitee_id),
    INDEX idx_inviter (inviter_id),
    INDEX idx_invitee (invitee_id)
);
```

### seasons
```sql
CREATE TABLE seasons (
    id VARCHAR(20) PRIMARY KEY,  -- "2026-S04"
    starts_at TIMESTAMP NOT NULL,
    ends_at TIMESTAMP NOT NULL,
    status VARCHAR(20) DEFAULT 'active',
    rewards_distributed BOOLEAN DEFAULT FALSE
);
```

---

## Redis Structures

### Matchmaking Queue
```
Key: duel:queue
Type: Sorted Set
Score: MMR
Member: playerID:joinedAt

ZADD duel:queue 1650 "user_123:1706429000"
ZRANGEBYSCORE duel:queue 1600 1700 LIMIT 0 1
```

### Active Matches
```
Key: duel:active:{gameId}
Type: Hash

HSET duel:active:g_xyz789 status "in_progress" currentQuestion 3 ...
HGET duel:active:g_xyz789 status
```

### Player Online Status
```
Key: duel:online:{playerId}
Type: String (with TTL)

SETEX duel:online:user_123 60 "1"  # Online for 60s
```

### Seasonal Leaderboard
```
Key: duel:leaderboard:seasonal:{seasonId}
Type: Sorted Set
Score: MMR
Member: playerID

ZADD duel:leaderboard:seasonal:2026-S04 1678 "user_123"
ZREVRANK duel:leaderboard:seasonal:2026-S04 "user_123"
```

---

## Error Types

```go
var (
    ErrGameNotFound       = errors.New("match not found")
    ErrAlreadyInQueue      = errors.New("already in queue")
    ErrAlreadyInGame      = errors.New("already in match")
    ErrNoOpponentFound     = errors.New("no opponent found")
    ErrInsufficientTickets = errors.New("insufficient tickets")
    ErrChallengeExpired    = errors.New("challenge expired")
    ErrFriendBusy          = errors.New("friend is busy")
    ErrNotYourTurn         = errors.New("not your turn")
    ErrQuestionTimeout     = errors.New("question timeout")
    ErrInvalidAnswer       = errors.New("invalid answer")
)
```

---

## Integration with Other Contexts

### User Context
```
DuelGame --[PlayerID]--> user.User
Tickets consumed → user.Inventory
Rewards → user.Inventory
```

### Quiz Context
```
DuelGame uses quiz.Question
Questions filtered by difficulty (medium only for duels)
```

### Daily Challenge Context
```
Daily Challenge → Earn PvP tickets
```

### Notification Context
```
ChallengeService → NotificationService
ReferralService → NotificationService
SeasonService → NotificationService
```
