---
trigger: glob
globs: *.go
---

## Goal
Ship Go code that is **simple**, **testable**, **maintainable**, and **secure**—without bloated comments, cryptic names, or hidden contracts.

---

## 1. Architecture & Design

### 1.1 One job per unit
- A package models a single business concept.
- A file groups closely related types or functions.
- A function does exactly one thing.

### 1.2 Hide internals
- Export only what other packages truly need.
- Keep the public API minimal and intentional.

---

## 2. Naming & Comments

### 2.1 Spell it out
- Never abbreviate variable, function, or package names.
- Prefer `connectionStatus` over `connStat`.

### 2.2 Value-add comments only
- Eliminate obvious or redundant comments.
- Never comment:
  - Constructors
  - Structs
  - Functions
  - Methods
- **Comments are allowed only on interface declarations.**

❌ Bad
```go
// NewRedisAuctionEndedEventDispatcher creates a new RedisAuctionEndedEventDispatcher.
func NewRedisAuctionEndedEventDispatcher(...) *RedisAuctionEndedEventDispatcher {
    ...
}

// AuctionEndedPayloadData contains event-specific fields.
type AuctionEndedPayloadData struct {
    ...
}