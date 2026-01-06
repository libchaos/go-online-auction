package websocket

import (
	"sync"
)

type AuctionSubscriberRegistry struct {
	mu      sync.RWMutex
	clients map[uint64]map[*Client]struct{}
}

func NewAuctionSubscriberRegistry() *AuctionSubscriberRegistry {
	return &AuctionSubscriberRegistry{
		clients: make(map[uint64]map[*Client]struct{}),
	}
}

func (r *AuctionSubscriberRegistry) Add(auctionID uint64, client *Client) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.clients[auctionID] == nil {
		r.clients[auctionID] = make(map[*Client]struct{})
	}
	r.clients[auctionID][client] = struct{}{}
}

func (r *AuctionSubscriberRegistry) Remove(auctionID uint64, client *Client) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if clients, ok := r.clients[auctionID]; ok {
		delete(clients, client)
		if len(clients) == 0 {
			delete(r.clients, auctionID)
		}
	}
}

func (r *AuctionSubscriberRegistry) Broadcast(auctionID uint64, message []byte) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	clients := r.clients[auctionID]
	for client := range clients {
		select {
		case client.send <- message:
		default:
		}
	}
}
