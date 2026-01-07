import PropTypes from 'prop-types';
import { formatCurrency, formatRelativeTime } from '../utils/formatters';

/**
 * BidList component - Display list of bids for an auction
 */
function BidList({ bids }) {
  if (!bids || bids.length === 0) {
    return (
      <div className="text-center py-8 text-gray-500">
        No bids yet
      </div>
    );
  }

  return (
    <div className="space-y-3">
      {bids.map((bid) => (
        <div
          key={bid.id}
          className="bg-white border border-gray-200 rounded-lg p-4 hover:shadow-sm transition-shadow"
        >
          <div className="flex items-start justify-between">
            <div className="flex-1">
              <div className="flex items-center gap-2 mb-1">
                <span className="text-sm font-medium text-gray-900">
                  User #{bid.user_id}
                </span>
                <span className="text-xs text-gray-500">
                  {formatRelativeTime(bid.created_at)}
                </span>
              </div>
              <div className="text-xs text-gray-600">
                Bid ID: {bid.id}
              </div>
            </div>
            <div className="text-right">
              <div className="text-lg font-semibold text-emerald-600">
                {formatCurrency(bid.amount_in_cents)}
              </div>
            </div>
          </div>
        </div>
      ))}
    </div>
  );
}

BidList.propTypes = {
  bids: PropTypes.arrayOf(
    PropTypes.shape({
      id: PropTypes.number.isRequired,
      auction_id: PropTypes.number.isRequired,
      user_id: PropTypes.number.isRequired,
      amount_in_cents: PropTypes.number.isRequired,
      created_at: PropTypes.string.isRequired,
    })
  ),
};

BidList.defaultProps = {
  bids: [],
};

export default BidList;
