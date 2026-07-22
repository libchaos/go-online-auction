package sse

import "sync"

// SubscriberRegistry maps each user id to the set of that user's connected SSE
// clients so an event can be fanned out to every open tab or device.
type SubscriberRegistry struct {
	mu      sync.RWMutex
	clients map[uint64]map[*Client]struct{}
}

func NewSubscriberRegistry() *SubscriberRegistry {
	return &SubscriberRegistry{
		clients: make(map[uint64]map[*Client]struct{}),
	}
}

func (registry *SubscriberRegistry) Add(userID uint64, client *Client) {
	registry.mu.Lock()
	defer registry.mu.Unlock()

	if registry.clients[userID] == nil {
		registry.clients[userID] = make(map[*Client]struct{})
	}

	registry.clients[userID][client] = struct{}{}
}

func (registry *SubscriberRegistry) Remove(userID uint64, client *Client) {
	registry.mu.Lock()
	defer registry.mu.Unlock()

	if clients, ok := registry.clients[userID]; ok {
		delete(clients, client)
		if len(clients) == 0 {
			delete(registry.clients, userID)
		}
	}
}

func (registry *SubscriberRegistry) PublishToUser(userID uint64, message []byte) {
	registry.mu.RLock()
	defer registry.mu.RUnlock()

	for client := range registry.clients[userID] {
		client.enqueue(message)
	}
}
