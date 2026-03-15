package messaging

import (
	"log"

	appDuel "github.com/barsukov/quiz-sprint/backend/internal/application/quick_duel"
	domainDuel "github.com/barsukov/quiz-sprint/backend/internal/domain/quick_duel"
)

// LobbyEventBus implements appDuel.EventBus by routing domain events
// to the lobby WebSocket hub.
type LobbyEventBus struct {
	hub appDuel.LobbyHub
}

func NewLobbyEventBus(hub appDuel.LobbyHub) *LobbyEventBus {
	return &LobbyEventBus{hub: hub}
}

// Publish routes a domain event to the appropriate player(s) via lobby WS.
func (b *LobbyEventBus) Publish(event domainDuel.Event) {
	switch e := event.(type) {
	case domainDuel.ChallengeCreatedEvent:
		// Direct challenge: notify invitee if connected
		if e.ChallengedID() != nil {
			b.hub.Notify(e.ChallengedID().String(), appDuel.LobbyEvent{
				Type: "challenge_received",
				Data: map[string]interface{}{
					"challengeId": e.ChallengeID().String(),
					"expiresIn":   domainDuel.DirectChallengeExpirySeconds,
				},
			})
		}

	case domainDuel.ChallengeAcceptedEvent:
		// Notify the challenger (inviter) that invitee accepted
		b.hub.Notify(e.ChallengerID().String(), appDuel.LobbyEvent{
			Type: "challenge_accepted",
			Data: map[string]interface{}{
				"challengeId": e.ChallengeID().String(),
				"inviteeId":   e.AccepterID().String(),
			},
		})

	case domainDuel.ChallengeDeclinedEvent:
		// Notify inviter that invitee declined
		b.hub.Notify(e.ChallengerID().String(), appDuel.LobbyEvent{
			Type: "challenge_declined",
			Data: map[string]interface{}{
				"challengeId": e.ChallengeID().String(),
			},
		})

	case domainDuel.ChallengeExpiredEvent:
		b.hub.Notify(e.ChallengerID().String(), appDuel.LobbyEvent{
			Type: "challenge_expired",
			Data: map[string]interface{}{"challengeId": e.ChallengeID().String()},
		})
		if e.ChallengedID() != nil {
			b.hub.Notify(e.ChallengedID().String(), appDuel.LobbyEvent{
				Type: "challenge_expired",
				Data: map[string]interface{}{"challengeId": e.ChallengeID().String()},
			})
		}

	case domainDuel.GameReadyEvent:
		if e.PlayerID() != nil {
			b.hub.Notify(e.PlayerID().String(), appDuel.LobbyEvent{
				Type: "game_ready",
				Data: map[string]string{"gameId": e.GameID()},
			})
		}

	case domainDuel.PlayerSurrenderedEvent:
		// Notify the opponent that the other player surrendered (game is over)
		b.hub.Notify(e.OpponentID().String(), appDuel.LobbyEvent{
			Type: "game_surrendered",
			Data: map[string]interface{}{
				"gameId":      e.GameID().String(),
				"surrenderId": e.PlayerID().String(),
				"winnerId":    e.OpponentID().String(),
			},
		})

	default:
		log.Printf("[LobbyEventBus] unhandled event type: %T", event)
	}
}
