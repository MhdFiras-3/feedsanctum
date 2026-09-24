package handlers

import (
	"os"
	"testing"

	"github.com/MhdFiras-3/feedsanctum/internal/testingutils"
	_ "github.com/lib/pq"
)

var testCfg *APIConfig

func TestMain(m *testing.M) {
	queries, db, cleanup := testingutils.SetupTestDB()
	testCfg = &APIConfig{
		DB:     queries,
		DBConn: db,
	}

	code := m.Run()

	cleanup()

	os.Exit(code)
}
