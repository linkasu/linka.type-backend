package feature

import (
	"testing"

	"github.com/linkasu/linka.type-backend/internal/config"
)

func TestUseYDB(t *testing.T) {
	cfg := config.FeatureConfig{ReadSource: ReadFirebaseOnly}
	if UseYDB("user", cfg) {
		t.Fatalf("expected firebase_only to be false")
	}

	cfg.ReadSource = ReadYDBPrimary
	if !UseYDB("user", cfg) {
		t.Fatalf("expected ydb_primary to be true")
	}

	cfg.ReadSource = ReadCohort
	cfg.CohortPercent = 0
	if UseYDB("user", cfg) {
		t.Fatalf("expected cohort=0 to be false")
	}
	cfg.CohortPercent = 100
	if !UseYDB("user", cfg) {
		t.Fatalf("expected cohort=100 to be true")
	}
}
