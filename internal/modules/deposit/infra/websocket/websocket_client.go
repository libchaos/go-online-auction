package websocket

import (
	"time"

	"github.com/gorilla/websocket"
)

const (
	sendBufferSize    = 64
	writeWait         = 10 * time.Second
	pongWait          = 60 * time.Second
	pingPeriodDivisor = 10
	pingPeriodFactor  = 9
	maxMessageSize    = 512
)

var pingPeriod = (pongWait * pingPeriodFactor) / pingPeriodDivisor

type Client struct {
	conn   *websocket.Conn
	userID uint64
	send   chan []byte
	hub    *Hub
}

func NewClient(conn *websocket.Conn, userID uint64, hub *Hub) *Client {
	return &Client{
		conn:   conn,
		userID: userID,
		send:   make(chan []byte, sendBufferSize),
		hub:    hub,
	}
}

func (client *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		_ = client.conn.Close()
	}()

	for {
		select {
		case message, ok := <-client.send:
			_ = client.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				_ = client.conn.WriteMessage(websocket.CloseMessage, []byte{})

				return
			}

			if err := client.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}

		case <-ticker.C:
			_ = client.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := client.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (client *Client) ReadPump() {
	defer func() {
		client.hub.unregister <- client
		_ = client.conn.Close()
	}()

	client.conn.SetReadLimit(maxMessageSize)
	_ = client.conn.SetReadDeadline(time.Now().Add(pongWait))
	client.conn.SetPongHandler(func(string) error {
		_ = client.conn.SetReadDeadline(time.Now().Add(pongWait))

		return nil
	})

	for {
		_, _, err := client.conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

func (client *Client) Close() {
	close(client.send)
}

func (client *Client) UserID() uint64 {
	return client.userID
}

func (client *Client) Send() chan<- []byte {
	return client.send
}
