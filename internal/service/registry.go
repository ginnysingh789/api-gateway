package service

import (
	"errors"
	"sync"

	"api-gateway/internal/config"
)

type Service struct {
	Name      string
	URLs      []string
	HealthURL string
	Active    bool
}

type Registry struct {
	services map[string]*Service
	mu       sync.RWMutex
}

func NewRegistry(services []config.ServiceConfig) *Registry {
	r := &Registry{
		services: make(map[string]*Service),
	}

	// Register services from config
	for _, svc := range services {
		r.Register(svc.Name, svc.URLs, svc.HealthURL)
	}

	return r
}

func (r *Registry) Register(name string, urls []string, healthURL string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.services[name] = &Service{
		Name:      name,
		URLs:      urls,
		HealthURL: healthURL,
		Active:    true,
	}
}

func (r *Registry) Unregister(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.services[name]; !exists {
		return errors.New("service not found")
	}

	delete(r.services, name)
	return nil
}

func (r *Registry) Get(name string) (*Service, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	svc, exists := r.services[name]
	if !exists {
		return nil, errors.New("service not found")
	}

	if !svc.Active {
		return nil, errors.New("service is inactive")
	}

	return svc, nil
}

func (r *Registry) List() []*Service {
	r.mu.RLock()
	defer r.mu.RUnlock()

	services := make([]*Service, 0, len(r.services))
	for _, svc := range r.services {
		services = append(services, svc)
	}

	return services
}

func (r *Registry) SetActive(name string, active bool) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	svc, exists := r.services[name]
	if !exists {
		return errors.New("service not found")
	}

	svc.Active = active
	return nil
}
