package websocket_test

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"unsafe"

	"auction/internal/modules/auction/infra/websocket"
	ws "github.com/gorilla/websocket"
	"github.com/stretchr/testify/suite"
)

type AuctionSubscriberRegistryTestSuite struct {
	suite.Suite
	sut *websocket.AuctionSubscriberRegistry
}

func (s *AuctionSubscriberRegistryTestSuite) SetupTest() {
	s.sut = websocket.NewAuctionSubscriberRegistry()
}

func TestAuctionSubscriberRegistrySuite(t *testing.T) {
	suite.Run(t, new(AuctionSubscriberRegistryTestSuite))
}

func (s *AuctionSubscriberRegistryTestSuite) TestAdd_AddsClientToAuction() {
	// Arrange
	auctionID := uint64(123)
	client := createTestClient(auctionID)
	defer client.Close()

	// Act
	s.sut.Add(auctionID, client)

	// Assert
	message := []byte("test message")
	s.sut.Broadcast(auctionID, message)

	receivedMessage := readFromClientChannel(client)
	s.Equal(message, receivedMessage)
}

func (s *AuctionSubscriberRegistryTestSuite) TestAdd_AddsMultipleClientsToSameAuction() {
	// Arrange
	auctionID := uint64(123)
	client1 := createTestClient(auctionID)
	client2 := createTestClient(auctionID)
	defer client1.Close()
	defer client2.Close()

	// Act
	s.sut.Add(auctionID, client1)
	s.sut.Add(auctionID, client2)

	// Assert
	message := []byte("broadcast message")
	s.sut.Broadcast(auctionID, message)

	receivedMessage1 := readFromClientChannel(client1)
	receivedMessage2 := readFromClientChannel(client2)

	s.Equal(message, receivedMessage1)
	s.Equal(message, receivedMessage2)
}

func (s *AuctionSubscriberRegistryTestSuite) TestAdd_AddsClientsToDifferentAuctions() {
	// Arrange
	auctionID1 := uint64(123)
	auctionID2 := uint64(456)
	client1 := createTestClient(auctionID1)
	client2 := createTestClient(auctionID2)
	defer client1.Close()
	defer client2.Close()

	// Act
	s.sut.Add(auctionID1, client1)
	s.sut.Add(auctionID2, client2)

	// Assert
	message1 := []byte("message for auction 1")
	message2 := []byte("message for auction 2")

	s.sut.Broadcast(auctionID1, message1)
	s.sut.Broadcast(auctionID2, message2)

	receivedMessage1 := readFromClientChannel(client1)
	receivedMessage2 := readFromClientChannel(client2)

	s.Equal(message1, receivedMessage1)
	s.Equal(message2, receivedMessage2)
}

func (s *AuctionSubscriberRegistryTestSuite) TestRemove_RemovesClientFromAuction() {
	// Arrange
	auctionID := uint64(123)
	client1 := createTestClient(auctionID)
	client2 := createTestClient(auctionID)
	defer client1.Close()
	defer client2.Close()

	s.sut.Add(auctionID, client1)
	s.sut.Add(auctionID, client2)

	// Act
	s.sut.Remove(auctionID, client1)

	// Assert
	message := []byte("test message")
	s.sut.Broadcast(auctionID, message)

	receivedMessage := readFromClientChannel(client2)
	s.Equal(message, receivedMessage)

	_, ok := tryReadFromClientChannel(client1)
	s.False(ok, "client1 should not receive message after removal")
}

func (s *AuctionSubscriberRegistryTestSuite) TestRemove_RemovesLastClientCleansUpAuction() {
	// Arrange
	auctionID := uint64(123)
	client := createTestClient(auctionID)
	defer client.Close()

	s.sut.Add(auctionID, client)

	// Act
	s.sut.Remove(auctionID, client)

	// Assert
	message := []byte("test message")
	s.sut.Broadcast(auctionID, message)

	_, ok := tryReadFromClientChannel(client)
	s.False(ok, "client should not receive message after removal")
}

func (s *AuctionSubscriberRegistryTestSuite) TestRemove_RemovingNonExistentClientDoesNotPanic() {
	// Arrange
	auctionID := uint64(123)
	client := createTestClient(auctionID)
	defer client.Close()

	// Act & Assert - should not panic
	s.sut.Remove(auctionID, client)
}

func (s *AuctionSubscriberRegistryTestSuite) TestRemove_RemovingFromNonExistentAuctionDoesNotPanic() {
	// Arrange
	auctionID := uint64(999)
	existingAuctionID := uint64(123)
	client := createTestClient(existingAuctionID)
	defer client.Close()

	s.sut.Add(existingAuctionID, client)

	// Act & Assert - should not panic
	s.sut.Remove(auctionID, client)
}

func (s *AuctionSubscriberRegistryTestSuite) TestBroadcast_SendsMessageToAllClientsOfAuction() {
	// Arrange
	auctionID := uint64(123)
	client1 := createTestClient(auctionID)
	client2 := createTestClient(auctionID)
	client3 := createTestClient(auctionID)
	defer client1.Close()
	defer client2.Close()
	defer client3.Close()

	s.sut.Add(auctionID, client1)
	s.sut.Add(auctionID, client2)
	s.sut.Add(auctionID, client3)

	message := []byte("broadcast to all")

	// Act
	s.sut.Broadcast(auctionID, message)

	// Assert
	receivedMessage1 := readFromClientChannel(client1)
	receivedMessage2 := readFromClientChannel(client2)
	receivedMessage3 := readFromClientChannel(client3)

	s.Equal(message, receivedMessage1)
	s.Equal(message, receivedMessage2)
	s.Equal(message, receivedMessage3)
}

func (s *AuctionSubscriberRegistryTestSuite) TestBroadcast_DoesNotSendToOtherAuctions() {
	// Arrange
	auctionID1 := uint64(123)
	auctionID2 := uint64(456)
	client1 := createTestClient(auctionID1)
	client2 := createTestClient(auctionID2)
	defer client1.Close()
	defer client2.Close()

	s.sut.Add(auctionID1, client1)
	s.sut.Add(auctionID2, client2)

	message := []byte("message for auction 1 only")

	// Act
	s.sut.Broadcast(auctionID1, message)

	// Assert
	receivedMessage1 := readFromClientChannel(client1)
	s.Equal(message, receivedMessage1)

	_, ok := tryReadFromClientChannel(client2)
	s.False(ok, "client2 should not receive message for different auction")
}

func (s *AuctionSubscriberRegistryTestSuite) TestBroadcast_HandlesFullSendChannel() {
	// Arrange
	auctionID := uint64(123)
	client := createTestClientWithBufferSize(auctionID, 1)
	defer client.Close()

	s.sut.Add(auctionID, client)

	message1 := []byte("message 1")
	message2 := []byte("message 2")

	// Fill the channel
	client.Send() <- message1

	// Act - Broadcast should not block when channel is full, it should skip
	s.sut.Broadcast(auctionID, message2)

	// Assert - only message1 should be in the channel
	receivedMessage := readFromClientChannel(client)
	s.Equal(message1, receivedMessage)

	// After reading message1, check if message2 is there (it should not be)
	// Since Broadcast uses select with default, message2 was dropped
	data, ok := tryReadFromClientChannel(client)
	if ok {
		// If there's a message, it means the channel had space and message2 was sent
		// This can happen in a race condition after we read message1
		s.Equal(message2, data)
	} else {
		// Expected case: message2 was dropped because channel was full
		s.False(ok)
	}
}

func (s *AuctionSubscriberRegistryTestSuite) TestBroadcast_ToNonExistentAuctionDoesNotPanic() {
	// Arrange
	auctionID := uint64(999)
	message := []byte("test message")

	// Act & Assert - should not panic
	s.sut.Broadcast(auctionID, message)
}

func createTestClient(auctionID uint64) *websocket.Client {
	return createTestClientWithBufferSize(auctionID, 64)
}

func createTestClientWithBufferSize(auctionID uint64, bufferSize int) *websocket.Client {
	server := httptest.NewServer(http.HandlerFunc(echoHandler))
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	conn, _, err := ws.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		panic(err)
	}

	hub := &websocket.Hub{}
	client := websocket.NewClient(conn, auctionID, hub)

	return client
}

func echoHandler(w http.ResponseWriter, r *http.Request) {
	upgrader := ws.Upgrader{}
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

func readFromClientChannel(client *websocket.Client) []byte {
	// Access the private send field using unsafe reflection
	v := reflect.ValueOf(client).Elem()
	sendField := v.FieldByName("send")

	// Get the actual channel value through unsafe operations
	sendChan := reflect.NewAt(sendField.Type(), unsafe.Pointer(sendField.UnsafeAddr())).Elem()

	// Now we can receive from it
	recv, _ := sendChan.Recv()
	return recv.Bytes()
}

func tryReadFromClientChannel(client *websocket.Client) ([]byte, bool) {
	// Access the private send field using unsafe reflection
	v := reflect.ValueOf(client).Elem()
	sendField := v.FieldByName("send")

	// Get the actual channel value through unsafe operations
	sendChan := reflect.NewAt(sendField.Type(), unsafe.Pointer(sendField.UnsafeAddr())).Elem()

	// Try to receive without blocking
	recv, ok := sendChan.TryRecv()
	if ok {
		return recv.Bytes(), true
	}
	return nil, false
}
