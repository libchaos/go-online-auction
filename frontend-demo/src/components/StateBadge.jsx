import PropTypes from 'prop-types';

/**
 * StateBadge component - Displays auction state with color-coded badge
 */
function StateBadge({ state }) {
  const getStateStyles = (state) => {
    switch (state?.toLowerCase()) {
      case 'draft':
        return {
          bg: 'bg-gray-100',
          text: 'text-gray-800',
          label: 'Draft'
        };
      case 'active':
        return {
          bg: 'bg-emerald-100',
          text: 'text-emerald-800',
          label: 'Active'
        };
      case 'closed':
        return {
          bg: 'bg-blue-100',
          text: 'text-blue-800',
          label: 'Closed'
        };
      case 'cancelled':
        return {
          bg: 'bg-red-100',
          text: 'text-red-800',
          label: 'Cancelled'
        };
      default:
        return {
          bg: 'bg-gray-100',
          text: 'text-gray-800',
          label: state || 'Unknown'
        };
    }
  };

  const styles = getStateStyles(state);

  return (
    <span
      className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${styles.bg} ${styles.text}`}
    >
      {styles.label}
    </span>
  );
}

StateBadge.propTypes = {
  state: PropTypes.string.isRequired,
};

export default StateBadge;
