import { useState, useEffect } from 'react';
import { useSearchParams } from 'react-router-dom';
import toast from 'react-hot-toast';
import { getAuctions } from '../services/apiClient';
import AuctionCard from '../components/AuctionCard';
import Pagination from '../components/Pagination';

/**
 * AuctionListPage - Display paginated list of auctions with filtering
 */
function AuctionListPage() {
  const [searchParams, setSearchParams] = useSearchParams();
  const [auctions, setAuctions] = useState([]);
  const [loading, setLoading] = useState(true);
  const [totalCount, setTotalCount] = useState(0);

  // Get query parameters with defaults
  const stateFilter = searchParams.get('state') || '';
  const limit = parseInt(searchParams.get('limit') || '10', 10);
  const currentPage = parseInt(searchParams.get('page') || '1', 10);
  const offset = (currentPage - 1) * limit;

  // Fetch auctions on mount and when filters change
  useEffect(() => {
    const fetchAuctions = async () => {
      setLoading(true);
      try {
        const params = {
          limit,
          offset,
        };
        
        // Only add state filter if it's not empty
        if (stateFilter) {
          params.state = stateFilter;
        }

        const response = await getAuctions(params);
        setAuctions(response.data.auctions || []);
        setTotalCount(response.data.total || 0);
      } catch (error) {
        console.error('Failed to fetch auctions:', error);
        // Error toast is handled by interceptor
      } finally {
        setLoading(false);
      }
    };

    fetchAuctions();
  }, [stateFilter, limit, offset]);

  const handleStateFilterChange = (e) => {
    const newState = e.target.value;
    const newParams = { page: '1', limit: limit.toString() };
    if (newState) {
      newParams.state = newState;
    }
    setSearchParams(newParams);
  };

  const handlePageChange = (newPage) => {
    const newParams = { page: newPage.toString(), limit: limit.toString() };
    if (stateFilter) {
      newParams.state = stateFilter;
    }
    setSearchParams(newParams);
  };

  const handlePageSizeChange = (newPageSize) => {
    const newParams = { page: '1', limit: newPageSize.toString() };
    if (stateFilter) {
      newParams.state = stateFilter;
    }
    setSearchParams(newParams);
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <h1 className="text-3xl font-bold text-gray-900">Auctions</h1>

        {/* State Filter */}
        <div className="flex items-center gap-2">
          <label htmlFor="stateFilter" className="text-sm font-medium text-gray-700">
            Filter by State:
          </label>
          <select
            id="stateFilter"
            value={stateFilter}
            onChange={handleStateFilterChange}
            className="block w-40 rounded-md border-gray-300 py-2 px-3 text-sm focus:border-blue-500 focus:ring-blue-500"
          >
            <option value="">All</option>
            <option value="draft">Draft</option>
            <option value="active">Active</option>
            <option value="closed">Closed</option>
            <option value="cancelled">Cancelled</option>
          </select>
        </div>
      </div>

      {/* Loading State */}
      {loading && (
        <div className="text-center py-12">
          <div className="inline-block animate-spin rounded-full h-8 w-8 border-4 border-blue-500 border-t-transparent"></div>
          <p className="mt-2 text-gray-600">Loading auctions...</p>
        </div>
      )}

      {/* Empty State */}
      {!loading && auctions.length === 0 && (
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
              d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2"
            />
          </svg>
          <h3 className="mt-2 text-sm font-medium text-gray-900">No auctions found</h3>
          <p className="mt-1 text-sm text-gray-500">
            {stateFilter 
              ? `No auctions with state "${stateFilter}"`
              : 'Get started by creating a new auction'}
          </p>
        </div>
      )}

      {/* Auction Grid */}
      {!loading && auctions.length > 0 && (
        <>
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            {auctions.map((auction) => (
              <AuctionCard key={auction.id} auction={auction} />
            ))}
          </div>

          {/* Pagination */}
          <Pagination
            currentPage={currentPage}
            totalCount={totalCount}
            pageSize={limit}
            onPageChange={handlePageChange}
            onPageSizeChange={handlePageSizeChange}
          />
        </>
      )}
    </div>
  );
}

export default AuctionListPage;
