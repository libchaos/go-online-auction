/**
 * Validate listing ID
 * @param {string|number} value - Listing ID value
 * @returns {string|null} Error message or null if valid
 */
export const validateListingId = (value) => {
  if (!value) {
    return 'Listing ID is required';
  }
  
  const numValue = Number(value);
  
  if (isNaN(numValue)) {
    return 'Listing ID must be a number';
  }
  
  if (numValue <= 0) {
    return 'Listing ID must be a positive number';
  }
  
  if (!Number.isInteger(numValue)) {
    return 'Listing ID must be an integer';
  }
  
  return null;
};

/**
 * Validate end time
 * @param {string} value - End time value (ISO string or datetime-local format)
 * @returns {string|null} Error message or null if valid
 */
export const validateEndTime = (value) => {
  if (!value) {
    return 'End time is required';
  }
  
  try {
    const endTime = new Date(value);
    const now = new Date();
    
    if (isNaN(endTime.getTime())) {
      return 'Invalid date format';
    }
    
    if (endTime <= now) {
      return 'End time must be in the future';
    }
    
    return null;
  } catch {
    return 'Invalid date format';
  }
};

/**
 * Validate bid amount
 * @param {string|number} value - Bid amount value in cents
 * @returns {string|null} Error message or null if valid
 */
export const validateBidAmount = (value) => {
  if (!value) {
    return 'Bid amount is required';
  }
  
  const numValue = Number(value);
  
  if (isNaN(numValue)) {
    return 'Bid amount must be a number';
  }
  
  if (numValue <= 0) {
    return 'Bid amount must be positive';
  }
  
  if (!Number.isInteger(numValue)) {
    return 'Bid amount must be in cents (whole number)';
  }
  
  return null;
};
