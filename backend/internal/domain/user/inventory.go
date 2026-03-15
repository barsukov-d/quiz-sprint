package user

import "fmt"

const (
	ResourceCoins      = "coins"
	ResourcePvpTickets = "pvp_tickets"
	ResourceShield     = "shield"
	ResourceFiftyFifty = "fifty_fifty"
	ResourceSkip       = "skip"
	ResourceFreeze     = "freeze"
)

var validResources = map[string]bool{
	ResourceCoins:      true,
	ResourcePvpTickets: true,
	ResourceShield:     true,
	ResourceFiftyFifty: true,
	ResourceSkip:       true,
	ResourceFreeze:     true,
}

type Inventory struct {
	playerID   UserID
	coins      int
	pvpTickets int
	shield     int
	fiftyFifty int
	skip       int
	freeze     int
	updatedAt  int64
}

func NewInventory(playerID UserID, createdAt int64) (*Inventory, error) {
	if playerID.IsZero() {
		return nil, ErrInvalidUserID
	}

	return &Inventory{
		playerID:   playerID,
		pvpTickets: 3,
		updatedAt:  createdAt,
	}, nil
}

func ReconstructInventory(
	playerID UserID,
	coins int,
	pvpTickets int,
	shield int,
	fiftyFifty int,
	skip int,
	freeze int,
	updatedAt int64,
) *Inventory {
	return &Inventory{
		playerID:   playerID,
		coins:      coins,
		pvpTickets: pvpTickets,
		shield:     shield,
		fiftyFifty: fiftyFifty,
		skip:       skip,
		freeze:     freeze,
		updatedAt:  updatedAt,
	}
}

func (i *Inventory) Credit(resource string, amount int, updatedAt int64) error {
	if amount <= 0 {
		return ErrInvalidAmount
	}
	if !validResources[resource] {
		return fmt.Errorf("%w: %s", ErrInvalidResource, resource)
	}

	i.setResource(resource, i.getResource(resource)+amount)
	i.updatedAt = updatedAt
	return nil
}

func (i *Inventory) Debit(resource string, amount int, updatedAt int64) error {
	if amount <= 0 {
		return ErrInvalidAmount
	}
	if !validResources[resource] {
		return fmt.Errorf("%w: %s", ErrInvalidResource, resource)
	}
	if i.getResource(resource) < amount {
		return fmt.Errorf("%w: %s (have %d, need %d)", ErrInsufficientBalance, resource, i.getResource(resource), amount)
	}

	i.setResource(resource, i.getResource(resource)-amount)
	i.updatedAt = updatedAt
	return nil
}

func (i *Inventory) CreditMultiple(credits map[string]int, updatedAt int64) error {
	for resource, amount := range credits {
		if amount <= 0 {
			return ErrInvalidAmount
		}
		if !validResources[resource] {
			return fmt.Errorf("%w: %s", ErrInvalidResource, resource)
		}
	}

	for resource, amount := range credits {
		i.setResource(resource, i.getResource(resource)+amount)
	}
	i.updatedAt = updatedAt
	return nil
}

func (i *Inventory) DebitMultiple(debits map[string]int, updatedAt int64) error {
	for resource, amount := range debits {
		if amount <= 0 {
			return ErrInvalidAmount
		}
		if !validResources[resource] {
			return fmt.Errorf("%w: %s", ErrInvalidResource, resource)
		}
		if i.getResource(resource) < amount {
			return fmt.Errorf("%w: %s (have %d, need %d)", ErrInsufficientBalance, resource, i.getResource(resource), amount)
		}
	}

	for resource, amount := range debits {
		i.setResource(resource, i.getResource(resource)-amount)
	}
	i.updatedAt = updatedAt
	return nil
}

func (i *Inventory) getResource(resource string) int {
	switch resource {
	case ResourceCoins:
		return i.coins
	case ResourcePvpTickets:
		return i.pvpTickets
	case ResourceShield:
		return i.shield
	case ResourceFiftyFifty:
		return i.fiftyFifty
	case ResourceSkip:
		return i.skip
	case ResourceFreeze:
		return i.freeze
	default:
		return 0
	}
}

func (i *Inventory) setResource(resource string, value int) {
	switch resource {
	case ResourceCoins:
		i.coins = value
	case ResourcePvpTickets:
		i.pvpTickets = value
	case ResourceShield:
		i.shield = value
	case ResourceFiftyFifty:
		i.fiftyFifty = value
	case ResourceSkip:
		i.skip = value
	case ResourceFreeze:
		i.freeze = value
	}
}

func (i *Inventory) PlayerID() UserID { return i.playerID }
func (i *Inventory) Coins() int       { return i.coins }
func (i *Inventory) PvpTickets() int  { return i.pvpTickets }
func (i *Inventory) Shield() int      { return i.shield }
func (i *Inventory) FiftyFifty() int  { return i.fiftyFifty }
func (i *Inventory) Skip() int        { return i.skip }
func (i *Inventory) Freeze() int      { return i.freeze }
func (i *Inventory) UpdatedAt() int64 { return i.updatedAt }
