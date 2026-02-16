package quick_duel

import "github.com/google/uuid"

// Referral milestones
const (
	MilestoneRegistered   = "registered"
	MilestonePlayedFive   = "played_5_duels"
	MilestoneReachedSilver = "reached_silver"
	MilestoneReachedGold   = "reached_gold"
	MilestoneReachedPlatinum = "reached_platinum"
)

// Referral rewards
type ReferralReward struct {
	Tickets int
	Coins   int
	Badge   string
	Avatar  string
	Title   string
}

// Milestone rewards configuration
var MilestoneRewards = map[string]struct {
	Inviter ReferralReward
	Invitee ReferralReward
}{
	MilestoneRegistered: {
		Inviter: ReferralReward{Tickets: 3, Coins: 100},
		Invitee: ReferralReward{Tickets: 3, Coins: 100},
	},
	MilestonePlayedFive: {
		Inviter: ReferralReward{Tickets: 5, Coins: 300},
		Invitee: ReferralReward{Coins: 200},
	},
	MilestoneReachedSilver: {
		Inviter: ReferralReward{Tickets: 10, Coins: 500, Badge: "Наставник"},
		Invitee: ReferralReward{Coins: 300},
	},
	MilestoneReachedGold: {
		Inviter: ReferralReward{Tickets: 20, Coins: 1000, Avatar: "exclusive_referrer"},
		Invitee: ReferralReward{Coins: 500},
	},
	MilestoneReachedPlatinum: {
		Inviter: ReferralReward{Tickets: 50, Coins: 3000, Title: "Легендарный наставник"},
		Invitee: ReferralReward{Coins: 1000},
	},
}

// ReferralID uniquely identifies a referral
type ReferralID struct {
	value string
}

func NewReferralID() ReferralID {
	return ReferralID{value: uuid.New().String()}
}

func NewReferralIDFromString(value string) ReferralID {
	return ReferralID{value: value}
}

func (id ReferralID) String() string { return id.value }
func (id ReferralID) IsZero() bool   { return id.value == "" }

// Referral represents a friend referral relationship (aggregate root)
type Referral struct {
	id                     ReferralID
	inviterID              UserID
	inviteeID              UserID
	milestoneRegistered    bool
	milestonePlayedFive    bool
	milestoneReachedSilver bool
	milestoneReachedGold   bool
	milestoneReachedPlatinum bool
	inviterRewardsClaimed  map[string]bool
	inviteeRewardsClaimed  map[string]bool
	createdAt              int64

	events []Event
}

// NewReferral creates a new referral relationship
func NewReferral(
	inviterID UserID,
	inviteeID UserID,
	createdAt int64,
) (*Referral, error) {
	if inviterID.IsZero() || inviteeID.IsZero() {
		return nil, ErrReferralNotFound
	}

	if inviterID.Equals(inviteeID) {
		return nil, ErrSelfReferral
	}

	referral := &Referral{
		id:                     NewReferralID(),
		inviterID:              inviterID,
		inviteeID:              inviteeID,
		milestoneRegistered:    true, // Automatically achieved on creation
		milestonePlayedFive:    false,
		milestoneReachedSilver: false,
		milestoneReachedGold:   false,
		milestoneReachedPlatinum: false,
		inviterRewardsClaimed:  make(map[string]bool),
		inviteeRewardsClaimed:  make(map[string]bool),
		createdAt:              createdAt,
		events:                 make([]Event, 0),
	}

	referral.events = append(referral.events, NewReferralCreatedEvent(
		ChallengeID{value: referral.id.value}, // Reusing ChallengeID type for simplicity
		inviterID,
		inviteeID,
		createdAt,
	))

	referral.events = append(referral.events, NewReferralMilestoneEvent(
		ChallengeID{value: referral.id.value},
		inviterID,
		inviteeID,
		MilestoneRegistered,
		createdAt,
	))

	return referral, nil
}

// ReconstructReferral reconstructs from persistence
func ReconstructReferral(
	id ReferralID,
	inviterID UserID,
	inviteeID UserID,
	milestoneRegistered bool,
	milestonePlayedFive bool,
	milestoneReachedSilver bool,
	milestoneReachedGold bool,
	milestoneReachedPlatinum bool,
	inviterRewardsClaimed map[string]bool,
	inviteeRewardsClaimed map[string]bool,
	createdAt int64,
) *Referral {
	return &Referral{
		id:                     id,
		inviterID:              inviterID,
		inviteeID:              inviteeID,
		milestoneRegistered:    milestoneRegistered,
		milestonePlayedFive:    milestonePlayedFive,
		milestoneReachedSilver: milestoneReachedSilver,
		milestoneReachedGold:   milestoneReachedGold,
		milestoneReachedPlatinum: milestoneReachedPlatinum,
		inviterRewardsClaimed:  inviterRewardsClaimed,
		inviteeRewardsClaimed:  inviteeRewardsClaimed,
		createdAt:              createdAt,
		events:                 make([]Event, 0),
	}
}

// UpdateProgress updates milestone progress based on invitee's stats
func (r *Referral) UpdateProgress(duelsPlayed int, currentLeague League, updatedAt int64) {
	// Check played 5 duels
	if !r.milestonePlayedFive && duelsPlayed >= 5 {
		r.milestonePlayedFive = true
		r.events = append(r.events, NewReferralMilestoneEvent(
			ChallengeID{value: r.id.value},
			r.inviterID,
			r.inviteeID,
			MilestonePlayedFive,
			updatedAt,
		))
	}

	// Check reached silver
	if !r.milestoneReachedSilver && currentLeague >= LeagueSilver {
		r.milestoneReachedSilver = true
		r.events = append(r.events, NewReferralMilestoneEvent(
			ChallengeID{value: r.id.value},
			r.inviterID,
			r.inviteeID,
			MilestoneReachedSilver,
			updatedAt,
		))
	}

	// Check reached gold
	if !r.milestoneReachedGold && currentLeague >= LeagueGold {
		r.milestoneReachedGold = true
		r.events = append(r.events, NewReferralMilestoneEvent(
			ChallengeID{value: r.id.value},
			r.inviterID,
			r.inviteeID,
			MilestoneReachedGold,
			updatedAt,
		))
	}

	// Check reached platinum
	if !r.milestoneReachedPlatinum && currentLeague >= LeaguePlatinum {
		r.milestoneReachedPlatinum = true
		r.events = append(r.events, NewReferralMilestoneEvent(
			ChallengeID{value: r.id.value},
			r.inviterID,
			r.inviteeID,
			MilestoneReachedPlatinum,
			updatedAt,
		))
	}
}

// ClaimInviterReward claims a milestone reward for the inviter
func (r *Referral) ClaimInviterReward(milestone string) (*ReferralReward, error) {
	if !r.IsMilestoneReached(milestone) {
		return nil, ErrMilestoneNotReached
	}

	if r.inviterRewardsClaimed[milestone] {
		return nil, ErrRewardAlreadyClaimed
	}

	rewards, exists := MilestoneRewards[milestone]
	if !exists {
		return nil, ErrMilestoneNotReached
	}

	r.inviterRewardsClaimed[milestone] = true
	return &rewards.Inviter, nil
}

// ClaimInviteeReward claims a milestone reward for the invitee
func (r *Referral) ClaimInviteeReward(milestone string) (*ReferralReward, error) {
	if !r.IsMilestoneReached(milestone) {
		return nil, ErrMilestoneNotReached
	}

	if r.inviteeRewardsClaimed[milestone] {
		return nil, ErrRewardAlreadyClaimed
	}

	rewards, exists := MilestoneRewards[milestone]
	if !exists {
		return nil, ErrMilestoneNotReached
	}

	r.inviteeRewardsClaimed[milestone] = true
	return &rewards.Invitee, nil
}

// IsMilestoneReached checks if a milestone has been reached
func (r *Referral) IsMilestoneReached(milestone string) bool {
	switch milestone {
	case MilestoneRegistered:
		return r.milestoneRegistered
	case MilestonePlayedFive:
		return r.milestonePlayedFive
	case MilestoneReachedSilver:
		return r.milestoneReachedSilver
	case MilestoneReachedGold:
		return r.milestoneReachedGold
	case MilestoneReachedPlatinum:
		return r.milestoneReachedPlatinum
	default:
		return false
	}
}

// GetPendingInviterRewards returns milestones that are reached but not claimed by inviter
func (r *Referral) GetPendingInviterRewards() []string {
	var pending []string
	milestones := []string{
		MilestoneRegistered,
		MilestonePlayedFive,
		MilestoneReachedSilver,
		MilestoneReachedGold,
		MilestoneReachedPlatinum,
	}

	for _, m := range milestones {
		if r.IsMilestoneReached(m) && !r.inviterRewardsClaimed[m] {
			pending = append(pending, m)
		}
	}

	return pending
}

// GetPendingInviteeRewards returns milestones that are reached but not claimed by invitee
func (r *Referral) GetPendingInviteeRewards() []string {
	var pending []string
	milestones := []string{
		MilestoneRegistered,
		MilestonePlayedFive,
		MilestoneReachedSilver,
		MilestoneReachedGold,
		MilestoneReachedPlatinum,
	}

	for _, m := range milestones {
		if r.IsMilestoneReached(m) && !r.inviteeRewardsClaimed[m] {
			pending = append(pending, m)
		}
	}

	return pending
}

// Getters
func (r *Referral) ID() ReferralID                   { return r.id }
func (r *Referral) InviterID() UserID                { return r.inviterID }
func (r *Referral) InviteeID() UserID                { return r.inviteeID }
func (r *Referral) MilestoneRegistered() bool        { return r.milestoneRegistered }
func (r *Referral) MilestonePlayedFive() bool        { return r.milestonePlayedFive }
func (r *Referral) MilestoneReachedSilver() bool     { return r.milestoneReachedSilver }
func (r *Referral) MilestoneReachedGold() bool       { return r.milestoneReachedGold }
func (r *Referral) MilestoneReachedPlatinum() bool   { return r.milestoneReachedPlatinum }
func (r *Referral) InviterRewardsClaimed() map[string]bool { return r.inviterRewardsClaimed }
func (r *Referral) InviteeRewardsClaimed() map[string]bool { return r.inviteeRewardsClaimed }
func (r *Referral) CreatedAt() int64                 { return r.createdAt }

// Events returns collected domain events and clears them
func (r *Referral) Events() []Event {
	events := r.events
	r.events = make([]Event, 0)
	return events
}
