/**
 * Color palette constants for the auction frontend demo.
 * Inspired by modern livestream shopping platforms (e.g., Trofee).
 */

export const colors = {
  primary: '#FF5555',      // Coral Red - Primary accent for CTAs and live indicators
  success: '#10B981',      // Emerald Green - Active auctions, success states
  warning: '#F59E0B',      // Amber - Pending/connecting states
  error: '#EF4444',        // Red - Cancelled auctions, errors
  info: '#3B82F6',         // Blue - Links, informational elements
  
  // State-specific colors
  draft: '#6B7280',        // Gray - Draft auctions
  active: '#10B981',       // Green - Active auctions
  closed: '#3B82F6',       // Blue - Closed auctions
  cancelled: '#EF4444',    // Red - Cancelled auctions
  
  // Connection status
  connected: '#10B981',    // Green - Connected
  connecting: '#F59E0B',   // Amber - Connecting
  disconnected: '#EF4444', // Red - Disconnected
};
