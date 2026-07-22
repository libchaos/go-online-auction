# Auction Frontend Demo

Demo React application for testing the Go auction backend APIs.

## Prerequisites

- Node.js 20.x LTS or later
- npm 9.x or later
- Backend API running on http://localhost:9000

## Installation

1. Navigate to frontend-demo directory:
```bash
cd frontend-demo
```

2. Install dependencies:
```bash
npm install
```

## Configuration

The application uses environment variables to configure the backend API URL.

**Default Configuration (.env):**
```
VITE_API_BASE_URL=http://localhost:9000/api/v1
```

**Local Overrides (.env.local - gitignored):**

Create `.env.local` if you need to override the backend URL:
```
VITE_API_BASE_URL=http://your-backend-url/api/v1
```

## Running

**Development Server:**
```bash
npm run dev
```
Opens at http://localhost:5173

**Production Build:**
```bash
npm run build
```
Outputs to `dist/` directory

**Preview Production Build:**
```bash
npm run preview
```
Preview built files at http://localhost:4173

**Linting:**
```bash
npm run lint
```

## Features

### 1. Auction List Page
- View all auctions in a responsive grid layout
- Filter by auction state (All, Draft, Active, Closed, Cancelled)
- Pagination controls with configurable page size (10, 25, 50 items)
- Color-coded state badges for visual distinction
- Click any auction card to view details

### 2. Create Auction Page
- Simple form to create test auctions
- Input fields: Listing ID (number), End Time (datetime)
- Real-time validation with inline error messages
- Automatic navigation to auction detail page on success

### 3. Auction Detail Page
- Comprehensive auction information display
- Complete bid history in reverse chronological order
- Currency formatting (cents to dollars)
- Human-readable timestamps
- Live countdown timer for active auctions
- State-dependent action buttons:
  - **Start Auction** (visible for draft auctions)
  - **Cancel Auction** (visible for draft/active auctions)
  - **Place Bid** (visible for active auctions)
- Real-time data refresh after actions

### 4. WebSocket Event Monitor
- Live event feed for real-time auction updates
- Visual connection status indicator
- Event types:
  - Bid placed (with amount and user ID)
  - Auction started (with timestamp)
  - Auction ended (with winner information)
- Expandable raw JSON payload for debugging
- Manual reconnect capability
- Clear event log functionality
- Event counter

## Architecture

### Project Structure

```
frontend-demo/
├── public/
│   └── favicon.ico
├── src/
│   ├── components/
│   │   ├── AuctionCard.jsx       # Single auction card
│   │   ├── BidList.jsx            # Bid history display
│   │   ├── StateBadge.jsx         # State badge with colors
│   │   ├── EventItem.jsx          # WebSocket event display
│   │   ├── Pagination.jsx         # Pagination controls
│   │   └── ErrorBoundary.jsx      # Error boundary wrapper
│   ├── pages/
│   │   ├── AuctionListPage.jsx    # Main list view
│   │   ├── CreateAuctionPage.jsx  # Auction creation form
│   │   ├── AuctionDetailPage.jsx  # Detail view with actions
│   │   └── WebSocketPage.jsx      # WebSocket subscription
│   ├── services/
│   │   ├── apiClient.js           # Axios HTTP client
│   │   └── websocketClient.js     # WebSocket factory
│   ├── utils/
│   │   ├── formatters.js          # Currency/date formatters
│   │   ├── validators.js          # Form validation
│   │   └── colors.js              # Color utilities
│   ├── App.jsx                    # Router and layout
│   ├── main.jsx                   # Entry point
│   └── index.css                  # TailwindCSS imports
├── .env                           # Environment variables
├── .env.local                     # Local overrides (gitignored)
├── index.html                     # HTML template
├── vite.config.js                 # Vite configuration
├── tailwind.config.js             # TailwindCSS configuration
├── postcss.config.js              # PostCSS configuration
├── eslint.config.js               # ESLint configuration
├── package.json                   # Dependencies and scripts
└── README.md                      # This file
```

### Tech Stack

- **Framework**: React 19 with Vite 6
- **Styling**: TailwindCSS v4
- **Routing**: React Router v7
- **HTTP Client**: Axios
- **WebSocket**: reconnecting-websocket
- **Notifications**: react-hot-toast
- **Date Formatting**: date-fns

### Design Decisions

- **No Global State**: Simple component state with `useState/useEffect`
- **URL Query Parameters**: Filter state preserved in URL for shareability
- **Centralized Error Handling**: Axios interceptors with toast notifications
- **Error Boundaries**: Graceful error handling for runtime errors
- **Loading States**: Visual feedback during all async operations

## Browser Support

Latest 2 versions of:
- Chrome
- Firefox
- Safari
- Edge

## API Integration

The application integrates with the following backend endpoints:

**Auction Endpoints:**
- `GET /api/v1/auctions` - List auctions with filtering
- `POST /api/v1/auctions` - Create new auction
- `GET /api/v1/auctions/:id` - Get auction details
- `PUT /api/v1/auctions/:id/start` - Start auction
- `PUT /api/v1/auctions/:id/cancel` - Cancel auction

**Bid Endpoints:**
- `POST /api/v1/auctions/:id/bids` - Place bid

**WebSocket:**
- `ws://localhost:9000/ws/v1/auctions/:id` - Live event stream

## Troubleshooting

**Backend Connection Issues:**
- Verify the backend is running on port 9000
- Check CORS configuration allows `http://localhost:5173`
- Confirm `VITE_API_BASE_URL` in `.env` or `.env.local`

**WebSocket Connection Failures:**
- Ensure backend WebSocket endpoint is accessible
- Check for firewall or proxy blocking WebSocket connections
- Verify auction ID is valid and exists

**Build Errors:**
- Clear `node_modules` and reinstall: `rm -rf node_modules && npm install`
- Clear Vite cache: `rm -rf node_modules/.vite`
- Update Node.js to latest LTS version

## Development Notes

This is a **demo application** for testing purposes. It does not include:
- User authentication or authorization
- Payment processing
- Multi-user session management
- Production monitoring or analytics
- Automated test suite (manual testing only)
- Data persistence beyond backend API

The application prioritizes simplicity and rapid development feedback over production-grade features.

## License

This is a demo application for educational and testing purposes.
