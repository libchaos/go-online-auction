import { format, formatDistanceToNow, parseISO } from 'date-fns';

/**
 * Format cents to dollar currency string
 * @param {number|null} cents - Amount in cents
 * @returns {string} Formatted currency string (e.g., "$123.45")
 */
export const formatCurrency = (cents) => {
  if (cents === null || cents === undefined) {
    return '$0.00';
  }
  const dollars = cents / 100;
  return `$${dollars.toFixed(2)}`;
};

/**
 * Format ISO date string to human-readable format
 * @param {string} isoString - ISO 8601 date string
 * @returns {string} Formatted date string (e.g., "Jan 7, 2026 14:30")
 */
export const formatDate = (isoString) => {
  if (!isoString) {
    return 'N/A';
  }
  try {
    const date = parseISO(isoString);
    return format(date, 'MMM d, yyyy HH:mm');
  } catch (error) {
    console.error('Date formatting error:', error);
    return 'Invalid Date';
  }
};

/**
 * Format ISO date string to relative time
 * @param {string} isoString - ISO 8601 date string
 * @returns {string} Relative time string (e.g., "2 hours ago")
 */
export const formatRelativeTime = (isoString) => {
  if (!isoString) {
    return 'N/A';
  }
  try {
    const date = parseISO(isoString);
    return formatDistanceToNow(date, { addSuffix: true });
  } catch (error) {
    console.error('Relative time formatting error:', error);
    return 'Invalid Date';
  }
};
