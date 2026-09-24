package scraper

import (
	"os"
	"testing"

	"github.com/MhdFiras-3/feedsanctum/internal/handlers"
	"github.com/MhdFiras-3/feedsanctum/internal/testingutils"
)

var testCfg *handlers.APIConfig

func TestMain(m *testing.M) {
	queries, _, cleanup := testingutils.SetupTestDB()
	testCfg = &handlers.APIConfig{
		DB: queries,
	}

	code := m.Run()

	cleanup()

	os.Exit(code)
}
