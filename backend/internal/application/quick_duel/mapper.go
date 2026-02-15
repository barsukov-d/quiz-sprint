package quick_duel

import (
	"github.com/barsukov/quiz-sprint/backend/internal/domain/quick_duel"
	"github.com/barsukov/quiz-sprint/backend/internal/domain/quiz"
)

// ToPlayerRatingDTO converts domain PlayerRating to DTO
func ToPlayerRatingDTO(rating *quick_duel.PlayerRating) PlayerRatingDTO {
	leagueInfo := quick_duel.GetLeagueFromMMR(rating.MMR())
	return PlayerRatingDTO{
		PlayerID:     rating.PlayerID().String(),
		MMR:          rating.MMR(),
		League:       rating.League().String(),
		Division:     rating.Division().Value(),
		LeagueLabel:  leagueInfo.Label(),
		LeagueIcon:   rating.League().Icon(),
		PeakMMR:      rating.PeakMMR(),
		PeakLeague:   rating.PeakLeague().String(),
		SeasonWins:   rating.SeasonWins(),
		SeasonLosses: rating.SeasonLosses(),
		WinRate:      rating.WinRate(),
		GamesAtRank:  rating.GamesAtRank(),
		CanDemote:    rating.CanDemote(),
	}
}

// ToDuelPlayerDTO converts domain DuelPlayer to DTO
func ToDuelPlayerDTO(player quick_duel.DuelPlayer, rating *quick_duel.PlayerRating) DuelPlayerDTO {
	league := quick_duel.LeagueBronze
	division := quick_duel.DivisionIV
	if rating != nil {
		league = rating.League()
		division = rating.Division()
	}

	return DuelPlayerDTO{
		ID:         player.UserID().String(),
		Username:   player.Username(),
		Avatar:     "", // TODO: get from user profile
		MMR:        player.Elo().Rating(),
		League:     league.String(),
		Division:   division.Value(),
		LeagueIcon: league.Icon(),
		Score:      player.Score(),
		Connected:  player.Connected(),
	}
}

// ToDuelGameDTO converts domain DuelGame to DTO
func ToDuelGameDTO(game *quick_duel.DuelGame, player1Rating, player2Rating *quick_duel.PlayerRating) DuelGameDTO {
	var winnerID *string
	if game.Status() == quick_duel.GameStatusFinished {
		if game.Player1().Score() > game.Player2().Score() {
			id := game.Player1().UserID().String()
			winnerID = &id
		} else if game.Player2().Score() > game.Player1().Score() {
			id := game.Player2().UserID().String()
			winnerID = &id
		}
	}

	return DuelGameDTO{
		ID:           game.ID().String(),
		Status:       string(game.Status()),
		Player1:      ToDuelPlayerDTO(game.Player1(), player1Rating),
		Player2:      ToDuelPlayerDTO(game.Player2(), player2Rating),
		CurrentRound: game.CurrentRound(),
		TotalRounds:  quick_duel.QuestionsPerDuel,
		StartedAt:    game.StartedAt(),
		FinishedAt:   game.FinishedAt(),
		WinnerID:     winnerID,
		IsFriendGame: false, // TODO: check if friend game
	}
}

// ToDuelQuestionDTO converts domain Question to DTO (without isCorrect)
func ToDuelQuestionDTO(question *quiz.Question, questionNum int, serverTime int64) DuelQuestionDTO {
	answers := make([]DuelAnswerDTO, 0, len(question.Answers()))
	for _, ans := range question.Answers() {
		answers = append(answers, DuelAnswerDTO{
			ID:   ans.ID().String(),
			Text: ans.Text().String(),
		})
	}

	return DuelQuestionDTO{
		ID:          question.ID().String(),
		QuestionNum: questionNum,
		Text:        question.Text().String(),
		Answers:     answers,
		TimeLimit:   quick_duel.TimePerQuestionSec,
		ServerTime:  serverTime,
	}
}

// ToChallengeDTO converts domain DuelChallenge to DTO
func ToChallengeDTO(challenge *quick_duel.DuelChallenge, now int64) ChallengeDTO {
	var challengedID *string
	if challenge.ChallengedID() != nil {
		id := challenge.ChallengedID().String()
		challengedID = &id
	}

	expiresIn := int(challenge.ExpiresAt() - now)
	if expiresIn < 0 {
		expiresIn = 0
	}

	return ChallengeDTO{
		ID:            challenge.ID().String(),
		ChallengerID:  challenge.ChallengerID().String(),
		ChallengedID:  challengedID,
		Type:          string(challenge.Type()),
		Status:        string(challenge.Status()),
		ChallengeLink: challenge.ChallengeLink(),
		ExpiresAt:     challenge.ExpiresAt(),
		ExpiresIn:     expiresIn,
		CreatedAt:     challenge.CreatedAt(),
	}
}

// ToGameHistoryEntryDTO converts domain DuelGame to history entry DTO
func ToGameHistoryEntryDTO(game *quick_duel.DuelGame, playerID string, opponentUsername string) GameHistoryEntryDTO {
	isPlayer1 := game.Player1().UserID().String() == playerID

	var result string
	var playerScore, opponentScore int
	var opponentMMR int
	var mmrChange int

	if isPlayer1 {
		playerScore = game.Player1().Score()
		opponentScore = game.Player2().Score()
		opponentMMR = game.Player2().Elo().Rating()
		mmrChange = game.Player1().Elo().Rating() - game.Player1().Elo().Rating() // TODO: store before/after
	} else {
		playerScore = game.Player2().Score()
		opponentScore = game.Player1().Score()
		opponentMMR = game.Player1().Elo().Rating()
		mmrChange = game.Player2().Elo().Rating() - game.Player2().Elo().Rating()
	}

	if playerScore > opponentScore {
		result = "win"
	} else if playerScore < opponentScore {
		result = "loss"
	} else {
		result = "draw"
	}

	return GameHistoryEntryDTO{
		GameID:        game.ID().String(),
		Opponent:      opponentUsername,
		OpponentMMR:   opponentMMR,
		Result:        result,
		PlayerScore:   playerScore,
		OpponentScore: opponentScore,
		MMRChange:     mmrChange,
		IsFriendGame:  false, // TODO
		CompletedAt:   game.FinishedAt(),
	}
}

// ToLeaderboardEntryDTO converts domain PlayerRating to leaderboard entry DTO
func ToLeaderboardEntryDTO(rating *quick_duel.PlayerRating, rank int, username string) LeaderboardEntryDTO {
	return LeaderboardEntryDTO{
		Rank:       rank,
		PlayerID:   rating.PlayerID().String(),
		Username:   username,
		Avatar:     "",
		MMR:        rating.MMR(),
		League:     rating.League().String(),
		LeagueIcon: rating.League().Icon(),
		Wins:       rating.SeasonWins(),
		Losses:     rating.SeasonLosses(),
		WinRate:    rating.WinRate(),
	}
}

// ToReferralDTO converts domain Referral to DTO
func ToReferralDTO(referral *quick_duel.Referral, inviteeUsername string) ReferralDTO {
	milestonesReached := make([]string, 0)
	if referral.MilestoneRegistered() {
		milestonesReached = append(milestonesReached, quick_duel.MilestoneRegistered)
	}
	if referral.MilestonePlayedFive() {
		milestonesReached = append(milestonesReached, quick_duel.MilestonePlayedFive)
	}
	if referral.MilestoneReachedSilver() {
		milestonesReached = append(milestonesReached, quick_duel.MilestoneReachedSilver)
	}
	if referral.MilestoneReachedGold() {
		milestonesReached = append(milestonesReached, quick_duel.MilestoneReachedGold)
	}
	if referral.MilestoneReachedPlatinum() {
		milestonesReached = append(milestonesReached, quick_duel.MilestoneReachedPlatinum)
	}

	return ReferralDTO{
		ID:                referral.ID().String(),
		InviteeID:         referral.InviteeID().String(),
		InviteeUsername:   inviteeUsername,
		MilestonesReached: milestonesReached,
		PendingRewards:    referral.GetPendingInviterRewards(),
		CreatedAt:         referral.CreatedAt(),
	}
}

// ToReferralRewardDTO converts domain ReferralReward to DTO
func ToReferralRewardDTO(reward *quick_duel.ReferralReward) ReferralRewardDTO {
	return ReferralRewardDTO{
		Tickets: reward.Tickets,
		Coins:   reward.Coins,
		Badge:   reward.Badge,
		Avatar:  reward.Avatar,
		Title:   reward.Title,
	}
}
