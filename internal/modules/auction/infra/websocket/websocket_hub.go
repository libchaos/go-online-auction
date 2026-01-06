package websocket

type Hub struct {
	unregister chan *Client
}
