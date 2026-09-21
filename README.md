# gofeed
[![Tests](https://github.com/MhdFiras-3/gofeed/actions/workflows/test.yml/badge.svg)](https://github.com/MhdFiras-3/gofeed/actions/workflows/test.yml)

`gofeed` is an RSS aggregator and REST API backend written in Go. Beyond CRUD endpoints for users, feeds and subscriptions, it runs a background scraper that fans out across feeds on a timer, handles duplicate posts through Postgres constraints, and tolerates malformed feed dates. Built on Chi, sqlc, goose, and PostgreSQL, with JWT auth, refresh-token rotation, and graceful shutdown.

---

## 🚀 Getting Started

### Prerequisites

- [Go](https://go.dev/doc/install) (v1.25 or later)
- [Docker](https://www.docker.com/) & [Docker Compose](https://docs.docker.com/compose/)
- [goose](https://github.com/pressly/goose) for database migrations


### Clone & Configure Environment

Clone the repository and create a `.env` file in the root directory:

```bash
git clone https://github.com/MhdFiras-3/gofeed.git
cd gofeed
```

`.env`:
```env
PORT=8080
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=yourpassword
DB_NAME=gofeed
TEST_DB_URL=postgres://postgres:yourpassword@localhost:5432/gofeed_test?sslmode=disable
JWT_SECRET=yoursecret
DUMMY_HASH='$argon2id$v=19$m=65536,t=1,p=12$LpTO8GOk8ajNAFczIs12uQ$e48btjY28JEWiIfEDYcJBjb1GLBZGqXoOyrClQ9EIr0'
```


Start PostgreSQL Container:

```bash
docker compose up -d
```
Apply schema migrations using goose and run the API server:
```bash
goose -dir sql/migrations postgres "postgres://postgres:yourpassword@localhost:5432/gofeed?sslmode=disable" up
go run cmd/server/main.go
```
To run tests:
```bash
go test -p 1 ./...
```
*Note:-p 1 runs packages sequentially. The shared test database doesn't support parallel packages.*

---

## 📐 Key Design Decisions & Trade-Offs

### 1. Atomic Operations via Database Transactions
Multi-step database state changes are wrapped in explicit database transactions:
- **Feed Creation (`HandlerCreateFeed`):** Creating a feed and automatically following it must be an atomic operation. If inserting the `feed_follows` fails, the entire transaction is rolled back, preventing orphaned feed records without owners.
- **Refresh Token Rotation (`HandlerRefresh`):** During token rotation, revoking the existing refresh token and persisting the newly issued refresh token must succeed together. A transaction prevents race conditions where a user loses their session mid-rotation or an old token remains valid after failure.

### 2. Mitigation of Timing Attacks on Authentication
A server that immediately returns `401 Unauthorized` when an email does not exist responds significantly faster than when performing an expensive cryptographic hash comparison for an existing user. 
- To prevent **User Enumeration via Timing Attacks**, `HandlerLogin` executes an Argon2id comparison against a dummy hash (`DUMMY_HASH`) whenever a user lookup returns no rows. 
- This ensures constant-time response latency regardless of whether an email is registered.

### 3. Error Handling Trade-Offs in Background Scraping
The `scrapeFeed` background worker prioritizes system resiliency over fail-fast behavior:
- **Batch Processing:** When parsing a feed with multiple items, a single malformed post or duplicate URL does not abort the batch. Instead, individual errors are logged, the failed post is skipped, and remaining valid posts in the feed are processed and inserted.
- **Feed Polling Safety:** The feed's `last_fetched_at` timestamp is updated immediately before post insertion. This prevents a persistently broken feed or corrupted post payload from trapping the scraper in an infinite retry loop during subsequent fetch cycles.

### 4. Decoupled API Contracts from Database Models
Handlers explicitly define custom `requestParam` and `response` struct types rather than directly embedding or exposing `sqlc`-generated database structs:
- **Encapsulation & Security:** Internal schema details are never inadvertently leaked to clients via JSON serialization.
- **Decoupled Contracts:** Changes to the underlying database schema or migrations do not directly break external client contracts, allowing independent evolution of the API and database layers.
- **Null-Value Handling:** Custom response types map `sql.NullString` and `sql.NullTime` into clean nullable JSON primitives (e.g., `*string`, `*time.Time`) rather than exposing raw database struct types to the caller.

---

## 🏗️ System Architecture

```mermaid
graph TD
    Client[HTTP Client] -->|REST API Requests| Router[Chi Router]
    Router -->|MiddlewareAuth| Handlers[HTTP Handlers]
    Handlers -->|sqlc Queries| DB[(PostgreSQL)]
    
    subgraph Background Service
        Scraper[Scraper Worker] -->|Fetch RSS XML| ExternalFeeds[External RSS Feeds]
        Scraper -->|Insert New Posts| DB
    end
```

---

## 🌟 Key Features

- **Concurrent Feed Scraping:** Runs a concurrent scraper that polls feeds on a configurable `time.Ticker` interval to fetch multiple RSS feeds concurrently.
- **Feed Parsing:** Parses RSS 2.0 XML structures, unescapes HTML entities, and handles multiple `pubDate` formats (RFC1123, RFC3339, etc.).
- **Authentication:** User registration and authentication by **Argon2id** password hashing and **JWT** for HTTP authorization.
- **Transactional DB Operations:** Uses PostgreSQL transactions to ensure atomicity when creating feeds and feed follow and when refreshing access token.
- **Type-Safe Database Access:** Type-safe SQL query generation using **sqlc** and schema migrations managed via **goose**.
- **Integration Test Suite:** Test running migrations against an isolated Postgres instance via **goose**, with coverage of the scraper logic, JWT auth middleware, and the create feed handler.

---

## 🛠️ Tech Stack

- **Language:** [Go 1.25](https://go.dev/)
- **HTTP Router:** [Chi Router](https://github.com/go-chi/chi)
- **Database:** [PostgreSQL](https://www.postgresql.org/)
- **SQL Code Generation:** [sqlc](https://sqlc.dev/)
- **Database Migrations:** [goose](https://github.com/pressly/goose)
- **Password Hashing:** `golang.org/x/crypto/argon2`
- **OpenAPI/Swagger UI**
- **Docker**

---

## API Endpoints Summary
#### Public Routes
| Method | Endpoint | Description |
| :--- | :--- | :--- |
| `POST` | `/api/v1/register` | Register a new user account |
| `POST` | `/api/v1/login` | Authenticate user and receive JWT access token |
| `POST` | `/api/v1/refresh` | Rotate Refresh token and issue new access token |
| `POST` | `/api/v1/logout` | Revoke refresh token |
| `GET` | `/api/v1/feeds` | Get all created RSS feeds |
| `GET` | `/api/v1/feeds/{feedID}` | Get a specific RSS feed by ID |

#### Protected Routes
| Method | Endpoint | Description |
| :--- | :--- | :--- |
| `GET` | `/api/v1/me` | Retrieve current user profile |
| `PATCH` | `/api/v1/me` | Update user profile details |
| `DELETE` | `/api/v1/me` | Delete current user account |
| `POST` | `/api/v1/feeds` | Create a new RSS feed and automatically follow it |
| `GET` | `/api/v1/follows` | Get all feed follows for current user |
| `POST` | `/api/v1/follows` | Follow a feed |
| `DELETE` | `/api/v1/follows/{feedID}` | Unfollow a feed by ID |
| `GET` | `/api/v1/posts` | Fetch RSS posts for followed feeds |
| `POST` | `/api/v1/posts/{postID}/read` | Mark a specific post as read |
| `GET` | `/api/v1/posts/read` | Retrieve all posts marked as read by the user |

---

## 📄 License

This project is open-source and available under the [MIT License](LICENSE).
