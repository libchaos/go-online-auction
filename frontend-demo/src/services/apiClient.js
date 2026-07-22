import axios from 'axios';
import toast from 'react-hot-toast';

const apiClient = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || 'http://localhost:9000/api/v1',
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Response interceptor for centralized error handling
apiClient.interceptors.response.use(
  (response) => response,
  (error) => {
    // Extract error message from different possible structures
    let message = 'An error occurred';
    
    if (error.response?.data) {
      const data = error.response.data;
      // Check for nested error object with message (e.g., {"error":{"code":"AUCTION_13","message":"Auction has expired"}})
      if (data.error?.message) {
        message = data.error.message;
      }
      // Check for direct message property (e.g., {"message":"Error text"})
      else if (data.message) {
        message = data.message;
      }
    }
    // Fallback to axios error message
    else if (error.message) {
      message = error.message;
    }
    
    toast.error(message);
    return Promise.reject(error);
  }
);

/**
 * Get list of auctions with optional filters
 * @param {Object} params - Query parameters (state, limit, offset)
 * @returns {Promise} Axios response with auction list
 */
export const getAuctions = (params) => 
  apiClient.get('/auctions', { params });

/**
 * Get auction details by ID
 * @param {number} id - Auction ID
 * @returns {Promise} Axios response with auction details and bids
 */
export const getAuctionById = (id) => 
  apiClient.get(`/auctions/${id}`);

/**
 * Create a new auction
 * @param {Object} data - Auction data (listing_id, end_time)
 * @returns {Promise} Axios response with created auction
 */
export const createAuction = (data) => 
  apiClient.post('/auctions', data);

/**
 * Start an auction
 * @param {number} id - Auction ID
 * @returns {Promise} Axios response with updated auction
 */
export const startAuction = (id) => 
  apiClient.put(`/auctions/${id}/start`);

/**
 * Cancel an auction
 * @param {number} id - Auction ID
 * @returns {Promise} Axios response with updated auction
 */
export const cancelAuction = (id) => 
  apiClient.put(`/auctions/${id}/cancel`);

/**
 * Place a bid on an auction
 * @param {number} id - Auction ID
 * @param {Object} data - Bid data (amount_in_cents)
 * @returns {Promise} Axios response (204 No Content on success)
 */
export const placeBid = (id, data) => 
  apiClient.post(`/auctions/${id}/bids`, data);

export default apiClient;
