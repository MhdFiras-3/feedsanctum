package handlers

import (
	"database/sql"
	"time"

	"github.com/MhdFiras-3/feedsanctum/internal/database"
)

type APIConfig struct {
	DB        *database.Queries
	DBConn    *sql.DB
	JWTSecret string
	JWTExpiry time.Duration
	Ticker    time.Duration
}
