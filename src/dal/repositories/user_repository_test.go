package repositories

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"goproxy/dal/repositories/mocks"
	"goproxy/domain/aggregates"
	"os"
	"testing"
	"time"
)

const sampleValidArgon2idHash = "$argon2id$v=19$m=65536,t=3,p=2$c29tZXNhbHQ$RdescudvJCsgt3ub+b+dWRWJTmaaJObG"

func TestUserRepository(t *testing.T) {
	setEnvErr := os.Setenv("DB_DATABASE", "proxydb")
	if setEnvErr != nil {
		t.Fatal(setEnvErr)
	}

	db, cleanup := prepareDb(t)
	defer cleanup()
	defer func(db *sql.DB) {
		_ = db.Close()
	}(db)

	cache, err := NewBigCacheUserRepositoryCache(15*time.Minute, 5*time.Minute, 16, 512)
	if err != nil {
		t.Fatal(err)
	}

	repo := NewUserRepository(db, cache, mocks.NewMockMessageBusService())

	t.Run("GetByUsername", func(t *testing.T) {
		userId := insertTestUser(repo, t)
		user, err := repo.GetById(userId)
		assertNoError(t, err, "Failed to load user by Id")
		loadedUser, err := repo.GetByUsername(user.Username())
		assertNoError(t, err, "Failed to load user by Username")
		assertUsersEqual(t, user, loadedUser)
	})

	t.Run("GetById", func(t *testing.T) {
		userId := insertTestUser(repo, t)
		user, GetErr := repo.GetById(userId)
		assertNoError(t, GetErr, "Failed to load user by Id")

		loadedUser, loadedUserErr := repo.GetById(user.Id())
		assertNoError(t, loadedUserErr, "Failed to load user by ID")
		assertUsersEqual(t, user, loadedUser)
	})

	t.Run("GetByEmail", func(t *testing.T) {
		userId := insertTestUser(repo, t)
		user, GetErr := repo.GetById(userId)
		assertNoError(t, GetErr, "Failed to load user by Id")

		loadedUser, loadedUserErr := repo.GetByEmail(user.Email())
		assertNoError(t, loadedUserErr, "Failed to load user by ID")
		assertUsersEqual(t, user, loadedUser)
	})

	t.Run("Insert", func(t *testing.T) {
		userId := insertTestUser(repo, t)
		user, userErr := repo.GetById(userId)
		assertNoError(t, userErr, "Failed to load user by Id")

		loadedUser, loadedUserErr := repo.GetByUsername(user.Username())
		assertNoError(t, loadedUserErr, "Failed to load inserted user")
		assertUsersEqual(t, user, loadedUser)
	})

	t.Run("Update", func(t *testing.T) {
		userId := insertTestUser(repo, t)
		user, userErr := repo.GetById(userId)
		assertNoError(t, userErr, "Failed to load user by Id")

		updatedUser, updatedUserErr := aggregates.NewUser(userId, "updated_user", "updated@example.com", "$argon2id$v=19$m=65536,t=3,p=2$c29tZXNhbHQ$RdescudvJCsgt5ub+b+dWRWJTmaaJObG")
		if updatedUserErr != nil {
			t.Fatal(updatedUserErr)
		}
		assertNoError(t, repo.Update(updatedUser), "Failed to update user")

		loadedUser, loadedUserErr := repo.GetById(user.Id())
		assertNoError(t, loadedUserErr, "Failed to load updated user")
		assertUsersNotEqual(t, user, loadedUser)
	})
}

func insertTestUser(repo *UserRepository, t *testing.T) int {
	username := fmt.Sprintf("test_user_%d", time.Now().UTC().UnixNano())
	email := fmt.Sprintf("%s@example.com", username)
	user, err := aggregates.NewUser(-1, username, email, sampleValidArgon2idHash)
	assertNoError(t, err, "Failed to create test user")
	id, err := repo.Create(user)
	assertNoError(t, err, "Failed to insert test user")
	return id
}

func assertUsersEqual(t *testing.T, expected, actual aggregates.User) {
	if expected.Username() != actual.Username() {
		t.Errorf("Expected Username %s, got %s", expected.Username(), actual.Username())
	}
	if expected.PasswordHash() != actual.PasswordHash() {
		t.Errorf("Expected password hash %x, got %x", expected.PasswordHash(), actual.PasswordHash())
	}
}

func assertUsersNotEqual(t *testing.T, expected, actual aggregates.User) {
	if expected.Username() == actual.Username() {
		t.Errorf("Unexpected equal usernames: %s", expected.Username())
	}
	if expected.PasswordHash() == actual.PasswordHash() {
		t.Errorf("Unexpected equal password hash %v and %v", expected.PasswordHash(), actual.PasswordHash())
	}
}
