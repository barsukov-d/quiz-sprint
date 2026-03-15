package user

import (
	"time"

	"github.com/google/uuid"

	"github.com/barsukov/quiz-sprint/backend/internal/domain/user"
)

// InventoryService orchestrates inventory and transaction operations
type InventoryService interface {
	GetBalance(playerID string) (*InventoryDTO, error)
	Credit(playerID string, source string, details map[string]int) error
	Debit(playerID string, source string, details map[string]int) error
}

type inventoryServiceImpl struct {
	inventoryRepo user.InventoryRepository
	txRepo        user.TransactionRepository
}

// NewInventoryService creates a new InventoryService
func NewInventoryService(inventoryRepo user.InventoryRepository, txRepo user.TransactionRepository) InventoryService {
	return &inventoryServiceImpl{
		inventoryRepo: inventoryRepo,
		txRepo:        txRepo,
	}
}

// GetBalance returns the current inventory for a player
func (s *inventoryServiceImpl) GetBalance(playerID string) (*InventoryDTO, error) {
	uid, err := user.NewUserID(playerID)
	if err != nil {
		return nil, err
	}

	inventory, err := s.inventoryRepo.FindByPlayerID(uid)
	if err != nil {
		return nil, err
	}

	return &InventoryDTO{
		Coins:      inventory.Coins(),
		PvpTickets: inventory.PvpTickets(),
		Shield:     inventory.Shield(),
		FiftyFifty: inventory.FiftyFifty(),
		Skip:       inventory.Skip(),
		Freeze:     inventory.Freeze(),
	}, nil
}

// Credit adds resources to a player's inventory and logs the transaction
func (s *inventoryServiceImpl) Credit(playerID string, source string, details map[string]int) error {
	uid, err := user.NewUserID(playerID)
	if err != nil {
		return err
	}

	inventory, err := s.inventoryRepo.FindByPlayerID(uid)
	if err != nil {
		return err
	}

	now := time.Now().Unix()

	if err := inventory.CreditMultiple(details, now); err != nil {
		return err
	}

	if err := s.inventoryRepo.Save(inventory); err != nil {
		return err
	}

	txLog, err := user.NewTransactionLog(uuid.New().String(), uid, user.TransactionCredit, source, details, now)
	if err != nil {
		return err
	}

	return s.txRepo.Save(txLog)
}

// Debit removes resources from a player's inventory and logs the transaction
func (s *inventoryServiceImpl) Debit(playerID string, source string, details map[string]int) error {
	uid, err := user.NewUserID(playerID)
	if err != nil {
		return err
	}

	inventory, err := s.inventoryRepo.FindByPlayerID(uid)
	if err != nil {
		return err
	}

	now := time.Now().Unix()

	if err := inventory.DebitMultiple(details, now); err != nil {
		return err
	}

	if err := s.inventoryRepo.Save(inventory); err != nil {
		return err
	}

	txLog, err := user.NewTransactionLog(uuid.New().String(), uid, user.TransactionDebit, source, details, now)
	if err != nil {
		return err
	}

	return s.txRepo.Save(txLog)
}
