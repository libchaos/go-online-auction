# Go Online Auction System

A real-time online auction bidding system built with Go and React, featuring WebSocket-based live updates, CQRS architecture, and domain-driven design principles.

---

## Table of Contents

- [Overview](#overview)
- [Engineering Challenges & Solutions](#engineering-challenges--solutions)
- [Architecture Diagrams](#architecture-diagrams)
- [Project Structure](#project-structure)
- [Tech Stack](#tech-stack)
- [Features](#features)
- [Getting Started](#getting-started)
- [API Endpoints](#api-endpoints)
- [Frontend Demo](#frontend-demo)
- [Out of Scope](#out-of-scope)
- [Development](#development)

---

## Overview

This project solves the complex engineering challenge of building a concurrent, consistent auction bidding system at scale. The architecture demonstrates advanced backend patterns including hexagonal architecture, CQRS, event-driven design, and optimistic concurrency control to handle high-throughput bidding scenarios while maintaining data integrity.

**Engineering Focus:**
- **Concurrent Bid Handling**: Optimistic locking with database-level consistency guarantees
- **Event-Driven Architecture**: Redis Pub/Sub enabling horizontal scalability
- **CQRS Pattern**: Separation of write/read operations for performance and clarity
- **Hexagonal Architecture**: Ports & adapters pattern for testability and maintainability
- **Domain-Driven Design**: Rich domain models with enforced business invariants
- **Comprehensive CI/CD**: Automated testing, linting, security scanning, and nil safety analysis

---

## Engineering Challenges & Solutions

### 1. Concurrent Bid Handling
**Challenge**: Multiple users bidding simultaneously on the same auction creates race conditions where:
- Two bids might read the same "current highest bid" value
- Both might place bids thinking they're winning
- Database could accept both, violating business rules

**Solution**: Multi-layered concurrency control
```go
// Layer 1: Domain-level validation
if bid.Amount <= auction.HighestBidAmount {
    return ErrBidTooLow
}
```

```sql
// Layer 2: Database-level locking (repository layer)
SELECT * FROM auctions WHERE id = $1 FOR UPDATE NOWAIT

// Layer 3: Version-based optimistic locking
UPDATE auctions SET version = version + 1 WHERE id = $1 AND version = $2

-- Layer 4: Database trigger as final safety net (fallback)
-- Enforces business rule even if application logic is bypassed
CREATE TRIGGER enforce_higher_bids
    BEFORE INSERT ON bids
    FOR EACH ROW
    EXECUTE FUNCTION check_bid_amount();
```

**Trade-offs**:
- `NOWAIT` fails fast under contention → clients retry with exponential backoff
- Higher lock contention on popular auctions → acceptable for consistency guarantees
- Database trigger adds slight overhead but ensures data integrity as fallback
- Alternative considered: Serializable isolation level (rejected due to performance overhead)

### 2. Event Consistency
**Challenge**: Events must only be published after database transaction commits. Publishing before commit risks:
- WebSocket clients receiving events for failed transactions
- Inconsistent state between database and event stream

**Solution**: Unit of Work pattern with post-commit event dispatch
```go
func (u *AuctionUnitOfWork) Complete(ctx context.Context) error {
    if err := u.tx.Commit(ctx); err != nil {
        return err
    }
    // Events only dispatched after successful commit
    return u.dispatchEvents(ctx)
}
```

**Trade-offs**:
- Events dispatched in-process (not guaranteed delivery) → acceptable for real-time updates
- For critical events, would implement outbox pattern with separate worker

### 3. Horizontal Scalability
**Challenge**: Multiple backend instances need to share WebSocket connections and event distribution.

**Solution**: Redis Pub/Sub as event bus
- Each backend instance subscribes to `auction:{id}:events` channels
- Commands publish to Redis after transaction commit
- WebSocket hubs relay to connected clients
- Enables stateless backend instances behind load balancer

**Trade-offs**:
- Redis as single point of failure → mitigated with Redis Sentinel/Cluster in production
- At-most-once delivery semantics → acceptable for real-time notifications

### 4. Testing & Maintainability
**Challenge**: Complex business logic and infrastructure dependencies make testing difficult.

**Solution**: Hexagonal architecture with dependency inversion
```go
// Domain ports (interfaces)
type AuctionRepository interface {
    Save(ctx context.Context, auction *model.AuctionModel) error
    GetByID(ctx context.Context, id int64) (*model.AuctionModel, error)
}

// Command handlers depend on interfaces, not concrete implementations
type PlaceBidCommandHandler struct {
    uowFactory ports.AuctionUnitOfWorkFactory
}
```

**Benefits**:
- 100% unit test coverage of business logic without database
- Can test concurrency edge cases with mock implementations
- Infrastructure changes don't affect domain layer

---

## Architecture Diagrams

### C4 Model - Level 1: System Context
![C4 Model - Level 1 (System Context)](docs/diagram_001.png)

### C4 Model - Level 2: Container Diagram
![C4 Model - Level 1 (System Context)](docs/diagram_002.png)

### C4 Model - Level 3: Component Diagram (Backend API)
![C4 Model - Level 1 (System Context)](docs/diagram_003.png)

### 1. Hexagonal Architecture (Ports & Adapters)
The system follows hexagonal architecture principles to achieve testability and technology independence:

**Domain Layer** (Pure business logic, zero external dependencies)
- **Aggregates**: `AuctionModel` (root), `BidModel`
  - Enforce invariants: Cannot place bid on non-active auction
  - Encapsulate state transitions: Draft → Active → Closed/Cancelled
  - Generate domain events as side effects of state changes
- **Value Objects**: `MoneyModel`, `AuctionStateEnum`, `BidStatusEnum`
  - Immutable, self-validating
  - Example: `MoneyModel` ensures amounts are non-negative
- **Domain Events**: `AuctionStartedEvent`, `BidPlacedEvent`, `AuctionEndedEvent`
  - Captured during command execution
  - Dispatched only after successful transaction commit
- **Business Rules**:
  - New bid must exceed current highest by at least $1
  - Auctions cannot be started if already active
  - Bids cannot be placed after auction end time

**Application Layer** (Use case orchestration)
- **Commands** (Write operations):
  ```go
  type PlaceBidCommandHandler struct {
      uowFactory ports.AuctionUnitOfWorkFactory
  }
  
  func (h *PlaceBidCommandHandler) Execute(ctx context.Context, cmd PlaceBidCommand) (*BidModel, error) {
      uow := h.uowFactory.Create()
      defer uow.Rollback(ctx)
      
      // 1. Load auction with lock
      auction, err := uow.GetAuctionRepository().GetByID(ctx, cmd.AuctionID)
      // 2. Domain validation
      bid, err := auction.PlaceBid(cmd.UserID, cmd.Amount)
      // 3. Persist changes
      err = uow.GetBidRepository().Save(ctx, bid)
      err = uow.GetAuctionRepository().Save(ctx, auction)
      // 4. Commit & dispatch events
      return bid, uow.Complete(ctx)
  }
  ```
- **Queries** (Read operations):
  - Direct database reads without business logic
  - Optimized for presentation layer needs
  - Return DTOs, not domain models

**Infrastructure Layer** (Technology-specific implementations)
- **Adapters**:
  - `PostgreSQLAuctionRepository` implements `ports.AuctionRepository`
  - `RedisEventDispatcher` implements `ports.EventDispatcher`
  - `ChiAuctionHandler` adapts HTTP to command/query handlers
- **Benefits**:
  - Swap PostgreSQL for MySQL without touching domain code
  - Test with in-memory repository instead of real database
  - Replace Redis with Kafka without changing event contracts

### 2. CQRS (Command Query Responsibility Segregation)
Strict separation between write and read operations:

**Commands** (Write model):
- Load domain aggregates from repository
- Execute business logic through domain methods
- Emit domain events
- Use Unit of Work for transactional consistency
- Optimistic locking to handle concurrent writes
- Example: `PlaceBidCommand` → `PlaceBidCommandHandler`

**Queries** (Read model):
- Direct SQL queries optimized for UI needs
- No domain model instantiation
- No state modifications or side effects
- Can use database views or denormalized tables
- Example: `ListAuctionsQuery` returns paginated DTOs with total count

**Benefits**:
- **Performance**: Queries bypass domain layer overhead
- **Scalability**: Can use read replicas for queries
- **Clarity**: Clear distinction between state changes and data retrieval
- **Optimization**: Queries can be optimized independently (indexes, caching)

**Evolution Path**:
- Current: Shared PostgreSQL database
- Future: Separate read/write databases with event-driven synchronization
- Future: Read model as projection from event stream

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

## Project Structure

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

## Tech Stack

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

### Frontend

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

### Infrastructure

- **Docker Compose** - Local development environment
- **PostgreSQL 17.5 Alpine** - Database container
- **Redis 8 Alpine** - Caching and Pub/Sub container

---

## Features

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

### Frontend Structure

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

---

## Getting Started

### Prerequisites

- **Go** 1.25.5 or later
- **Node.js** 18+ and npm
- **Docker** and Docker Compose
- **Make** (optional, for convenience commands)

### Installation

### 1. Clone the repository
```bash
git clone https://github.com/cristiano-pacheco/go-online-auction
cd go-online-auction
```

### 2. Start infrastructure services
```bash
docker-compose up -d
```

This starts:
- PostgreSQL on `localhost:5432`
- Redis on `localhost:6379`

### 3. Make a copy of .env
```
cp .env.example .env
```

### 4. Run database migrations
```bash
make migrate
# or
go run ./main.go db:migrate
```

### 5. Start the backend server
```bash
# Run all modules (HTTP + WebSocket)
make run

# Or run separately:
# Auction HTTP API
go run ./main.go auction

# WebSocket server
go run ./main.go websocket
```

Backend runs on `http://localhost:9000`

### 5. Start the frontend development server
```bash
cd frontend-demo
cp .env.example .env
npm install
npm run dev -- --host
```

Frontend runs on `http://localhost:5173`

### Configuration

Backend configuration via environment variables or `.env` file (see [.env.example](.env.example) for a full list of available variables):

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
```

---

## API Endpoints

### Auction Endpoints

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
  "data": {
    "id": 1,
    "listing_id": 1,
    "state": "draft",
    "start_time": null,
    "end_time": "2026-01-15T18:00:00Z",
    "highest_bid_amount_in_cents": null,
    "created_at": "2026-01-07T14:30:00Z"
  }
}
```

### List Auctions
```http
GET /api/v1/auctions?state=active&limit=25&offset=0

Response: 200 OK
{
  "data": {
    "auctions": [...],
    "total_count": 45,
    "limit": 25,
    "offset": 0
  }
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
  "data": {
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
}
```

Returns auction with top 10 bids ordered by amount descending.

### Start Auction
```http
PUT /api/v1/auctions/:id/start

Response: 200 OK
{
  "data": {
    "id": 1,
    "listing_id": 1,
    "state": "active",
    "start_time": "2026-01-07T14:30:00Z",
    "end_time": "2026-01-15T18:00:00Z",
    "created_at": "2026-01-07T14:00:00Z"
  }
}
```

Transitions auction from `draft` to `active` state.

### Cancel Auction
```http
PUT /api/v1/auctions/:id/cancel

Response: 200 OK
{
  "data": {
    "id": 1,
    "listing_id": 1,
    "state": "cancelled",
    "start_time": null,
    "end_time": "2026-01-15T18:00:00Z",
    "created_at": "2026-01-07T14:00:00Z"
  }
}
```

Transitions auction from `draft` or `active` to `cancelled` state.

### Bid Endpoints

### Place Bid
```http
POST /api/v1/auctions/:id/bids
Content-Type: application/json

{
  "amount_in_cents": 55000
}

Response: 201 Created
{
  "data": {
    "id": 6,
    "auction_id": 1,
    "user_id": 42,
    "amount_in_cents": 55000,
    "created_at": "2026-01-07T15:30:00Z"
  }
}
```

**Note**: User ID is auto-generated for demo purposes. In production, this would come from authentication.

### WebSocket Endpoint

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

### Error Responses

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

---

## Frontend Demo

A minimal React SPA for testing and demonstrating the backend APIs and WebSocket functionality.

**Tech Stack**: React 19, Vite, TailwindCSS, Axios, React Router

**Setup**: See [frontend-demo/README.md](frontend-demo/README.md) for installation instructions.

### Screens

**Auction List** - Browse auctions with state filtering and pagination

![Auction List](docs/list.png)

**Create Auction** - Simple form to create new auctions

![Create Auction](docs/create.png)

**Place Bid** - Real-time bidding with state management

![Place Bid](docs/placebid.gif)

**WebSocket Events** - Live event subscription for testing

![WebSocket Events](docs/realtime.gif)

---

## Out of Scope

The following features are explicitly **not implemented** in the current version:

### Authentication & Authorization

- User registration and login
- JWT or session-based authentication
- Role-based access control (admin, seller, bidder)
- User profile management
- OAuth integration

**Current Behavior:** User IDs are randomly generated for demo purposes

### Payment Processing

- Payment gateway integration (Stripe, PayPal, etc.)
- Order fulfillment
- Invoice generation
- Refund processing
- Transaction history

### Advanced Auction Features

- Minimum bid increments
- Reserve prices (minimum price to sell)
- Buy-it-now option
- Bid withdrawal
- Auction extensions
- Maximum auction duration constraints
- Automatic auction scheduling
- Recurring auctions
- Dutch auctions (descending price)

### Notification System

- Email notifications
- SMS notifications
- Push notifications
- In-app notification center
- Notification preferences

**Current Behavior:** Events are only available via WebSocket subscription

### Advanced Querying

- Full-text search
- Advanced filtering (price range, category, location)
- Sorting options
- Saved searches
- Auction recommendations

### Caching & Performance

- Redis caching for read operations
- CQRS read models/projections
- CDN integration
- Image optimization
- Database query caching

### Monitoring & Observability

- Metrics instrumentation (Prometheus)
- Distributed tracing (Jaeger, OpenTelemetry)
- Application performance monitoring (APM)
- Error tracking (Sentry)
- Custom dashboards

### Event Sourcing

- Event store persistence
- Event replay
- Outbox pattern for guaranteed event delivery
- Event versioning
- Temporal queries

### Additional Features

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

---

## Development

### Continuous Integration & Code Quality

The project enforces strict code quality gates via GitHub Actions CI pipeline that runs on every push and pull request:

**CI Jobs:**

**1. Unit Tests** (`go test ./...`)
- Runs entire test suite with race detector enabled
- Generates coverage reports
- Tests use mocks for external dependencies (database, Redis)
- Focus on domain logic and command/query handlers
- Example test coverage:
  - Domain models: Business rule validation
  - Command handlers: Success and error paths
  - Concurrency scenarios: Optimistic lock failures

**2. Linting** (`golangci-lint`)
- Runs 40+ linters including:
  - `errcheck`: Ensures all errors are checked
  - `govet`: Detects suspicious constructs
  - `staticcheck`: Advanced static analysis
  - `gosec`: Security-focused checks
  - `gocyclo`: Cyclomatic complexity limits
  - `dupl`: Duplicate code detection
- Configuration: `.golangci.yml` with custom rules
- Zero tolerance: All issues must be resolved

**3. Security Scanning** (`govulncheck`)
- Scans for known vulnerabilities in dependencies
- Checks both direct and transitive dependencies
- Integrated with Go's official vulnerability database
- Fails build on HIGH/CRITICAL vulnerabilities

**4. Nil Safety Analysis** (`nilaway`)
- Static analysis tool detecting nil pointer dereferences
- Catches potential runtime panics at compile time
- Enforces nil-safe patterns:
  ```go
  // Unsafe: nilaway catches this
  var auction *AuctionModel
  auction.Start() // potential nil dereference
  
  // Safe: explicit nil check
  if auction != nil {
      auction.Start()
  }
  ```

**Enforcement**:
- All checks must pass before merge to `main`
- Branch protection rules prevent bypassing
- Pre-commit hooks available for local validation: `make lint`

**Quality Metrics**:
- Test coverage target: >80% for domain layer
- Zero linting errors policy
- No known security vulnerabilities
- No potential nil pointer panics

### Available Make Commands

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

### Running Tests

```bash
# Backend tests
go test ./...

# With coverage
go test -cover ./...

# Frontend tests
cd frontend-demo
npm run lint
```

### Database Migrations

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

### Generating Mocks

Mocks are used for testing:

```bash
# Generate mocks for all interfaces
mockery --all --dir=internal/modules/auction/ports --output=tests/mocks
```

Existing mocks are located in `tests/mocks/`

### Frontend Development

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

### Project Commands

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

### Docker Compose Services

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

---

## License

[MIT](licence)

---

**Built with ❤️ using Go and React**
