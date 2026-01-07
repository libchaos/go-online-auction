import { useState, useEffect } from 'react';
import { useParams, useNavigate, Link } from 'react-router-dom';
import toast from 'react-hot-toast';
import {
  getAuctionById,
  startAuction,
  cancelAuction,
  placeBid,
} from '../services/apiClient';
import StateBadge from '../components/StateBadge';
import BidList from '../components/BidList';
import { formatCurrency, formatDate } from '../utils/formatters';
import { validateBidAmount } from '../utils/validators';

/**
 * AuctionDetailPage - Display auction details with management actions
 */
function AuctionDetailPage() {
  const { id } = useParams();
  const navigate = useNavigate();
  const [auction, setAuction] = useState(null);
  const [bids, setBids] = useState([]);
  const [loading, setLoading] = useState(true);
  const [actionLoading, setActionLoading] = useState(false);
  const [showBidForm, setShowBidForm] = useState(false);
  const [bidAmount, setBidAmount] = useState('');
  const [bidError, setBidError] = useState(null);
  const [countdown, setCountdown] = useState(null);

  // Fetch auction data
  const fetchAuctionData = async () => {
    try {
      const response = await getAuctionById(id);
      setAuction(response.data.data.auction);
      setBids(response.data.data.bids || []);
    } catch (error) {
      console.error('Failed to fetch auction:', error);
      // Error toast is handled by interceptor
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchAuctionData();
  }, [id]);

  // Countdown timer for active auctions
  useEffect(() => {
    if (!auction || auction.state !== 'active' || !auction.end_time) {
      setCountdown(null);
      return;
    }

    const updateCountdown = () => {
      const endTime = new Date(auction.end_time);
      const now = new Date();
      const diff = endTime - now;

      if (diff <= 0) {
        setCountdown('Auction ended');
        return;
      }

      const days = Math.floor(diff / (1000 * 60 * 60 * 24));
      const hours = Math.floor((diff % (1000 * 60 * 60 * 24)) / (1000 * 60 * 60));
      const minutes = Math.floor((diff % (1000 * 60 * 60)) / (1000 * 60));
      const seconds = Math.floor((diff % (1000 * 60)) / 1000);

      let countdownStr = '';
      if (days > 0) countdownStr += `${days}d `;
      if (hours > 0 || days > 0) countdownStr += `${hours}h `;
      if (minutes > 0 || hours > 0 || days > 0) countdownStr += `${minutes}m `;
      countdownStr += `${seconds}s`;

      setCountdown(countdownStr);
    };

    updateCountdown();
    const interval = setInterval(updateCountdown, 1000);

    return () => clearInterval(interval);
  }, [auction]);

  const handleStartAuction = async () => {
    setActionLoading(true);
    try {
      await startAuction(id);
      toast.success('Auction started successfully!');
      await fetchAuctionData();
    } catch (error) {
      console.error('Failed to start auction:', error);
    } finally {
      setActionLoading(false);
    }
  };

  const handleCancelAuction = async () => {
    if (!confirm('Are you sure you want to cancel this auction?')) {
      return;
    }

    setActionLoading(true);
    try {
      await cancelAuction(id);
      toast.success('Auction cancelled successfully!');
      await fetchAuctionData();
    } catch (error) {
      console.error('Failed to cancel auction:', error);
    } finally {
      setActionLoading(false);
    }
  };

  const handleBidSubmit = async (e) => {
    e.preventDefault();

    // Convert dollars to cents
    const amountInDollars = parseFloat(bidAmount);
    if (isNaN(amountInDollars)) {
      setBidError('Please enter a valid amount');
      return;
    }

    const amountInCents = Math.round(amountInDollars * 100);
    const error = validateBidAmount(amountInCents);

    if (error) {
      setBidError(error);
      return;
    }

    setActionLoading(true);
    try {
      await placeBid(id, { amount_in_cents: amountInCents });
      toast.success('Bid placed successfully!');
      setBidAmount('');
      setShowBidForm(false);
      await fetchAuctionData();
    } catch (error) {
      console.error('Failed to place bid:', error);
    } finally {
      setActionLoading(false);
    }
  };

  if (loading) {
    return (
      <div className="text-center py-12">
        <div className="inline-block animate-spin rounded-full h-8 w-8 border-4 border-blue-500 border-t-transparent"></div>
        <p className="mt-2 text-gray-600">Loading auction details...</p>
      </div>
    );
  }

  if (!auction) {
    return (
      <div className="text-center py-12">
        <h2 className="text-2xl font-bold text-gray-900">Auction not found</h2>
        <button
          onClick={() => navigate('/')}
          className="mt-4 text-blue-600 hover:text-blue-700"
        >
          Back to auctions
        </button>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold text-gray-900">
            Auction #{auction.id}
          </h1>
          <p className="mt-1 text-gray-600">Listing ID: {auction.listing_id}</p>
        </div>
        <StateBadge state={auction.state} />
      </div>

      {/* Countdown Timer for Active Auctions */}
      {auction.state === 'active' && countdown && (
        <div className="bg-emerald-50 border border-emerald-200 rounded-lg p-4">
          <div className="flex items-center justify-between">
            <div>
              <h3 className="text-sm font-medium text-emerald-900">Time Remaining</h3>
              <p className="text-2xl font-bold text-emerald-700 mt-1">{countdown}</p>
            </div>
            <svg
              className="h-8 w-8 text-emerald-500"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"
              />
            </svg>
          </div>
        </div>
      )}

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Left Column: Auction Details */}
        <div className="lg:col-span-2 space-y-6">
          {/* Auction Metadata */}
          <div className="bg-white shadow-sm rounded-lg border border-gray-200 p-6">
            <h2 className="text-lg font-semibold text-gray-900 mb-4">
              Auction Details
            </h2>
            <dl className="grid grid-cols-1 sm:grid-cols-2 gap-4">
              <div>
                <dt className="text-sm font-medium text-gray-500">Auction ID</dt>
                <dd className="mt-1 text-sm text-gray-900">{auction.id}</dd>
              </div>
              <div>
                <dt className="text-sm font-medium text-gray-500">Listing ID</dt>
                <dd className="mt-1 text-sm text-gray-900">{auction.listing_id}</dd>
              </div>
              <div>
                <dt className="text-sm font-medium text-gray-500">State</dt>
                <dd className="mt-1">
                  <StateBadge state={auction.state} />
                </dd>
              </div>
              <div>
                <dt className="text-sm font-medium text-gray-500">Highest Bid</dt>
                <dd className="mt-1 text-lg font-semibold text-emerald-600">
                  {formatCurrency(auction.highest_bid_amount_in_cents)}
                </dd>
              </div>
              {auction.start_time && (
                <div>
                  <dt className="text-sm font-medium text-gray-500">Start Time</dt>
                  <dd className="mt-1 text-sm text-gray-900">
                    {formatDate(auction.start_time)}
                  </dd>
                </div>
              )}
              <div>
                <dt className="text-sm font-medium text-gray-500">End Time</dt>
                <dd className="mt-1 text-sm text-gray-900">
                  {formatDate(auction.end_time)}
                </dd>
              </div>
              <div>
                <dt className="text-sm font-medium text-gray-500">Created At</dt>
                <dd className="mt-1 text-sm text-gray-900">
                  {formatDate(auction.created_at)}
                </dd>
              </div>
              {auction.updated_at && (
                <div>
                  <dt className="text-sm font-medium text-gray-500">Updated At</dt>
                  <dd className="mt-1 text-sm text-gray-900">
                    {formatDate(auction.updated_at)}
                  </dd>
                </div>
              )}
            </dl>
          </div>

          {/* Bid History */}
          <div className="bg-white shadow-sm rounded-lg border border-gray-200 p-6">
            <h2 className="text-lg font-semibold text-gray-900 mb-4">
              Bid History ({bids.length})
            </h2>
            <BidList bids={bids} />
          </div>
        </div>

        {/* Right Column: Actions */}
        <div className="space-y-6">
          {/* Action Buttons */}
          <div className="bg-white shadow-sm rounded-lg border border-gray-200 p-6">
            <h2 className="text-lg font-semibold text-gray-900 mb-4">Actions</h2>
            <div className="space-y-3">
              {/* Start Auction (Draft only) */}
              {auction.state === 'draft' && (
                <button
                  onClick={handleStartAuction}
                  disabled={actionLoading}
                  className="w-full bg-emerald-600 text-white py-2 px-4 rounded-md hover:bg-emerald-700 focus:outline-none focus:ring-2 focus:ring-emerald-500 disabled:opacity-50 disabled:cursor-not-allowed font-medium"
                >
                  Start Auction
                </button>
              )}

              {/* Place Bid (Active only) */}
              {auction.state === 'active' && !showBidForm && (
                <button
                  onClick={() => setShowBidForm(true)}
                  disabled={actionLoading}
                  className="w-full bg-blue-600 text-white py-2 px-4 rounded-md hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:opacity-50 disabled:cursor-not-allowed font-medium"
                >
                  Place Bid
                </button>
              )}

              {/* Bid Form */}
              {auction.state === 'active' && showBidForm && (
                <form onSubmit={handleBidSubmit} className="space-y-3">
                  <div>
                    <label
                      htmlFor="bidAmount"
                      className="block text-sm font-medium text-gray-700 mb-1"
                    >
                      Bid Amount (USD)
                    </label>
                    <input
                      type="number"
                      id="bidAmount"
                      value={bidAmount}
                      onChange={(e) => {
                        setBidAmount(e.target.value);
                        setBidError(null);
                      }}
                      step="0.01"
                      min="0"
                      placeholder="0.00"
                      className={`block w-full rounded-md border ${
                        bidError ? 'border-red-500' : 'border-gray-300'
                      } px-3 py-2 focus:border-blue-500 focus:ring-blue-500`}
                      disabled={actionLoading}
                    />
                    {bidError && (
                      <p className="mt-1 text-sm text-red-600">{bidError}</p>
                    )}
                  </div>
                  <div className="flex gap-2">
                    <button
                      type="submit"
                      disabled={actionLoading}
                      className="flex-1 bg-blue-600 text-white py-2 px-4 rounded-md hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:opacity-50 disabled:cursor-not-allowed text-sm font-medium"
                    >
                      Submit Bid
                    </button>
                    <button
                      type="button"
                      onClick={() => {
                        setShowBidForm(false);
                        setBidAmount('');
                        setBidError(null);
                      }}
                      disabled={actionLoading}
                      className="flex-1 bg-white text-gray-700 py-2 px-4 rounded-md border border-gray-300 hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-gray-500 disabled:opacity-50 disabled:cursor-not-allowed text-sm font-medium"
                    >
                      Cancel
                    </button>
                  </div>
                </form>
              )}

              {/* Cancel Auction (Draft or Active) */}
              {(auction.state === 'draft' || auction.state === 'active') && (
                <button
                  onClick={handleCancelAuction}
                  disabled={actionLoading}
                  className="w-full bg-red-600 text-white py-2 px-4 rounded-md hover:bg-red-700 focus:outline-none focus:ring-2 focus:ring-red-500 disabled:opacity-50 disabled:cursor-not-allowed font-medium"
                >
                  Cancel Auction
                </button>
              )}

              {/* WebSocket Link */}
              <Link
                to={`/auctions/${id}/subscribe`}
                className="block w-full text-center bg-purple-600 text-white py-2 px-4 rounded-md hover:bg-purple-700 focus:outline-none focus:ring-2 focus:ring-purple-500 font-medium"
              >
                Live Events (WebSocket)
              </Link>

              {/* Back to List */}
              <button
                onClick={() => navigate('/')}
                className="w-full text-gray-700 py-2 px-4 rounded-md border border-gray-300 hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-gray-500 font-medium"
              >
                Back to Auctions
              </button>
            </div>
          </div>

          {/* Info Box */}
          {auction.state === 'active' && (
            <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
              <h3 className="text-sm font-medium text-blue-900 mb-2">
                Bidding Tips
              </h3>
              <ul className="text-sm text-blue-800 space-y-1 list-disc list-inside">
                <li>Enter your bid amount in dollars</li>
                <li>Your bid must be higher than the current highest bid</li>
                <li>Bids are final and cannot be retracted</li>
                <li>Monitor the countdown timer</li>
              </ul>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}

export default AuctionDetailPage;
