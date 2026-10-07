package handlers

import (
	"context"
	"database/sql"
	"time"

	"github.com/MhdFiras-3/feedsanctum/internal/database"
	"github.com/go-chi/httprate"
)

type APIConfig struct {
	DB           *database.Queries
	DBConn       *sql.DB
	JWTSecret    string
	JWTExpiry    time.Duration
	Ticker       time.Duration
	LoginLimiter *httprate.RateLimiter
	ServerCTX    context.Context
}
