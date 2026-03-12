package quick_duel

import "sync"

// MemoryRoundCache is an in-process implementation of DuelRoundCache.
// Use this when Redis is unavailable (single-instance dev mode) or in tests.
type MemoryRoundCache struct {
	mu   sync.Mutex
	data map[string]map[int][]PlayerAnswer // gameID -> round -> answers
}

// NewMemoryRoundCache creates a new in-memory round cache.
func NewMemoryRoundCache() *MemoryRoundCache {
	return &MemoryRoundCache{
		data: make(map[string]map[int][]PlayerAnswer),
	}
}

func (c *MemoryRoundCache) AddAnswer(gameID string, round int, answer PlayerAnswer) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.data[gameID] == nil {
		c.data[gameID] = make(map[int][]PlayerAnswer)
	}
	c.data[gameID][round] = append(c.data[gameID][round], answer)
	return nil
}

func (c *MemoryRoundCache) GetAnswers(gameID string, round int) ([]PlayerAnswer, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if rounds, ok := c.data[gameID]; ok {
		if answers, ok := rounds[round]; ok {
			// Return a copy to avoid races on the caller side
			out := make([]PlayerAnswer, len(answers))
			copy(out, answers)
			return out, nil
		}
	}
	return []PlayerAnswer{}, nil
}

func (c *MemoryRoundCache) DeleteGame(gameID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.data, gameID)
	return nil
}
