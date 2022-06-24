package store

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
)

func (db *DB) LoginUse(ctx context.Context, login string) (bool, error) {
	var userID string

	textQuery := `SELECT "userID" FROM users WHERE "login" = $1`
	err := db.Connect.QueryRowContext(ctx, textQuery, login).Scan(&userID)

	switch {
	case err == sql.ErrNoRows:
		return false, nil
	case err != nil:
		return false, err
	default:
		return true, nil
	}
}

func (db *DB) WriteNewUser(ctx context.Context, login string, hash string) (string, error) {

	userID := uuid.New().String()
	textInsert := `
	INSERT INTO users ("userID", "login", "hash", "balanse")
	VALUES ($1, $2, $3, $4)`
	_, err := db.Connect.ExecContext(ctx, textInsert, userID, login, hash, 0)

	if err != nil {
		return "", err
	}

	return userID, nil
}

func (db *DB) ReadUser(ctx context.Context, login string, hash string) (string, error) {
	var userID string

	textQuery := `SELECT "userID" FROM users WHERE "login" = $1 AND "hash" = $2`
	err := db.Connect.QueryRowContext(ctx, textQuery, login, hash).Scan(&userID)

	switch {
	case err == sql.ErrNoRows:
		return "", nil
	case err != nil:
		return "", err
	default:
		return userID, nil
	}
}

func (db *DB) ExistsUserID(ctx context.Context, userID string) (bool, error) {
	var login string

	textQuery := `SELECT "login" FROM users WHERE "userID" = $1`
	err := db.Connect.QueryRowContext(ctx, textQuery, userID).Scan(&login)

	switch {
	case err == sql.ErrNoRows:
		return false, nil
	case err != nil:
		return false, err
	default:
		return true, nil
	}
}
