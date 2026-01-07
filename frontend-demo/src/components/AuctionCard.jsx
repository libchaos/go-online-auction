import PropTypes from 'prop-types';
import { Link } from 'react-router-dom';
import StateBadge from './StateBadge';
import { formatCurrency, formatDate } from '../utils/formatters';

/**
 * AuctionCard component - Display auction information in a card format
 */
function AuctionCard({ auction }) {
  return (
    <Link
      to={`/auctions/${auction.id}`}
      className="block bg-white rounded-lg shadow hover:shadow-md transition-shadow border border-gray-200 p-4"
    >
      <div className="flex items-start justify-between mb-3">
        <div>
          <h3 className="text-lg font-semibold text-gray-900">
            Auction #{auction.id}
          </h3>
          <p className="text-sm text-gray-600">
            Listing ID: {auction.listing_id}
          </p>
        </div>
        <StateBadge state={auction.state} />
      </div>

      <div className="space-y-2">
        <div className="flex justify-between text-sm">
          <span className="text-gray-600">Highest Bid:</span>
          <span className="font-medium text-gray-900">
            {formatCurrency(auction.highest_bid_amount_in_cents)}
          </span>
        </div>

        <div className="flex justify-between text-sm">
          <span className="text-gray-600">End Time:</span>
          <span className="text-gray-900">
            {formatDate(auction.end_time)}
          </span>
        </div>

        {auction.start_time && (
          <div className="flex justify-between text-sm">
            <span className="text-gray-600">Started:</span>
            <span className="text-gray-900">
              {formatDate(auction.start_time)}
            </span>
          </div>
        )}

        <div className="flex justify-between text-sm">
          <span className="text-gray-600">Created:</span>
          <span className="text-gray-900">
            {formatDate(auction.created_at)}
          </span>
        </div>
      </div>
    </Link>
  );
}

AuctionCard.propTypes = {
  auction: PropTypes.shape({
    id: PropTypes.number.isRequired,
    listing_id: PropTypes.number.isRequired,
    state: PropTypes.string.isRequired,
    start_time: PropTypes.string,
    end_time: PropTypes.string.isRequired,
    highest_bid_amount_in_cents: PropTypes.number,
    created_at: PropTypes.string.isRequired,
  }).isRequired,
};

export default AuctionCard;
