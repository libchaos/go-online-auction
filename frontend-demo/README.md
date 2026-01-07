# Auction Frontend Demo

A React-based frontend demo application for testing the Go online auction backend APIs.

## Prerequisites

- Node.js 20.x or later
- npm 9.x or later
- Go auction backend running on `http://localhost:8080`

## Setup

1. Install dependencies:
```bash
npm install
```

2. Configure environment variables:
   - `.env` file is already configured with default backend URL
   - For local overrides, create `.env.local` (gitignored)

3. Start development server:
```bash
npm run dev
```

The application will be available at `http://localhost:5173`

## Project Structure

```
frontend-demo/
├── src/
│   ├── components/     # Reusable React components
│   ├── pages/          # Page components
│   ├── services/       # API and WebSocket clients
│   ├── utils/          # Helper functions
│   ├── App.jsx         # Main app component
│   └── main.jsx        # Entry point
├── .env                # Environment variables
└── package.json        # Dependencies
```

## Environment Variables

- `VITE_API_BASE_URL`: Backend API base URL (default: `http://localhost:8080/api/v1`)

## Development

- Dev server: `npm run dev`
- Build: `npm run build`
- Preview: `npm run preview`
- Lint: `npm run lint`

## Features (In Development)

- [ ] Auction list view with filtering
- [ ] Create auction form
- [ ] Auction detail view
- [ ] Real-time WebSocket updates
- [ ] Bid placement

## Tech Stack

- **Framework**: React 19 with Vite
- **Styling**: TailwindCSS v4
- **Routing**: React Router v7
- **HTTP Client**: Axios
- **WebSocket**: reconnecting-websocket
- **Notifications**: react-hot-toast
- **Date Formatting**: date-fns
