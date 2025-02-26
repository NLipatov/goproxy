package repositories

import (
	"database/sql"
	"encoding/json"
	"goproxy/dal"
	"reflect"
	"testing"
)

func assertNoError(t *testing.T, err error, message string) {
	if err != nil {
		t.Fatalf("%s: %v", message, err)
	}
}

func assertJSONEqual(t *testing.T, expected, actual string) {
	var expectedMap, actualMap map[string]interface{}

	err := json.Unmarshal([]byte(expected), &expectedMap)
	assertNoError(t, err, "Failed to unmarshal expected JSON")

	err = json.Unmarshal([]byte(actual), &actualMap)
	assertNoError(t, err, "Failed to unmarshal actual JSON")

	if !reflect.DeepEqual(expectedMap, actualMap) {
		t.Errorf("Expected JSON %s, got %s", expected, actual)
	}
}

func prepareCockroachDB(t *testing.T) (*sql.DB, func()) {
	_, db, cleanup := dal.SetupCockroachContainer(t)
	dal.Migrate(db)

	return db, cleanup
}

func prepareDb(t *testing.T) (*sql.DB, func()) {
	_, db, cleanup := dal.SetupPostgresContainer(t)
	dal.Migrate(db)

	return db, cleanup
}

func equalSlices[T comparable](a, b []T) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
