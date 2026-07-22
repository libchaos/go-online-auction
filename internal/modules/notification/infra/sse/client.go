package sse

import "sync"

const clientSendBuffer = 64

// Client is a single connected Server-Sent-Events subscriber. The HTTP handler
// owns the network connection and drains Send in its write loop; the hub only
// enqueues messages and signals closure via Done.
type Client struct {
	userID    uint64
	send      chan []byte
	done      chan struct{}
	closeOnce sync.Once
}

func NewClient(userID uint64) *Client {
	return &Client{
		userID: userID,
		send:   make(chan []byte, clientSendBuffer),
		done:   make(chan struct{}),
	}
}

func (client *Client) UserID() uint64 {
	return client.userID
}

func (client *Client) Send() <-chan []byte {
	return client.send
}

func (client *Client) Done() <-chan struct{} {
	return client.done
}

// enqueue performs a non-blocking push. A slow subscriber whose buffer is full
// simply drops the event rather than stalling the fan-out for every other user.
func (client *Client) enqueue(message []byte) bool {
	select {
	case client.send <- message:
		return true
	default:
		return false
	}
}

func (client *Client) Close() {
	client.closeOnce.Do(func() {
		close(client.done)
	})
}
