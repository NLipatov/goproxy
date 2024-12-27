package Repositories

import (
	"bytes"
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"goproxy/DataAccess"
	"goproxy/Domain/Aggregates"
	"testing"
	"time"
)

func TestUserRepository(t *testing.T) {
	db, cleanup := prepareDb(t)
	defer cleanup()
	defer func(db *sql.DB) {
		_ = db.Close()
	}(db)

	repo := NewUserRepository(db)

	t.Run("GetByUsername", func(t *testing.T) {
		userId := insertTestUser(repo, t)
		user, err := repo.GetById(userId)
		assertNoError(t, err, "Failed to load user by id")
		loadedUser, err := repo.GetByUsername(user.Username())
		assertNoError(t, err, "Failed to load user by username")
		assertUsersEqual(t, user, loadedUser)
	})

	t.Run("GetById", func(t *testing.T) {
		userId := insertTestUser(repo, t)
		user, err := repo.GetById(userId)
		assertNoError(t, err, "Failed to load user by id")
		loadedUser, err := repo.GetById(user.Id())
		assertNoError(t, err, "Failed to load user by ID")
		assertUsersEqual(t, user, loadedUser)
	})

	t.Run("Insert", func(t *testing.T) {
		userId := insertTestUser(repo, t)
		user, err := repo.GetById(userId)
		assertNoError(t, err, "Failed to load user by id")
		loadedUser, err := repo.GetByUsername(user.Username())
		assertNoError(t, err, "Failed to load inserted user")
		assertUsersEqual(t, user, loadedUser)
	})

	t.Run("Update", func(t *testing.T) {
		userId := insertTestUser(repo, t)
		user, err := repo.GetById(userId)
		assertNoError(t, err, "Failed to load user by id")
		updatedUser, _ := Aggregates.NewUser(userId, "updated_user", []byte("new_hash"), []byte("new_salt"))
		assertNoError(t, repo.Update(updatedUser), "Failed to update user")
		loadedUser, err := repo.GetById(user.Id())
		assertNoError(t, err, "Failed to load updated user")
		assertUsersNotEqual(t, user, loadedUser)
	})
}

func insertTestUser(repo *UserRepository, t *testing.T) int {
	username := fmt.Sprintf("test_user_%d", time.Now().UnixNano())
	user, err := Aggregates.NewUser(-1, username, []byte("hashed_password"), []byte("salt"))
	assertNoError(t, err, "Failed to create test user aggregate")
	id, err := repo.Create(user)
	assertNoError(t, err, "Failed to insert test user")
	return id
}

func assertNoError(t *testing.T, err error, message string) {
	if err != nil {
		t.Fatalf("%s: %v", message, err)
	}
}

func assertUsersEqual(t *testing.T, expected, actual Aggregates.User) {
	if expected.Username() != actual.Username() {
		t.Errorf("Expected username %s, got %s", expected.Username(), actual.Username())
	}
	if !bytes.Equal(expected.PasswordHash(), actual.PasswordHash()) {
		t.Errorf("Expected password hash %x, got %x", expected.PasswordHash(), actual.PasswordHash())
	}
	if !bytes.Equal(expected.PasswordSalt(), actual.PasswordSalt()) {
		t.Errorf("Expected password salt %x, got %x", expected.PasswordSalt(), actual.PasswordSalt())
	}
}

func assertUsersNotEqual(t *testing.T, expected, actual Aggregates.User) {
	if expected.Username() == actual.Username() {
		t.Errorf("Unexpected equal usernames: %s", expected.Username())
	}
	if bytes.Equal(expected.PasswordHash(), actual.PasswordHash()) {
		t.Errorf("Unexpected equal password hashes: %x", expected.PasswordHash())
	}
	if bytes.Equal(expected.PasswordSalt(), actual.PasswordSalt()) {
		t.Errorf("Unexpected equal password salts: %x", expected.PasswordSalt())
	}
}

func prepareDb(t *testing.T) (*sql.DB, func()) {
	_, db, cleanup := data_access.SetupPostgresContainer(t)
	data_access.Migrate(db)

	return db, cleanup
}