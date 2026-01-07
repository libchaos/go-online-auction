import { useState, useEffect, useRef } from 'react';
import { useParams, Link } from 'react-router-dom';
import { createAuctionWebSocket } from '../services/websocketClient';
import EventItem from '../components/EventItem';

/**
 * WebSocketPage - Live event feed for an auction
 */
function WebSocketPage() {
  const { id } = useParams();
  const [events, setEvents] = useState([]);
  const [connectionStatus, setConnectionStatus] = useState('connecting');
  const [eventCount, setEventCount] = useState(0);
  const wsRef = useRef(null);

  useEffect(() => {
    // Create WebSocket connection
    const ws = createAuctionWebSocket(id, {
      onOpen: () => {
        console.log('WebSocket connection established');
        setConnectionStatus('connected');
      },
      onMessage: (data) => {
        console.log('WebSocket message received:', data);
        const newEvent = {
          ...data,
          id: Date.now() + Math.random(), // Generate unique ID for React key
          timestamp: new Date().toISOString(),
        };
        setEvents((prev) => [newEvent, ...prev]); // Add to beginning (newest first)
        setEventCount((prev) => prev + 1);
      },
      onClose: () => {
        console.log('WebSocket connection closed');
        setConnectionStatus('disconnected');
      },
      onError: (error) => {
        console.error('WebSocket error:', error);
        setConnectionStatus('error');
      },
    });

    wsRef.current = ws;

    // Cleanup: close WebSocket when component unmounts
    return () => {
      if (wsRef.current) {
        wsRef.current.close();
        wsRef.current = null;
      }
    };
  }, [id]);

  const handleReconnect = () => {
    if (wsRef.current) {
      wsRef.current.reconnect();
      setConnectionStatus('connecting');
    }
  };

  const handleClearLog = () => {
    setEvents([]);
    setEventCount(0);
  };

  const getStatusColor = (status) => {
    switch (status) {
      case 'connected':
        return 'bg-emerald-100 text-emerald-800 border-emerald-200';
      case 'connecting':
        return 'bg-yellow-100 text-yellow-800 border-yellow-200';
      case 'disconnected':
      case 'error':
        return 'bg-red-100 text-red-800 border-red-200';
      default:
        return 'bg-gray-100 text-gray-800 border-gray-200';
    }
  };

  const getStatusIcon = (status) => {
    switch (status) {
      case 'connected':
        return '●';
      case 'connecting':
        return '◐';
      case 'disconnected':
      case 'error':
        return '○';
      default:
        return '?';
    }
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold text-gray-900">
            Live Events - Auction #{id}
          </h1>
          <p className="mt-1 text-gray-600">
            Real-time WebSocket event feed
          </p>
        </div>
        <Link
          to={`/auctions/${id}`}
          className="text-blue-600 hover:text-blue-700 font-medium"
        >
          Back to Auction
        </Link>
      </div>

      {/* Connection Status */}
      <div className="bg-white shadow-sm rounded-lg border border-gray-200 p-6">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-4">
            <div>
              <h2 className="text-sm font-medium text-gray-700">Connection Status</h2>
              <div className="mt-2 flex items-center gap-2">
                <span
                  className={`inline-flex items-center px-3 py-1 rounded-full text-sm font-medium border ${getStatusColor(
                    connectionStatus
                  )}`}
                >
                  <span className="mr-2">{getStatusIcon(connectionStatus)}</span>
                  {connectionStatus.charAt(0).toUpperCase() + connectionStatus.slice(1)}
                </span>
              </div>
            </div>
            <div>
              <h2 className="text-sm font-medium text-gray-700">Events Received</h2>
              <div className="mt-2 text-2xl font-bold text-gray-900">{eventCount}</div>
            </div>
          </div>

          <div className="flex gap-2">
            <button
              onClick={handleReconnect}
              disabled={connectionStatus === 'connected'}
              className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:opacity-50 disabled:cursor-not-allowed text-sm font-medium"
            >
              Reconnect
            </button>
            <button
              onClick={handleClearLog}
              className="px-4 py-2 bg-gray-600 text-white rounded-md hover:bg-gray-700 focus:outline-none focus:ring-2 focus:ring-gray-500 text-sm font-medium"
            >
              Clear Log
            </button>
          </div>
        </div>
      </div>

      {/* Event Feed */}
      <div className="bg-white shadow-sm rounded-lg border border-gray-200 p-6">
        <h2 className="text-lg font-semibold text-gray-900 mb-4">Event Feed</h2>

        {events.length === 0 && (
          <div className="text-center py-12">
            <svg
              className="mx-auto h-12 w-12 text-gray-400"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z"
              />
            </svg>
            <h3 className="mt-2 text-sm font-medium text-gray-900">No events yet</h3>
            <p className="mt-1 text-sm text-gray-500">
              {connectionStatus === 'connected'
                ? 'Waiting for events from the server...'
                : 'Connect to start receiving events'}
            </p>
          </div>
        )}

        {events.length > 0 && (
          <div className="space-y-3">
            {events.map((event) => (
              <EventItem key={event.id} event={event} />
            ))}
          </div>
        )}
      </div>

      {/* Info Box */}
      <div className="bg-purple-50 border border-purple-200 rounded-lg p-4">
        <h3 className="text-sm font-medium text-purple-900 mb-2">
          WebSocket Connection Info
        </h3>
        <ul className="text-sm text-purple-800 space-y-1 list-disc list-inside">
          <li>Connection establishes automatically when page loads</li>
          <li>Events appear in real-time as they occur</li>
          <li>Automatic reconnection on connection loss</li>
          <li>Connection closes when you navigate away</li>
          <li>Use the Reconnect button to manually reconnect</li>
        </ul>
      </div>
    </div>
  );
}

export default WebSocketPage;
