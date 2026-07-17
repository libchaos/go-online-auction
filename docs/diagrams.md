### C4 Model - Level 1: System Context

```plantuml
@startuml
!include https://raw.githubusercontent.com/plantuml-stdlib/C4-PlantUML/master/C4_Context.puml

Person(user, "User", "A person who creates auctions and places bids")

System(auctionSystem, "Go Online Auction System", "Allows users to create auctions, place bids, and receive real-time updates")

SystemDb_Ext(postgres, "PostgreSQL Database", "Stores auctions, bids, and related data")
SystemDb_Ext(nats, "NATS (JetStream)", "Command stream + event bus for reliable delivery")

Rel(user, auctionSystem, "Creates auctions, places bids, views real-time updates", "HTTPS, WebSocket")
Rel(auctionSystem, postgres, "Reads from and writes to", "SQL/TCP")
Rel(auctionSystem, nats, "Publishes commands/events and consumes them", "NATS Protocol")

@enduml
```

### C4 Model - Level 2: Container Diagram

```plantuml
@startuml
!include https://raw.githubusercontent.com/plantuml-stdlib/C4-PlantUML/master/C4_Container.puml

Person(user, "User", "A person interacting with the auction system")

System_Boundary(auctionSystem, "Go Online Auction System") {
    Container(spa, "Frontend", "React, Vite", "Provides auction UI, handles user interactions")
    Container(api, "Backend API", "Go, Chi Router", "Hexagonal Architecture with CQRS, handles business logic and persistence")
    Container(websocket, "WebSocket Server", "Go, Gorilla WebSocket", "Manages real-time connections and broadcasts events")
}

ContainerDb_Ext(postgres, "Database", "PostgreSQL", "Stores auctions and bids with strong consistency")
ContainerDb_Ext(nats, "Message Broker", "NATS JetStream", "Persists bid commands and distributes domain events")

Rel(user, spa, "Views auctions, places bids", "HTTPS")
Rel(user, websocket, "Subscribes to auction events", "WebSocket")

Rel(spa, api, "Makes API calls", "JSON/HTTPS")
Rel(spa, websocket, "Establishes WebSocket connection", "WebSocket")

Rel(api, postgres, "Reads/Writes data", "pgx/SQL")
Rel(api, nats, "Publishes bid commands and domain events", "nats.go")

Rel(websocket, nats, "Consumes events (durable consumer)", "nats.go")

@enduml
```

### C4 Model - Level 3: Component Diagram (Backend API)

```plantuml
@startuml
!include https://raw.githubusercontent.com/plantuml-stdlib/C4-PlantUML/master/C4_Component.puml

Container(spa, "Frontend", "React, Vite", "User interface")
Container(websocket, "WebSocket Server", "Go", "Real-time broadcasting")

Container_Boundary(api, "Backend API") {
    Component(httpHandlers, "HTTP Handlers", "Chi Router", "Handles HTTP requests and responses")
    
    Component(commands, "Commands", "Go", "Write operations (Create, Start, PlaceBid, Cancel, Close)")
    Component(queries, "Queries", "Go", "Read operations (GetAuctionByID, ListAuctions)")
    
    Component(domain, "Domain Models", "Go", "Business logic and rules (AuctionModel, BidModel)")
    
    Component(ports, "Ports", "Go Interfaces", "Repository and Event Dispatcher interfaces")
    
    Component(adapters, "Adapters", "Go", "Repository and Event Dispatcher implementations")
}

ContainerDb_Ext(postgres, "Database", "PostgreSQL", "Stores data")
ContainerDb_Ext(nats, "Message Broker", "NATS JetStream", "Command stream + event distribution")

Rel(spa, httpHandlers, "Makes API calls", "JSON/HTTPS")

Rel(httpHandlers, commands, "Invokes", "")
Rel(httpHandlers, queries, "Invokes", "")

Rel(commands, domain, "Uses and validates", "")
Rel(commands, ports, "Uses", "")

Rel(queries, ports, "Uses", "")

Rel(ports, adapters, "Implemented by", "")

Rel(adapters, postgres, "Reads/Writes", "pgx")
Rel(adapters, nats, "Publishes commands/events", "nats.go")

Rel(nats, websocket, "Delivers events to", "nats.go")

@enduml
```