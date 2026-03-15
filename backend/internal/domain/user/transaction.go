package user

type TransactionType string

const (
	TransactionCredit TransactionType = "credit"
	TransactionDebit  TransactionType = "debit"
)

type TransactionLog struct {
	id        string
	playerID  UserID
	txType    TransactionType
	source    string
	details   map[string]int
	createdAt int64
}

func NewTransactionLog(
	id string,
	playerID UserID,
	txType TransactionType,
	source string,
	details map[string]int,
	createdAt int64,
) (*TransactionLog, error) {
	if id == "" {
		return nil, ErrInvalidTransactionID
	}
	if playerID.IsZero() {
		return nil, ErrInvalidUserID
	}
	if txType != TransactionCredit && txType != TransactionDebit {
		return nil, ErrInvalidTransactionType
	}
	if source == "" {
		return nil, ErrInvalidTransactionSource
	}
	if len(details) == 0 {
		return nil, ErrEmptyTransactionDetails
	}

	detailsCopy := make(map[string]int, len(details))
	for k, v := range details {
		detailsCopy[k] = v
	}

	return &TransactionLog{
		id:        id,
		playerID:  playerID,
		txType:    txType,
		source:    source,
		details:   detailsCopy,
		createdAt: createdAt,
	}, nil
}

func ReconstructTransactionLog(
	id string,
	playerID UserID,
	txType TransactionType,
	source string,
	details map[string]int,
	createdAt int64,
) *TransactionLog {
	return &TransactionLog{
		id:        id,
		playerID:  playerID,
		txType:    txType,
		source:    source,
		details:   details,
		createdAt: createdAt,
	}
}

func (t *TransactionLog) ID() string              { return t.id }
func (t *TransactionLog) PlayerID() UserID         { return t.playerID }
func (t *TransactionLog) Type() TransactionType    { return t.txType }
func (t *TransactionLog) Source() string            { return t.source }
func (t *TransactionLog) CreatedAt() int64          { return t.createdAt }

func (t *TransactionLog) Details() map[string]int {
	copy := make(map[string]int, len(t.details))
	for k, v := range t.details {
		copy[k] = v
	}
	return copy
}
