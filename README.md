# Go Online Auction System

A real-time online auction bidding system built with Go and React, featuring WebSocket-based live updates, CQRS architecture, and domain-driven design principles.

---

## Table of Contents

- [Overview](#overview)
- [Tech Stack](#tech-stack)
- [Features](#features)
- [Architecture](#architecture)
- [Directory Structure](#directory-structure)
- [Getting Started](#getting-started)
- [API Endpoints](#api-endpoints)
- [Frontend Screens](#frontend-screens)
- [Out of Scope](#out-of-scope)
- [Development](#development)

---

## Overview

This project implements a fully functional online auction system with strong consistency guarantees for concurrent bidding scenarios. The backend is built using hexagonal architecture (ports & adapters pattern) with clear separation between domain logic, application orchestration, and infrastructure concerns. The frontend provides a clean, minimalist interface for testing and interacting with the auction system in real-time.

**Key Highlights:**
- Real-time bid updates via WebSocket with Redis Pub/Sub
- CQRS pattern separating read and write operations
- Optimistic locking for concurrent bid handling
- Domain-driven design with rich domain models
- Hexagonal architecture with ports & adapters pattern

---

### C4 Model - Level 1: System Context
![C4 Model - Level 1 (System Context)](docs/diagram_001.png)

### C4 Model - Level 2: Container Diagram
![C4 Model - Level 1 (System Context)](docs/diagram_002.png)

### C4 Model - Level 3: Component Diagram (Backend API)
![C4 Model - Level 1 (System Context)](docs/diagram_003.png)

### 1. Hexagonal Architecture (Ports & Adapters)
The system follows hexagonal architecture principles to maintain clean separation of concerns:

- **Domain Layer**: Pure business logic with no external dependencies
  - Entities: `AuctionModel`, `BidModel`
  - Value Objects: `MoneyModel`, `AuctionStateEnum`, `BidStatusEnum`
  - Domain Events: `AuctionStartedEvent`, `BidPlacedEvent`, `AuctionEndedEvent`
  - Business Rules: Bid validation, state transitions, optimistic locking

- **Application Layer**: Use case orchestration
  - Commands: `CreateAuctionCommand`, `StartAuctionCommand`, `PlaceBidCommand`, `CancelAuctionCommand`, `CloseAuctionCommand`
  - Queries: `GetAuctionByIDQuery`, `ListAuctionsQuery`
  - Each handler exposes a single `Execute(ctx, input) (output, error)` method

- **Infrastructure Layer**: External integrations
  - PostgreSQL repositories implementing domain ports
  - Redis event dispatcher for Pub/Sub
  - Chi HTTP handlers and routers
  - WebSocket hub for real-time connections

### 2. CQRS (Command Query Responsibility Segregation)
Write operations (commands) are separated from read operations (queries):
- **Commands** modify state, emit events, and use Unit of Work for transactional consistency
- **Queries** read data directly from PostgreSQL without triggering side effects
- Events are dispatched only after successful transaction commit

### 3. Event-Driven Architecture
Domain events enable decoupling and real-time features:
- Commands emit domain events after state changes
- Events are published to Redis Pub/Sub channels: `auction:{auctionID}:events`
- WebSocket hub subscribes to Redis channels and broadcasts to connected clients
- Supports horizontal scaling (multiple server instances share Redis)

### 4. Optimistic Locking for Concurrency
High-concurrency bid scenarios are handled with:
- Version field on auction aggregate
- `SELECT FOR UPDATE` row locking during updates
- `NOWAIT` lock acquisition to fail fast under contention
- `ErrConcurrencyConflict` returned for retry at application layer

### 5. Unit of Work Pattern
Coordinates multi-repository operations within a single transaction:
- Commands that modify multiple tables use `AuctionUnitOfWork`
- Provides transactional scope for `AuctionRepository` and `BidRepository`
- Auto-rollback if `Complete()` not called
- Events dispatched only after `Complete()` succeeds

### 6. Dependency Injection with Uber Fx
- Compile-time dependency graph validation
- Lifecycle hooks for graceful startup/shutdown
- Interface-based dependency declarations
- Module-based organization

### 7. Frontend Architecture
- **No global state management** - Simple component state with `useState/useEffect`
- **URL query parameters** - Filter state preserved in URL for shareability
- **Centralized error handling** - Axios interceptors with toast notifications
- **Error boundaries** - Graceful handling of runtime errors
- **Loading states** - Visual feedback during all async operations

## Tech Stack

<details>
<summary><strong>Backend</strong></summary>

| Technology | Version | Purpose |
|------------|---------|---------|
| **Go** | 1.25.5 | Primary language |
| **Chi** | v5.2.3 | HTTP router |
| **PostgreSQL** | 17.5 | Primary database |
| **Redis** | 8 | Pub/Sub for events |
| **pgx** | v5.7.4 | PostgreSQL driver |
| **go-redis** | v9.17.2 | Redis client |
| **Gorilla WebSocket** | v1.5.3 | WebSocket support |
| **Cobra** | v1.10.2 | CLI framework |
| **Viper** | v1.21.0 | Configuration |
| **Zerolog** | v1.34.0 | Structured logging |
| **Uber Fx** | v1.24.0 | Dependency injection |
| **golang-migrate** | v4.19.1 | Database migrations |
| **Testify** | v1.11.1 | Testing framework |

</details>

<details>
<summary><strong>Frontend</strong></summary>

| Technology | Version | Purpose |
|------------|---------|---------|
| **React** | 19.2.0 | UI framework |
| **Vite** | 7.2.4 | Build tool |
| **React Router** | 7.11.0 | Client-side routing |
| **TailwindCSS** | 4.1.18 | Styling |
| **Axios** | 1.13.2 | HTTP client |
| **reconnecting-websocket** | 4.4.0 | WebSocket client |
| **react-hot-toast** | 2.6.0 | Notifications |
| **date-fns** | 4.1.0 | Date formatting |

</details>

<details>
<summary><strong>Infrastructure</strong></summary>

- **Docker Compose** - Local development environment
- **PostgreSQL 17.5 Alpine** - Database container
- **Redis 8 Alpine** - Caching and Pub/Sub container

</details>

---

<details>
<summary><strong>Features</strong></summary>

### Auction Management
- ✅ Create auctions with configurable end times
- ✅ Start auctions (Draft → Active state transition)
- ✅ Cancel auctions (Draft/Active → Cancelled)
- ✅ Automatic auction closure upon end time
- ✅ State machine validation (Draft, Active, Closed, Cancelled)

### Bidding System
- ✅ Place bids on active auctions
- ✅ Real-time bid validation (must exceed current highest bid)
- ✅ Optimistic locking for concurrent bids
- ✅ Bid history tracking with status (Accepted, Rejected, Superseded)
- ✅ Highest bid amount denormalization for performance

### Real-Time Updates
- ✅ WebSocket connections per auction
- ✅ Live bid placement notifications
- ✅ Auction state change events (started, ended)
- ✅ Redis Pub/Sub for horizontal scalability
- ✅ Automatic reconnection with exponential backoff

### Query Features
- ✅ List auctions with pagination (limit/offset)
- ✅ Filter auctions by state (draft, active, closed, cancelled)
- ✅ Get auction details with top 10 bids
- ✅ Total count for pagination
</details>

---

## Directory Structure

<details>
<summary><strong>Backend Structure</strong></summary>

```
go-online-auction/
├── cmd/                                    # CLI commands
│   ├── all.go                             # Run all modules
│   ├── auction.go                         # Run auction HTTP server
│   ├── websocket.go                       # Run WebSocket server
│   ├── db_migrate.go                      # Database migrations
│   └── root.go                            # Root command
│
├── internal/                              # Private application code
│   ├── modules/
│   │   └── auction/                       # Auction bounded context
│   │       ├── domain/                    # Domain models
│   │       │   ├── model/                 # Aggregates & entities
│   │       │   │   ├── auction_model.go
│   │       │   │   ├── bid_model.go
│   │       │   │   ├── listing_model.go
│   │       │   │   └── user_model.go
│   │       │   ├── event/                 # Domain events
│   │       │   │   ├── auction_started_event.go
│   │       │   │   ├── bid_placed_event.go
│   │       │   │   └── auction_ended_event.go
│   │       │   └── enum/                  # Enumerations
│   │       │       ├── auction_state_enum.go
│   │       │       └── bid_status_enum.go
│   │       │
│   │       ├── application/               # Use cases
│   │       │   ├── command/               # Write operations
│   │       │   │   ├── create_auction_command.go
│   │       │   │   ├── start_auction_command.go
│   │       │   │   ├── place_bid_command.go
│   │       │   │   ├── cancel_auction_command.go
│   │       │   │   └── close_auction_command.go
│   │       │   └── query/                 # Read operations
│   │       │       ├── get_auction_by_id_query.go
│   │       │       └── list_auctions_query.go
│   │       │
│   │       ├── ports/                     # Interfaces (hexagonal architecture)
│   │       │   ├── auction_repository.go
│   │       │   ├── bid_repository.go
│   │       │   ├── auction_unit_of_work.go
│   │       │   └── event_dispatchers.go
│   │       │
│   │       ├── infra/                     # Infrastructure implementations
│   │       │   ├── persistence/
│   │       │   │   ├── repository/        # Repository implementations
│   │       │   │   │   ├── auction_repository.go
│   │       │   │   │   └── bid_repository.go
│   │       │   │   ├── uow/               # Unit of Work
│   │       │   │   │   └── auction_unit_of_work.go
│   │       │   │   ├── entity/            # Database entities
│   │       │   │   │   ├── auction_entity.go
│   │       │   │   │   └── bid_entity.go
│   │       │   │   └── mapper/            # Entity-Domain mappers
│   │       │   │       ├── auction_mapper.go
│   │       │   │       └── bid_mapper.go
│   │       │   │
│   │       │   ├── event/
│   │       │   │   └── dispatcher/        # Redis Pub/Sub dispatchers
│   │       │   │       ├── redis_auction_started_event_dispatcher.go
│   │       │   │       ├── redis_bid_placed_event_dispatcher.go
│   │       │   │       └── redis_auction_ended_event_dispatcher.go
│   │       │   │
│   │       │   ├── websocket/             # WebSocket hub
│   │       │   │   ├── hub.go
│   │       │   │   └── registry.go
│   │       │   │
│   │       │   └── http/
│   │       │       ├── chi/
│   │       │       │   ├── handler/       # HTTP handlers
│   │       │       │   │   ├── auction_handler.go
│   │       │       │   │   └── websocket_handler.go
│   │       │       │   └── router/        # Route registration
│   │       │       │       ├── auction_router.go
│   │       │       │       └── websocket_router.go
│   │       │       ├── dto/               # HTTP DTOs
│   │       │       │   ├── auction.go
│   │       │       │   └── bid.go
│   │       │       └── errs/              # HTTP errors
│   │       │           └── errs.go
│   │       │
│   │       └── module.go                  # Fx module definition
│   │
│   └── shared/                            # Shared modules
│       ├── modules/
│       │   ├── config/                    # Configuration
│       │   ├── database/                  # Database setup
│       │   ├── httpserver/                # HTTP server
│       │   ├── logger/                    # Logging
│       │   └── redis/                     # Redis client
│       └── sdk/
│           ├── http/                      # HTTP helpers
│           │   ├── request/
│           │   └── response/
│           └── transaction/               # Transaction helpers
│
├── pkg/                                   # Public packages
│   ├── database/                          # Database utilities
│   ├── errs/                              # Error types
│   ├── httpserver/                        # HTTP server config
│   ├── logger/                            # Logger config
│   └── redis/                             # Redis config
│
├── migrations/                            # SQL migrations
│   ├── 000001_create_auctions_table.up.sql
│   ├── 000001_create_auctions_table.down.sql
│   ├── 000002_create_bids_table.up.sql
│   └── 000002_create_bids_table.down.sql
│
├── tests/                                 # Test files
│   └── mocks/                             # Mock implementations
│
├── tasks/                                 # Task management
├── docs/                                  # Documentation
├── main.go                                # Application entry point
├── Makefile                               # Build automation
├── docker-compose.yaml                    # Local infrastructure
└── go.mod                                 # Go dependencies
```

</details>

<details>
<summary><strong>Frontend Structure</strong></summary>

```
frontend-demo/
├── public/                                # Static assets
│   └── vite.svg
│
├── src/
│   ├── components/                        # Reusable components
│   │   ├── AuctionCard.jsx               # Auction card for list view
│   │   ├── BidList.jsx                   # Bid history display
│   │   ├── ErrorBoundary.jsx             # Error boundary wrapper
│   │   ├── EventItem.jsx                 # WebSocket event display
│   │   ├── Pagination.jsx                # Pagination controls
│   │   └── StateBadge.jsx                # State badge with colors
│   │
│   ├── pages/                            # Route pages
│   │   ├── AuctionListPage.jsx           # List view with filtering
│   │   ├── CreateAuctionPage.jsx         # Auction creation form
│   │   ├── AuctionDetailPage.jsx         # Detail view with actions
│   │   └── WebSocketPage.jsx             # WebSocket subscription
│   │
│   ├── services/                         # API clients
│   │   ├── apiClient.js                  # Axios HTTP client
│   │   └── websocketClient.js            # WebSocket factory
│   │
│   ├── utils/                            # Utility functions
│   │   ├── colors.js                     # Color mappings
│   │   └── formatters.js                 # Format helpers
│   │
│   ├── App.jsx                           # Main app with router
│   ├── main.jsx                          # Entry point
│   ├── App.css                           # App styles
│   └── index.css                         # Global styles (Tailwind)
│
├── index.html                            # HTML template
├── vite.config.js                        # Vite configuration
├── tailwind.config.js                    # Tailwind configuration
├── postcss.config.js                     # PostCSS configuration
├── eslint.config.js                      # ESLint configuration
├── package.json                          # Dependencies and scripts
└── README.md                             # Setup instructions
```

</details>

---

## Getting Started

<details>
<summary><strong>Prerequisites</strong></summary>

- **Go** 1.25.5 or later
- **Node.js** 18+ and npm
- **Docker** and Docker Compose
- **Make** (optional, for convenience commands)

</details>

<details>
<summary><strong>Installation</strong></summary>

### 1. Clone the repository
```bash
git clone <repository-url>
cd go-online-auction
```

### 2. Start infrastructure services
```bash
docker-compose up -d
```

This starts:
- PostgreSQL on `localhost:5432`
- Redis on `localhost:6379`

### 3. Run database migrations
```bash
make migrate
# or
go run ./main.go db:migrate
```

### 4. Start the backend server
```bash
# Run all modules (HTTP + WebSocket)
make run
# or
go run ./main.go all

# Or run separately:
# Auction HTTP API
go run ./main.go auction

# WebSocket server
go run ./main.go websocket
```

Backend runs on `http://localhost:8080`

### 5. Start the frontend development server
```bash
cd frontend-demo
npm install
npm run dev
```

Frontend runs on `http://localhost:5173`

</details>

<details>
<summary><strong>Configuration</strong></summary>

Backend configuration via environment variables or `.env` file:

```env
# Database
DATABASE_HOST=localhost
DATABASE_PORT=5432
DATABASE_USER=postgres
DATABASE_PASSWORD=postgres
DATABASE_NAME=go-online-auction
DATABASE_SSL_MODE=disable

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0

# HTTP Server
HTTP_SERVER_HOST=0.0.0.0
HTTP_SERVER_PORT=8080

# Logging
LOG_LEVEL=info
```

Frontend configuration in `frontend-demo/.env`:

```env
VITE_API_BASE_URL=http://localhost:8080/api/v1
VITE_WS_BASE_URL=ws://localhost:8080/ws/v1
```

</details>

---

## API Endpoints

<details>
<summary><strong>Auction Endpoints</strong></summary>

### Create Auction
```http
POST /api/v1/auctions
Content-Type: application/json

{
  "listing_id": 1,
  "end_time": "2026-01-15T18:00:00Z"
}

Response: 201 Created
{
  "id": 1,
  "listing_id": 1,
  "state": "draft",
  "start_time": null,
  "end_time": "2026-01-15T18:00:00Z",
  "highest_bid_amount_in_cents": null,
  "created_at": "2026-01-07T14:30:00Z"
}
```

### List Auctions
```http
GET /api/v1/auctions?state=active&limit=25&offset=0

Response: 200 OK
{
  "auctions": [...],
  "total_count": 45,
  "limit": 25,
  "offset": 0
}
```

Query Parameters:
- `state` (optional): Filter by state (`draft`, `active`, `closed`, `cancelled`)
- `limit` (optional): Results per page (default: 10)
- `offset` (optional): Pagination offset (default: 0)

### Get Auction by ID
```http
GET /api/v1/auctions/:id

Response: 200 OK
{
  "auction": {
    "id": 1,
    "listing_id": 1,
    "state": "active",
    "start_time": "2026-01-07T14:30:00Z",
    "end_time": "2026-01-15T18:00:00Z",
    "highest_bid_amount_in_cents": 50000,
    "created_at": "2026-01-07T14:00:00Z"
  },
  "bids": [
    {
      "id": 5,
      "auction_id": 1,
      "user_id": 42,
      "amount_in_cents": 50000,
      "created_at": "2026-01-07T15:00:00Z"
    }
  ]
}
```

Returns auction with top 10 bids ordered by amount descending.

### Start Auction
```http
PUT /api/v1/auctions/:id/start

Response: 200 OK
{
  "id": 1,
  "listing_id": 1,
  "state": "active",
  "start_time": "2026-01-07T14:30:00Z",
  "end_time": "2026-01-15T18:00:00Z",
  "created_at": "2026-01-07T14:00:00Z"
}
```

Transitions auction from `draft` to `active` state.

### Cancel Auction
```http
PUT /api/v1/auctions/:id/cancel

Response: 200 OK
{
  "id": 1,
  "listing_id": 1,
  "state": "cancelled",
  "start_time": null,
  "end_time": "2026-01-15T18:00:00Z",
  "created_at": "2026-01-07T14:00:00Z"
}
```

Transitions auction from `draft` or `active` to `cancelled` state.

</details>

<details>
<summary><strong>Bid Endpoints</strong></summary>

### Place Bid
```http
POST /api/v1/auctions/:id/bids
Content-Type: application/json

{
  "amount_in_cents": 55000
}

Response: 201 Created
{
  "id": 6,
  "auction_id": 1,
  "user_id": 42,
  "amount_in_cents": 55000,
  "created_at": "2026-01-07T15:30:00Z"
}
```

**Note**: User ID is auto-generated for demo purposes. In production, this would come from authentication.

</details>

<details>
<summary><strong>WebSocket Endpoint</strong></summary>

### Subscribe to Auction Events
```
ws://localhost:8080/ws/v1/auctions/:id
```

**Connection Flow:**
1. Client connects to WebSocket URL
2. Server sends `subscription_confirmed` message
3. Client receives real-time events as JSON

**Event Types:**

**Auction Started Event:**
```json
{
  "type": "auction_started",
  "event_id": "uuid",
  "timestamp": "2026-01-07T14:30:00Z",
  "auction_id": 1,
  "start_time": "2026-01-07T14:30:00Z",
  "end_time": "2026-01-15T18:00:00Z"
}
```

**Bid Placed Event:**
```json
{
  "type": "bid_placed",
  "event_id": "uuid",
  "timestamp": "2026-01-07T15:00:00Z",
  "auction_id": 1,
  "bid_id": 5,
  "user_id": 42,
  "amount_in_cents": 50000
}
```

**Auction Ended Event:**
```json
{
  "type": "auction_ended",
  "event_id": "uuid",
  "timestamp": "2026-01-15T18:00:00Z",
  "auction_id": 1,
  "final_amount_in_cents": 50000,
  "winning_bid_id": 5,
  "winner_user_id": 42
}
```

</details>

<details>
<summary><strong>Error Responses</strong></summary>

All errors follow a consistent format:

```json
{
  "error": {
    "code": "AUCTION_NOT_FOUND",
    "message": "Auction with ID 999 not found"
  }
}
```

**HTTP Status Codes:**
- `200 OK` - Successful request
- `201 Created` - Resource created successfully
- `400 Bad Request` - Invalid input
- `404 Not Found` - Resource not found
- `409 Conflict` - Concurrency conflict (optimistic lock failure)
- `500 Internal Server Error` - Server error

**Error Codes:**
- `AUCTION_NOT_FOUND` - Auction does not exist
- `AUCTION_INVALID_STATE_TRANSITION` - Invalid state transition
- `BID_AMOUNT_TOO_LOW` - Bid must exceed current highest bid
- `CONCURRENCY_CONFLICT` - Optimistic lock failure (retry)
- `INVALID_AUCTION_ID` - Invalid auction ID format

</details>

---

## Frontend Screens

<details>
<summary><strong>1. Auction List Page (Route: `/`)</strong></summary>

**Features:**
- Responsive grid layout (1 column mobile, 2-3 columns desktop)
- State filter dropdown (All, Draft, Active, Closed, Cancelled)
- Pagination controls with configurable page size
- Color-coded state badges
- Click to navigate to auction details

**Components Used:**
- `AuctionCard` - Individual auction cards
- `StateBadge` - Color-coded state indicators
- `Pagination` - Page navigation controls

**API Integration:**
- `GET /api/v1/auctions?state={state}&limit={n}&offset={n}`

</details>

<details>
<summary><strong>2. Create Auction Page (Route: `/create`)</strong></summary>

**Features:**
- Form with two fields: Listing ID and End Time
- Client-side validation
- Datetime picker for end time (must be future date)
- Success toast notification
- Auto-redirect to auction detail page on success

**API Integration:**
- `POST /api/v1/auctions`

</details>

<details>
<summary><strong>3. Auction Detail Page (Route: `/auctions/:id`)</strong></summary>

**Features:**
- Display auction metadata (ID, listing, state, times, highest bid)
- Countdown timer for active auctions
- Bid history list (top 10 bids, newest first)
- State-dependent action buttons:
  - "Start Auction" (draft state only)
  - "Cancel Auction" (draft or active state)
  - "Place Bid" form (active state only)
- Link to WebSocket subscription page
- Auto-refresh after state changes

**Components Used:**
- `StateBadge` - State indicator
- `BidList` - Bid history table

**API Integration:**
- `GET /api/v1/auctions/:id`
- `PUT /api/v1/auctions/:id/start`
- `PUT /api/v1/auctions/:id/cancel`
- `POST /api/v1/auctions/:id/bids`

</details>

<details>
<summary><strong>4. WebSocket Subscription Page (Route: `/auctions/:id/subscribe`)</strong></summary>

**Features:**
- Real-time event feed (newest at top)
- Connection status indicator (Connected, Disconnected, Connecting)
- Formatted event messages with emoji indicators:
  - 💰 Bid Placed
  - 🚀 Auction Started
  - 🏁 Auction Ended
- Collapsible raw JSON payload for debugging
- Event counter
- Manual reconnect button
- Clear event log button
- Auto-reconnect with exponential backoff

**Components Used:**
- `EventItem` - Individual event display

**WebSocket Integration:**
- `ws://localhost:8080/ws/v1/auctions/:id`

</details>

<details>
<summary><strong>Navigation</strong></summary>

The app uses React Router with a persistent navigation header:

**Header Links:**
- "Auction Demo" logo - Home link
- "Auctions" - List page
- "Create Auction" - Creation form

All pages are mobile-responsive with touch-friendly controls.

</details>

---

## Out of Scope

The following features are explicitly **not implemented** in the current version:

<details>
<summary><strong>Authentication & Authorization</strong></summary>

- User registration and login
- JWT or session-based authentication
- Role-based access control (admin, seller, bidder)
- User profile management
- OAuth integration

**Current Behavior:** User IDs are randomly generated for demo purposes

</details>

<details>
<summary><strong>Payment Processing</strong></summary>

- Payment gateway integration (Stripe, PayPal, etc.)
- Order fulfillment
- Invoice generation
- Refund processing
- Transaction history

</details>

<details>
<summary><strong>Advanced Auction Features</strong></summary>

- Minimum bid increments
- Reserve prices (minimum price to sell)
- Buy-it-now option
- Bid withdrawal
- Auction extensions
- Maximum auction duration constraints
- Automatic auction scheduling
- Recurring auctions
- Dutch auctions (descending price)

</details>

<details>
<summary><strong>Notification System</strong></summary>

- Email notifications
- SMS notifications
- Push notifications
- In-app notification center
- Notification preferences

**Current Behavior:** Events are only available via WebSocket subscription

</details>

<details>
<summary><strong>Advanced Querying</strong></summary>

- Full-text search
- Advanced filtering (price range, category, location)
- Sorting options
- Saved searches
- Auction recommendations

</details>

<details>
<summary><strong>Caching & Performance</strong></summary>

- Redis caching for read operations
- CQRS read models/projections
- CDN integration
- Image optimization
- Database query caching

</details>

<details>
<summary><strong>Monitoring & Observability</strong></summary>

- Metrics instrumentation (Prometheus)
- Distributed tracing (Jaeger, OpenTelemetry)
- Application performance monitoring (APM)
- Error tracking (Sentry)
- Custom dashboards

</details>

<details>
<summary><strong>Event Sourcing</strong></summary>

- Event store persistence
- Event replay
- Outbox pattern for guaranteed event delivery
- Event versioning
- Temporal queries

</details>

<details>
<summary><strong>Additional Features</strong></summary>

- Listing management (create, edit, delete listings)
- User management beyond basic entity identification
- Image uploads for listings
- Categories and tags
- Watchlists/favorites
- Seller ratings and reviews
- Dispute resolution
- Auction history archives
- Data export
- Admin dashboard
- API rate limiting
- API versioning beyond `/v1/`
- Internationalization (i18n)
- Multiple currency support

</details>

---

## Development

<details>
<summary><strong>Continuous Integration</strong></summary>

The project includes automated CI checks that run on every push and pull request:

**CI Jobs:**
- 🧪 **Unit Tests** - Runs the full test suite (`go test ./...`)
- 🔍 **Linting** - Static code analysis with golangci-lint
- 🔒 **Security Scan** - Vulnerability checking with govulncheck
- 🛡️ **Nil Safety** - Nil pointer analysis with nilaway

All checks must pass before code can be merged to ensure code quality and reliability.

</details>

<details>
<summary><strong>Available Make Commands</strong></summary>

```bash
# Install development tools
make install-libs

# Run the application
make run              # All modules
make run-auction      # Auction HTTP API only
make run-websocket    # WebSocket server only

# Database migrations
make migrate

# Testing
make test             # Run tests
make cover            # Test coverage

# Code quality
make lint             # Run linter
make static           # Static analysis
make vuln-check       # Vulnerability check
make nilaway          # Nil pointer analysis
```

</details>

<details>
<summary><strong>Running Tests</strong></summary>

```bash
# Backend tests
go test ./...

# With coverage
go test -cover ./...

# Frontend tests
cd frontend-demo
npm run lint
```

</details>

<details>
<summary><strong>Database Migrations</strong></summary>

Migrations are located in `migrations/` directory:

```bash
# Run all pending migrations
go run ./main.go db:migrate

# Or using golang-migrate CLI directly
migrate -path migrations -database "postgresql://postgres:postgres@localhost:5432/go-online-auction?sslmode=disable" up
```

**Migration Files:**
- `000001_create_auctions_table.up.sql` - Creates auctions table
- `000002_create_bids_table.up.sql` - Creates bids table

</details>

<details>
<summary><strong>Generating Mocks</strong></summary>

Mocks are used for testing:

```bash
# Generate mocks for all interfaces
mockery --all --dir=internal/modules/auction/ports --output=tests/mocks
```

Existing mocks are located in `tests/mocks/`

</details>

<details>
<summary><strong>Frontend Development</strong></summary>

```bash
cd frontend-demo

# Development server with hot reload
npm run dev

# Production build
npm run build

# Preview production build
npm run preview

# Linting
npm run lint
```

**Environment Configuration:**

Create a `.env.local` file for local overrides:

```env
VITE_API_BASE_URL=http://localhost:8080/api/v1
VITE_WS_BASE_URL=ws://localhost:8080/ws/v1
```

</details>

<details>
<summary><strong>Project Commands</strong></summary>

The project uses Cobra CLI with the following commands:

```bash
# Run all modules (HTTP + WebSocket)
go run ./main.go all

# Run auction HTTP API
go run ./main.go auction

# Run WebSocket server
go run ./main.go websocket

# Run database migrations
go run ./main.go db:migrate

# Help
go run ./main.go --help
```

</details>

<details>
<summary><strong>Docker Compose Services</strong></summary>

```bash
# Start all services
docker-compose up -d

# Stop all services
docker-compose down

# View logs
docker-compose logs -f

# Reset data (removes volumes)
docker-compose down -v
```

**Services:**
- `postgres` - PostgreSQL database on port 5432
- `redis` - Redis server on port 6379

</details>

---

## License

[MIT](licence)

---

**Built with ❤️ using Go and React**
