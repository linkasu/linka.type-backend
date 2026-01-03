package feature

import (
	"hash/fnv"

	"github.com/linkasu/linka.type-backend/internal/config"
)

const (
	ReadFirebaseOnly = "firebase_only"
	ReadYDBPrimary   = "ydb_primary"
	ReadCohort       = "cohort"
)

// UseYDB returns true when the user should read from YDB.
func UseYDB(userID string, cfg config.FeatureConfig) bool {
	switch cfg.ReadSource {
	case ReadYDBPrimary:
		return true
	case ReadCohort:
		if cfg.CohortPercent <= 0 {
			return false
		}
		return hashPercent(userID) < cfg.CohortPercent
	default:
		return false
	}
}

func hashPercent(value string) int {
	h := fnv.New32a()
	_, _ = h.Write([]byte(value))
	return int(h.Sum32() % 100)
}
