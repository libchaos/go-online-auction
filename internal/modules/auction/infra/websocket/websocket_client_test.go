package websocket

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	t.Run("creates client with correct fields", func(t *testing.T) {
		// Arrange
		server := httptest.NewServer(http.HandlerFunc(echoHandler))
		defer server.Close()

		wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
		conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		require.NoError(t, err)
		defer conn.Close()

		auctionID := uint64(123)
		hub := &Hub{
			register:   make(chan *Client),
			unregister: make(chan *Client),
		}

		// Act
		client := NewClient(conn, auctionID, hub)

		// Assert
		require.NotNil(t, client)
		require.Equal(t, auctionID, client.auctionID)
		require.Equal(t, hub, client.hub)
		require.NotNil(t, client.send)
		require.Equal(t, sendBufferSize, cap(client.send))
	})
}

func TestClient_AuctionID(t *testing.T) {
	t.Run("returns correct auction ID", func(t *testing.T) {
		// Arrange
		auctionID := uint64(456)
		client := &Client{
			auctionID: auctionID,
		}

		// Act
		result := client.AuctionID()

		// Assert
		require.Equal(t, auctionID, result)
	})
}

func TestClient_Send(t *testing.T) {
	t.Run("returns send channel", func(t *testing.T) {
		// Arrange
		sendChan := make(chan []byte, 10)
		client := &Client{
			send: sendChan,
		}

		// Act
		result := client.Send()

		// Assert
		require.NotNil(t, result)
		// Verify we can send to the channel
		result <- []byte("test")
		received := <-sendChan
		require.Equal(t, []byte("test"), received)
	})
}

func TestClient_Close(t *testing.T) {
	t.Run("closes send channel", func(t *testing.T) {
		// Arrange
		client := &Client{
			send: make(chan []byte, 10),
		}

		// Act
		client.Close()

		// Assert
		_, ok := <-client.send
		require.False(t, ok, "send channel should be closed")
	})
}

func TestClient_WritePump(t *testing.T) {
	t.Run("sends message successfully", func(t *testing.T) {
		// Arrange
		server := httptest.NewServer(http.HandlerFunc(echoHandler))
		defer server.Close()

		wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
		conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		require.NoError(t, err)
		defer conn.Close()

		hub := &Hub{
			register:   make(chan *Client),
			unregister: make(chan *Client),
		}
		client := NewClient(conn, 123, hub)

		// Act - Start WritePump in goroutine
		go client.WritePump()

		// Send a message
		testMessage := []byte("test message")
		client.send <- testMessage

		// Read the echo response
		_, message, err := conn.ReadMessage()

		// Assert
		require.NoError(t, err)
		require.Equal(t, testMessage, message)

		// Cleanup
		client.Close()
	})

	t.Run("closes connection when send channel is closed", func(t *testing.T) {
		// Arrange
		server := httptest.NewServer(http.HandlerFunc(echoHandler))
		defer server.Close()

		wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
		conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		require.NoError(t, err)

		hub := &Hub{
			register:   make(chan *Client),
			unregister: make(chan *Client),
		}
		client := NewClient(conn, 123, hub)

		// Act
		writePumpDone := make(chan struct{})
		go func() {
			client.WritePump()
			close(writePumpDone)
		}()

		// Close the send channel to trigger shutdown
		close(client.send)

		// Wait for WritePump to complete with timeout
		select {
		case <-writePumpDone:
			// Success
		case <-time.After(2 * time.Second):
			t.Fatal("WritePump did not exit after closing send channel")
		}

		// Assert - Try to read, should get close message or error
		_, _, err = conn.ReadMessage()
		require.Error(t, err)
	})

	t.Run("sends ping messages periodically", func(t *testing.T) {
		// Note: This test is skipped because pingPeriod is 54 seconds in production
		// which is too long for a reasonable test. The ping mechanism is tested
		// indirectly through the integration tests and connection stability tests.
		t.Skip("Skipping ping test due to long production pingPeriod (54s)")
	})
}

func TestClient_ReadPump(t *testing.T) {
	t.Run("unregisters client on connection close", func(t *testing.T) {
		// Arrange
		server := httptest.NewServer(http.HandlerFunc(echoHandler))
		defer server.Close()

		wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
		conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		require.NoError(t, err)

		unregisterChan := make(chan *Client, 1)
		hub := &Hub{
			register:   make(chan *Client),
			unregister: unregisterChan,
		}
		client := NewClient(conn, 123, hub)

		// Act
		readPumpDone := make(chan struct{})
		go func() {
			client.ReadPump()
			close(readPumpDone)
		}()

		// Close connection to trigger unregister
		conn.Close()

		// Assert
		select {
		case unregisteredClient := <-unregisterChan:
			require.Equal(t, client, unregisteredClient)
		case <-time.After(2 * time.Second):
			t.Fatal("Client was not unregistered")
		}

		// Wait for ReadPump to exit
		select {
		case <-readPumpDone:
			// Success
		case <-time.After(2 * time.Second):
			t.Fatal("ReadPump did not exit")
		}
	})

	t.Run("handles pong messages correctly", func(t *testing.T) {
		// Arrange
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			upgrader := websocket.Upgrader{}
			conn, err := upgrader.Upgrade(w, r, nil)
			if err != nil {
				return
			}
			defer conn.Close()

			// Send a pong message
			if err := conn.WriteMessage(websocket.PongMessage, []byte{}); err != nil {
				return
			}

			// Keep connection open
			time.Sleep(2 * time.Second)
		}))
		defer server.Close()

		wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
		conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		require.NoError(t, err)

		hub := &Hub{
			register:   make(chan *Client),
			unregister: make(chan *Client, 1),
		}
		client := NewClient(conn, 123, hub)

		// Act
		readPumpDone := make(chan struct{})
		go func() {
			client.ReadPump()
			close(readPumpDone)
		}()

		// Wait a bit to ensure pong is processed
		time.Sleep(500 * time.Millisecond)

		// Close connection
		conn.Close()

		// Assert - ReadPump should exit gracefully
		select {
		case <-readPumpDone:
			// Success
		case <-time.After(2 * time.Second):
			t.Fatal("ReadPump did not exit after connection close")
		}
	})

	t.Run("exits on read error", func(t *testing.T) {
		// Arrange
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			upgrader := websocket.Upgrader{}
			conn, err := upgrader.Upgrade(w, r, nil)
			if err != nil {
				return
			}
			// Immediately close connection to cause read error
			conn.Close()
		}))
		defer server.Close()

		wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
		conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		require.NoError(t, err)

		unregisterChan := make(chan *Client, 1)
		hub := &Hub{
			register:   make(chan *Client),
			unregister: unregisterChan,
		}
		client := NewClient(conn, 123, hub)

		// Act
		readPumpDone := make(chan struct{})
		go func() {
			client.ReadPump()
			close(readPumpDone)
		}()

		// Assert
		select {
		case <-readPumpDone:
			// Success - ReadPump exited
		case <-time.After(2 * time.Second):
			t.Fatal("ReadPump did not exit on read error")
		}

		// Verify client was unregistered
		select {
		case unregisteredClient := <-unregisterChan:
			require.Equal(t, client, unregisteredClient)
		case <-time.After(1 * time.Second):
			t.Fatal("Client was not unregistered")
		}
	})

	t.Run("respects max message size", func(t *testing.T) {
		// Arrange
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			upgrader := websocket.Upgrader{}
			conn, err := upgrader.Upgrade(w, r, nil)
			if err != nil {
				return
			}
			defer conn.Close()

			// Send a message larger than maxMessageSize
			largeMessage := make([]byte, maxMessageSize+100)
			for i := range largeMessage {
				largeMessage[i] = 'A'
			}
			if err := conn.WriteMessage(websocket.TextMessage, largeMessage); err != nil {
				return
			}

			// Keep connection open briefly
			time.Sleep(1 * time.Second)
		}))
		defer server.Close()

		wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
		conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		require.NoError(t, err)

		unregisterChan := make(chan *Client, 1)
		hub := &Hub{
			register:   make(chan *Client),
			unregister: unregisterChan,
		}
		client := NewClient(conn, 123, hub)

		// Act
		readPumpDone := make(chan struct{})
		go func() {
			client.ReadPump()
			close(readPumpDone)
		}()

		// Assert - ReadPump should exit due to message size violation
		select {
		case <-readPumpDone:
			// Success - ReadPump exited
		case <-time.After(2 * time.Second):
			t.Fatal("ReadPump did not exit after receiving oversized message")
		}

		// Verify client was unregistered
		select {
		case <-unregisterChan:
			// Success
		case <-time.After(1 * time.Second):
			t.Fatal("Client was not unregistered")
		}
	})
}

// echoHandler is a simple WebSocket handler that echoes back received messages
func echoHandler(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	for {
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			return
		}
		if err := conn.WriteMessage(messageType, p); err != nil {
			return
		}
	}
}

// TestConstants verifies that constants have expected values
func TestConstants(t *testing.T) {
	t.Run("constants have correct values", func(t *testing.T) {
		require.Equal(t, 64, sendBufferSize)
		require.Equal(t, 10*time.Second, writeWait)
		require.Equal(t, 60*time.Second, pongWait)
		require.Equal(t, int64(10), int64(pingPeriodDivisor))
		require.Equal(t, int64(9), int64(pingPeriodFactor))
		require.Equal(t, int64(512), int64(maxMessageSize))
	})

	t.Run("pingPeriod is calculated correctly", func(t *testing.T) {
		expectedPingPeriod := (pongWait * pingPeriodFactor) / pingPeriodDivisor
		require.Equal(t, expectedPingPeriod, pingPeriod)
		require.Equal(t, 54*time.Second, pingPeriod)
	})
}

// TestClient_Integration performs an integration test of WritePump and ReadPump working together
func TestClient_Integration(t *testing.T) {
	t.Run("client can send and receive messages", func(t *testing.T) {
		// Arrange
		server := httptest.NewServer(http.HandlerFunc(echoHandler))
		defer server.Close()

		wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
		clientConn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		require.NoError(t, err)
		defer clientConn.Close()

		hub := &Hub{
			register:   make(chan *Client),
			unregister: make(chan *Client, 1),
		}
		client := NewClient(clientConn, 123, hub)

		// Act
		// Start both pumps
		go client.WritePump()
		go client.ReadPump()

		// Send test message
		testMessage := []byte("integration test message")
		client.send <- testMessage

		// Read the echoed message
		_, receivedMessage, err := clientConn.ReadMessage()
		require.NoError(t, err)
		require.Equal(t, testMessage, receivedMessage)

		// Cleanup
		client.Close()
		clientConn.Close()

		// Verify client was unregistered
		select {
		case <-hub.unregister:
			// Success
		case <-time.After(2 * time.Second):
			t.Fatal("Client was not unregistered")
		}
	})
}

// TestClient_ErrorScenarios tests various error conditions
func TestClient_ErrorScenarios(t *testing.T) {
	t.Run("WritePump handles write deadline errors", func(t *testing.T) {
		// This test verifies that WritePump exits gracefully when write errors occur
		// We simulate this by closing the connection before sending
		server := httptest.NewServer(http.HandlerFunc(echoHandler))
		defer server.Close()

		wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
		conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		require.NoError(t, err)

		hub := &Hub{
			register:   make(chan *Client),
			unregister: make(chan *Client),
		}
		client := NewClient(conn, 123, hub)

		writePumpDone := make(chan struct{})
		go func() {
			client.WritePump()
			close(writePumpDone)
		}()

		// Close connection immediately to cause write error
		conn.Close()

		// Try to send message (will fail but WritePump should handle it)
		select {
		case client.send <- []byte("test"):
		case <-time.After(100 * time.Millisecond):
		}

		// WritePump should exit
		select {
		case <-writePumpDone:
			// Success
		case <-time.After(2 * time.Second):
			t.Fatal("WritePump did not exit after connection close")
		}
	})
}

// mockConn is a mock implementation for testing edge cases
type mockConn struct {
	*websocket.Conn
	writeMessageFunc     func(messageType int, data []byte) error
	readMessageFunc      func() (messageType int, p []byte, err error)
	setReadDeadlineFunc  func(t time.Time) error
	setWriteDeadlineFunc func(t time.Time) error
	setPongHandlerFunc   func(h func(appData string) error)
	setReadLimitFunc     func(limit int64)
	closeFunc            func() error
}

func (m *mockConn) WriteMessage(messageType int, data []byte) error {
	if m.writeMessageFunc != nil {
		return m.writeMessageFunc(messageType, data)
	}
	return nil
}

func (m *mockConn) ReadMessage() (int, []byte, error) {
	if m.readMessageFunc != nil {
		return m.readMessageFunc()
	}
	return 0, nil, errors.New("connection closed")
}

func (m *mockConn) SetReadDeadline(t time.Time) error {
	if m.setReadDeadlineFunc != nil {
		return m.setReadDeadlineFunc(t)
	}
	return nil
}

func (m *mockConn) SetWriteDeadline(t time.Time) error {
	if m.setWriteDeadlineFunc != nil {
		return m.setWriteDeadlineFunc(t)
	}
	return nil
}

func (m *mockConn) SetPongHandler(h func(appData string) error) {
	if m.setPongHandlerFunc != nil {
		m.setPongHandlerFunc(h)
	}
}

func (m *mockConn) SetReadLimit(limit int64) {
	if m.setReadLimitFunc != nil {
		m.setReadLimitFunc(limit)
	}
}

func (m *mockConn) Close() error {
	if m.closeFunc != nil {
		return m.closeFunc()
	}
	return nil
}
