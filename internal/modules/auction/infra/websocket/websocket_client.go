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
	conn      *websocket.Conn
	auctionID uint64
	send      chan []byte
	hub       *Hub
}

func NewClient(conn *websocket.Conn, auctionID uint64, hub *Hub) *Client {
	return &Client{
		conn:      conn,
		auctionID: auctionID,
		send:      make(chan []byte, sendBufferSize),
		hub:       hub,
	}
}

func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		_ = c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}

		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Client) ReadPump() {
	defer func() {
		c.hub.unregister <- c
		_ = c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

func (c *Client) Close() {
	close(c.send)
}

func (c *Client) AuctionID() uint64 {
	return c.auctionID
}

func (c *Client) Send() chan<- []byte {
	return c.send
}
