package service

import (
	"errors"
	"sync"
)

type LoadBalancer struct {
	counters map[string]int
	mu       sync.Mutex
}

func NewLoadBalancer() *LoadBalancer {
	return &LoadBalancer{
		counters: make(map[string]int),
	}
}

// RoundRobin returns the next URL using round-robin algorithm
func (lb *LoadBalancer) RoundRobin(service *Service) (string, error) {
	if len(service.URLs) == 0 {
		return "", errors.New("no URLs available for service")
	}

	lb.mu.Lock()
	defer lb.mu.Unlock()

	counter := lb.counters[service.Name]
	url := service.URLs[counter%len(service.URLs)]
	lb.counters[service.Name] = (counter + 1) % len(service.URLs)

	return url, nil
}
