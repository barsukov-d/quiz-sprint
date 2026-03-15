package quick_duel

import (
	"testing"
	"time"

	"github.com/barsukov/quiz-sprint/backend/internal/domain/quick_duel"
)

// mockInventoryService is a simple in-memory inventory service for tests
type mockInventoryService struct {
	credits []inventoryCredit
}

type inventoryCredit struct {
	playerID string
	source   string
	details  map[string]int
}

func (m *mockInventoryService) GetPvpTickets(playerID string) (int, error) {
	return 0, nil
}

func (m *mockInventoryService) Credit(playerID string, source string, details map[string]int) error {
	m.credits = append(m.credits, inventoryCredit{playerID: playerID, source: source, details: details})
	return nil
}

func (m *mockInventoryService) Debit(playerID string, source string, details map[string]int) error {
	return nil
}

func TestSeasonalResetUseCase_ResetsAllPlayers(t *testing.T) {
	f := setupFixture(t)

	// Seed two players with different MMRs in the current season
	now := time.Now().UTC().Unix()
	p1, _ := f.playerRatingRepo.FindOrCreate(mustUserID(testPlayer1ID), "2026-02", now)
	// Artificially bump MMR via game results is complex; reconstruct directly
	_ = p1

	// Seed player2 with high MMR
	p2Rating := quick_duel.ReconstructPlayerRating(
		mustUserID(testPlayer2ID),
		2000, // Platinum
		quick_duel.LeaguePlatinum, quick_duel.DivisionIV,
		2000, quick_duel.LeaguePlatinum, quick_duel.DivisionIV,
		5, "2026-02", 10, 2, 1, now,
	)
	if err := f.playerRatingRepo.Save(p2Rating); err != nil {
		t.Fatalf("save p2 rating: %v", err)
	}

	uc := NewSeasonalResetUseCase(f.playerRatingRepo, f.seasonRepo)
	out, err := uc.Execute(SeasonalResetInput{NewSeasonID: "2026-03"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if out.OldSeasonID != "2026-02" {
		t.Errorf("OldSeasonID = %s, want 2026-02", out.OldSeasonID)
	}
	if out.NewSeasonID != "2026-03" {
		t.Errorf("NewSeasonID = %s, want 2026-03", out.NewSeasonID)
	}
	// 2 players were seeded (p1 via FindOrCreate, p2 via Save)
	if out.PlayersReset != 2 {
		t.Errorf("PlayersReset = %d, want 2", out.PlayersReset)
	}

	// After reset, p2 (2000 MMR) should be: 1000 + (2000-1000)*0.5 = 1500
	resetP2, err := f.playerRatingRepo.FindByPlayerID(mustUserID(testPlayer2ID))
	if err != nil {
		t.Fatalf("find p2 after reset: %v", err)
	}
	if resetP2.MMR() != 1500 {
		t.Errorf("p2 MMR after reset = %d, want 1500", resetP2.MMR())
	}
	if resetP2.SeasonID() != "2026-03" {
		t.Errorf("p2 SeasonID after reset = %s, want 2026-03", resetP2.SeasonID())
	}
}

func TestSeasonalResetUseCase_MinMMR500(t *testing.T) {
	f := setupFixture(t)
	now := time.Now().UTC().Unix()

	// Player with very low MMR (100): 1000 + (100-1000)*0.5 = 550, above 500 → 550
	lowRating := quick_duel.ReconstructPlayerRating(
		mustUserID(testPlayer1ID),
		100, quick_duel.LeagueBronze, quick_duel.DivisionIV,
		100, quick_duel.LeagueBronze, quick_duel.DivisionIV,
		0, "2026-02", 0, 5, 0, now,
	)
	if err := f.playerRatingRepo.Save(lowRating); err != nil {
		t.Fatalf("save: %v", err)
	}

	uc := NewSeasonalResetUseCase(f.playerRatingRepo, f.seasonRepo)
	_, err := uc.Execute(SeasonalResetInput{NewSeasonID: "2026-03"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	after, err := f.playerRatingRepo.FindByPlayerID(mustUserID(testPlayer1ID))
	if err != nil {
		t.Fatalf("find after reset: %v", err)
	}
	// 1000 + (100-1000)*0.5 = 1000 - 450 = 550 → above 500 minimum
	if after.MMR() != 550 {
		t.Errorf("MMR after reset = %d, want 550", after.MMR())
	}
}

func TestDistributeSeasonalRewardsUseCase_CreditsAndResets(t *testing.T) {
	f := setupFixture(t)
	now := time.Now().UTC().Unix()
	inv := &mockInventoryService{}

	// Seed a Gold player
	goldRating := quick_duel.ReconstructPlayerRating(
		mustUserID(testPlayer1ID),
		1700, quick_duel.LeagueGold, quick_duel.DivisionII,
		1700, quick_duel.LeagueGold, quick_duel.DivisionII,
		3, "2026-02", 5, 2, 0, now,
	)
	if err := f.playerRatingRepo.Save(goldRating); err != nil {
		t.Fatalf("save gold rating: %v", err)
	}

	resetUC := NewSeasonalResetUseCase(f.playerRatingRepo, f.seasonRepo)
	distUC := NewDistributeSeasonalRewardsUseCase(f.playerRatingRepo, f.seasonRepo, inv, resetUC)

	out, err := distUC.Execute(DistributeSeasonalRewardsInput{SeasonID: "2026-02"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if out.RewardsGranted != 1 {
		t.Errorf("RewardsGranted = %d, want 1", out.RewardsGranted)
	}
	if len(inv.credits) != 1 {
		t.Fatalf("expected 1 credit call, got %d", len(inv.credits))
	}

	credit := inv.credits[0]
	if credit.source != "seasonal_reward" {
		t.Errorf("source = %s, want seasonal_reward", credit.source)
	}
	if credit.details["coins"] != 1000 {
		t.Errorf("coins = %d, want 1000 (Gold)", credit.details["coins"])
	}
	if credit.details["pvp_tickets"] != 10 {
		t.Errorf("pvp_tickets = %d, want 10 (Gold)", credit.details["pvp_tickets"])
	}

	// After distribution the reset must have fired: player moves to new season
	after, err := f.playerRatingRepo.FindByPlayerID(mustUserID(testPlayer1ID))
	if err != nil {
		t.Fatalf("find after distribute: %v", err)
	}
	if after.SeasonID() == "2026-02" {
		t.Error("player should be in new season after reset, still in 2026-02")
	}
}
