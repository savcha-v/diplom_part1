package store

import (
	"context"
	"database/sql"
	"diplom_part1/internal/config"
	"net/http"
	"sync"
	"time"

	_ "github.com/jackc/pgx/v4/stdlib"
)

func (db *DB) GetBalanseSpent(ctx context.Context, userID string) (balance float32, spent float32, err error) {

	textQuery := `SELECT max(users."balanse"), sum(COALESCE(subtract."sum",0))
	FROM users left join subtract on users."userID" = subtract."userID"
	where users."userID" = $1`

	err = db.Connect.QueryRowContext(ctx, textQuery, userID).Scan(&balance, &spent)
	return
}

func (db *DB) AddOrder(ctx context.Context, order string, userID string, chanOrdersProc chan string) (int, string) {
	mu := &sync.Mutex{}
	mu.Lock()
	defer mu.Unlock()

	var receivedUserID string

	textQuery := `SELECT "userID" FROM accum WHERE "order" = $1`
	err := db.Connect.QueryRowContext(ctx, textQuery, order).Scan(&receivedUserID)

	switch {
	case err == sql.ErrNoRows:
		// add in db
		textInsert := `
		INSERT INTO accum ("userID", "order", "sum", "date", "status")
		VALUES ($1, $2, $3, $4, $5)`
		_, err = db.Connect.ExecContext(ctx, textInsert, userID, order, 0, time.Now(), "NEW")

		if err != nil {
			return http.StatusInternalServerError, ""
		}

		return http.StatusAccepted, order
	case err != nil:
		return http.StatusInternalServerError, ""
	case receivedUserID != userID:
		return http.StatusConflict, ""
	default:
		return http.StatusOK, ""
	}
}

func (db *DB) GetAccum(ctx context.Context, userID string) ([]config.OutAccum, error) {

	textQuery := `SELECT "order", "sum", "date", "status"
	FROM  accum 
	where "userID" = $1 ORDER BY "date"`

	var out []config.OutAccum

	rows, err := db.Connect.QueryContext(ctx, textQuery, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var item config.OutAccum
		err = rows.Scan(&item.Order, &item.Sum, &item.Date, &item.Status)
		if err != nil {
			return nil, err
		}

		out = append(out, item)
	}

	return out, err
}

func (db *DB) WriteWithdraw(ctx context.Context, order string, sum float32, userID string) int {

	mu := &sync.Mutex{}
	mu.Lock()
	defer mu.Unlock()

	tx, err := db.Connect.Begin()
	if err != nil {
		return http.StatusInternalServerError
	}
	defer tx.Rollback()

	balance, _, err := db.GetBalanseSpent(ctx, userID)
	if err != nil {
		return http.StatusInternalServerError
	}

	if balance < sum {
		return http.StatusPaymentRequired
	}

	// add in db
	textInsert := `
		INSERT INTO subtract ("userID", "order", "sum", "date")
		VALUES ($1, $2, $3, $4)`
	_, err = db.Connect.ExecContext(ctx, textInsert, userID, order, sum, time.Now())

	if err != nil {
		return http.StatusInternalServerError
	}

	textInsert = `
		UPDATE users set "balanse" = "balanse" - $1 where "userID" = $2`
	_, err = db.Connect.ExecContext(ctx, textInsert, sum, userID)
	if err != nil {
		return http.StatusInternalServerError
	}

	tx.Commit()

	return http.StatusOK
}

func (db *DB) GetWithdrawals(ctx context.Context, userID string) ([]config.OutWithdrawals, error) {

	textQuery := `SELECT "order", "sum", "date"
	FROM  subtruct 
	where "userID" = $1 ORDER BY "date"`

	var out []config.OutWithdrawals

	rows, err := db.Connect.QueryContext(ctx, textQuery, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	err = rows.Err()
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var item config.OutWithdrawals
		err = rows.Scan(&item.Order, &item.Sum, &item.Date)
		if err != nil {
			return nil, err
		}

		out = append(out, item)
	}

	return out, err
}
