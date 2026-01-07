import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import toast from 'react-hot-toast';
import { createAuction } from '../services/apiClient';
import { validateListingId, validateEndTime } from '../utils/validators';

/**
 * CreateAuctionPage - Form to create new auctions
 */
function CreateAuctionPage() {
  const navigate = useNavigate();
  const [loading, setLoading] = useState(false);
  const [formData, setFormData] = useState({
    listing_id: '',
    end_time: '',
  });
  const [errors, setErrors] = useState({
    listing_id: null,
    end_time: null,
  });

  const handleChange = (e) => {
    const { name, value } = e.target;
    setFormData((prev) => ({
      ...prev,
      [name]: value,
    }));

    // Clear error for this field when user starts typing
    setErrors((prev) => ({
      ...prev,
      [name]: null,
    }));
  };

  const validateForm = () => {
    const listingIdError = validateListingId(formData.listing_id);
    const endTimeError = validateEndTime(formData.end_time);

    setErrors({
      listing_id: listingIdError,
      end_time: endTimeError,
    });

    return !listingIdError && !endTimeError;
  };

  const handleSubmit = async (e) => {
    e.preventDefault();

    if (!validateForm()) {
      return;
    }

    setLoading(true);

    try {
      // Convert end_time from datetime-local to ISO 8601
      const endTimeISO = new Date(formData.end_time).toISOString();
      
      const response = await createAuction({
        listing_id: parseInt(formData.listing_id, 10),
        end_time: endTimeISO,
      });

      const createdAuction = response.data.data;
      toast.success(`Auction #${createdAuction.id} created successfully!`);
      
      // Navigate to auction detail page
      navigate(`/auctions/${createdAuction.id}`);
    } catch (error) {
      console.error('Failed to create auction:', error);
      // Error toast is handled by interceptor
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="max-w-2xl mx-auto">
      {/* Header */}
      <div className="mb-8">
        <h1 className="text-3xl font-bold text-gray-900">Create New Auction</h1>
        <p className="mt-2 text-gray-600">
          Fill in the details below to create a new auction
        </p>
      </div>

      {/* Form */}
      <div className="bg-white shadow-sm rounded-lg border border-gray-200 p-6">
        <form onSubmit={handleSubmit} className="space-y-6">
          {/* Listing ID Field */}
          <div>
            <label htmlFor="listing_id" className="block text-sm font-medium text-gray-700 mb-1">
              Listing ID <span className="text-red-500">*</span>
            </label>
            <input
              type="number"
              id="listing_id"
              name="listing_id"
              value={formData.listing_id}
              onChange={handleChange}
              className={`block w-full rounded-md border ${
                errors.listing_id ? 'border-red-500' : 'border-gray-300'
              } px-3 py-2 focus:border-blue-500 focus:ring-blue-500`}
              placeholder="Enter listing ID (positive integer)"
              disabled={loading}
            />
            {errors.listing_id && (
              <p className="mt-1 text-sm text-red-600">{errors.listing_id}</p>
            )}
            <p className="mt-1 text-sm text-gray-500">
              Enter a positive integer representing the listing ID
            </p>
          </div>

          {/* End Time Field */}
          <div>
            <label htmlFor="end_time" className="block text-sm font-medium text-gray-700 mb-1">
              End Time <span className="text-red-500">*</span>
            </label>
            <input
              type="datetime-local"
              id="end_time"
              name="end_time"
              value={formData.end_time}
              onChange={handleChange}
              className={`block w-full rounded-md border ${
                errors.end_time ? 'border-red-500' : 'border-gray-300'
              } px-3 py-2 focus:border-blue-500 focus:ring-blue-500`}
              disabled={loading}
            />
            {errors.end_time && (
              <p className="mt-1 text-sm text-red-600">{errors.end_time}</p>
            )}
            <p className="mt-1 text-sm text-gray-500">
              Select a future date and time when the auction should end
            </p>
          </div>

          {/* Form Actions */}
          <div className="flex gap-3">
            <button
              type="submit"
              disabled={loading}
              className="flex-1 bg-blue-600 text-white py-2 px-4 rounded-md hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 disabled:opacity-50 disabled:cursor-not-allowed font-medium"
            >
              {loading ? 'Creating...' : 'Create Auction'}
            </button>
            <button
              type="button"
              onClick={() => navigate('/')}
              disabled={loading}
              className="flex-1 bg-white text-gray-700 py-2 px-4 rounded-md border border-gray-300 hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-gray-500 focus:ring-offset-2 disabled:opacity-50 disabled:cursor-not-allowed font-medium"
            >
              Cancel
            </button>
          </div>
        </form>
      </div>

      {/* Help Text */}
      <div className="mt-6 bg-blue-50 border border-blue-200 rounded-lg p-4">
        <h3 className="text-sm font-medium text-blue-900 mb-2">What happens next?</h3>
        <ul className="text-sm text-blue-800 space-y-1 list-disc list-inside">
          <li>The auction will be created in <strong>draft</strong> state</li>
          <li>You can start the auction from the detail page</li>
          <li>Once started, users can place bids until the end time</li>
          <li>The auction will automatically close at the end time</li>
        </ul>
      </div>
    </div>
  );
}

export default CreateAuctionPage;
