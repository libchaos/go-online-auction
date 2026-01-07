import PropTypes from 'prop-types';
import { formatRelativeTime, formatCurrency } from '../utils/formatters';

/**
 * EventItem component - Display WebSocket event in formatted style
 */
function EventItem({ event }) {
  const getEventStyles = (eventType) => {
    switch (eventType) {
      case 'bid_placed':
        return {
          bg: 'bg-emerald-50',
          border: 'border-emerald-200',
          icon: '💰',
          title: 'Bid Placed',
        };
      case 'auction_started':
        return {
          bg: 'bg-blue-50',
          border: 'border-blue-200',
          icon: '🚀',
          title: 'Auction Started',
        };
      case 'auction_ended':
        return {
          bg: 'bg-gray-50',
          border: 'border-gray-200',
          icon: '🏁',
          title: 'Auction Ended',
        };
      case 'auction_cancelled':
        return {
          bg: 'bg-red-50',
          border: 'border-red-200',
          icon: '❌',
          title: 'Auction Cancelled',
        };
      case 'subscription_confirmed':
        return {
          bg: 'bg-purple-50',
          border: 'border-purple-200',
          icon: '✓',
          title: 'Subscription Confirmed',
        };
      default:
        return {
          bg: 'bg-gray-50',
          border: 'border-gray-200',
          icon: '📢',
          title: eventType || 'Event',
        };
    }
  };

  const styles = getEventStyles(event.event_type || event.type);

  const renderEventData = () => {
    if (!event.data) return null;

    if (event.event_type === 'bid_placed') {
      return (
        <div className="space-y-1">
          <div className="text-sm">
            <span className="font-medium">User:</span> #{event.data.user_id}
          </div>
          <div className="text-sm">
            <span className="font-medium">Amount:</span>{' '}
            {formatCurrency(event.data.amount?.amount_in_cents)}
          </div>
        </div>
      );
    }

    if (event.event_type === 'auction_started') {
      return (
        <div className="text-sm">
          <span className="font-medium">Auction ID:</span> {event.data.auction_id || event.auction_id}
        </div>
      );
    }

    if (event.event_type === 'auction_ended') {
      return (
        <div className="space-y-1">
          <div className="text-sm">
            <span className="font-medium">Final Bid:</span>{' '}
            {formatCurrency(event.data.highest_bid_amount_in_cents)}
          </div>
          {event.data.winning_user_id && (
            <div className="text-sm">
              <span className="font-medium">Winner:</span> #{event.data.winning_user_id}
            </div>
          )}
        </div>
      );
    }

    // Default: display all data as JSON
    return (
      <pre className="text-xs text-gray-600 overflow-x-auto">
        {JSON.stringify(event.data, null, 2)}
      </pre>
    );
  };

  return (
    <div className={`${styles.bg} border ${styles.border} rounded-lg p-4`}>
      <div className="flex items-start gap-3">
        <div className="text-2xl">{styles.icon}</div>
        <div className="flex-1 min-w-0">
          <div className="flex items-center justify-between mb-2">
            <h4 className="text-sm font-semibold text-gray-900">
              {styles.title}
            </h4>
            <span className="text-xs text-gray-500">
              {event.timestamp ? formatRelativeTime(event.timestamp) : 'Just now'}
            </span>
          </div>
          {renderEventData()}
          {event.event_id && (
            <div className="text-xs text-gray-500 mt-2">
              Event ID: {event.event_id}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}

EventItem.propTypes = {
  event: PropTypes.shape({
    type: PropTypes.string,
    event_type: PropTypes.string,
    event_id: PropTypes.string,
    timestamp: PropTypes.string,
    auction_id: PropTypes.number,
    data: PropTypes.object,
  }).isRequired,
};

export default EventItem;
