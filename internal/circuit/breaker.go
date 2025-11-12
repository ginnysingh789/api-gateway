package circuit

import (
	"sync"
	"time"

	"api-gateway/internal/config"

	"github.com/sony/gobreaker"
)

type BreakerManager struct {
	breakers map[string]*gobreaker.CircuitBreaker
	config   config.CircuitBreakerConfig
	mu       sync.RWMutex
}

func NewBreakerManager(cfg config.CircuitBreakerConfig) *BreakerManager {
	return &BreakerManager{
		breakers: make(map[string]*gobreaker.CircuitBreaker),
		config:   cfg,
	}
}

func (bm *BreakerManager) GetBreaker(serviceName string) *gobreaker.CircuitBreaker {
	bm.mu.RLock()
	breaker, exists := bm.breakers[serviceName]
	bm.mu.RUnlock()

	if exists {
		return breaker
	}

	// Create new breaker
	bm.mu.Lock()
	defer bm.mu.Unlock()

	// Double-check after acquiring write lock
	if breaker, exists := bm.breakers[serviceName]; exists {
		return breaker
	}

	breaker = gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:        serviceName,
		MaxRequests: 3,
		Interval:    time.Minute,
		Timeout:     bm.config.Timeout,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures >= uint32(bm.config.Threshold)
		},
		OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
			// Log state changes
		},
	})

	bm.breakers[serviceName] = breaker
	return breaker
}
