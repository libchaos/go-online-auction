package websocket

import "sync"

type UserSubscriberRegistry struct {
	mu      sync.RWMutex
	clients map[uint64]map[*Client]struct{}
}

func NewUserSubscriberRegistry() *UserSubscriberRegistry {
	return &UserSubscriberRegistry{
		clients: make(map[uint64]map[*Client]struct{}),
	}
}

func (registry *UserSubscriberRegistry) Add(userID uint64, client *Client) {
	registry.mu.Lock()
	defer registry.mu.Unlock()

	if registry.clients[userID] == nil {
		registry.clients[userID] = make(map[*Client]struct{})
	}

	registry.clients[userID][client] = struct{}{}
}

func (registry *UserSubscriberRegistry) Remove(userID uint64, client *Client) {
	registry.mu.Lock()
	defer registry.mu.Unlock()

	if clients, ok := registry.clients[userID]; ok {
		delete(clients, client)
		if len(clients) == 0 {
			delete(registry.clients, userID)
		}
	}
}

func (registry *UserSubscriberRegistry) PublishToUser(userID uint64, message []byte) {
	registry.mu.RLock()
	defer registry.mu.RUnlock()

	clients := registry.clients[userID]
	for client := range clients {
		select {
		case client.send <- message:
		default:
		}
	}
}
