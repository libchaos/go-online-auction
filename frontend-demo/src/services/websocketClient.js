import ReconnectingWebSocket from 'reconnecting-websocket';

/**
 * Create a WebSocket connection for an auction
 * @param {number} auctionId - Auction ID to subscribe to
 * @param {Object} callbacks - Event callbacks
 * @param {Function} callbacks.onOpen - Called when connection opens
 * @param {Function} callbacks.onMessage - Called when message received
 * @param {Function} callbacks.onClose - Called when connection closes
 * @param {Function} callbacks.onError - Called on error
 * @returns {ReconnectingWebSocket} WebSocket instance
 */
export const createAuctionWebSocket = (auctionId, callbacks = {}) => {
  const wsBaseUrl = (import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080')
    .replace('http://', 'ws://')
    .replace('https://', 'wss://')
    .replace('/api/v1', '');
  
  const wsUrl = `${wsBaseUrl}/ws/v1/auctions/${auctionId}`;
  
  const ws = new ReconnectingWebSocket(wsUrl, [], {
    maxRetries: Infinity,
    maxReconnectionDelay: 30000,
    minReconnectionDelay: 1000,
  });

  ws.addEventListener('open', () => {
    console.log(`WebSocket connected to auction ${auctionId}`);
    callbacks.onOpen?.();
  });

  ws.addEventListener('message', (event) => {
    try {
      const data = JSON.parse(event.data);
      callbacks.onMessage?.(data);
    } catch (err) {
      console.error('WebSocket message parse error:', err);
      callbacks.onError?.(err);
    }
  });

  ws.addEventListener('close', () => {
    console.log(`WebSocket disconnected from auction ${auctionId}`);
    callbacks.onClose?.();
  });

  ws.addEventListener('error', (err) => {
    console.error('WebSocket error:', err);
    callbacks.onError?.(err);
  });

  return ws;
};
