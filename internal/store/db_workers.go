package store

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
)

func (db *DB) GetUserID(ctx context.Context, number string) (string, error) {

	textQuery := `SELECT "userID" FROM accum WHERE "order" = $1`
	rows, err := db.Connect.QueryContext(ctx, textQuery, number)

	if err != nil {
		return "", errors.New("error get userID")
	}
	defer rows.Close()

	var userID string

	for rows.Next() {
		err = rows.Scan(&userID)
		if err != nil {
			return "", errors.New("error scan rows in db")
		}
	}

	err = rows.Err()
	if err != nil {
		return "", errors.New("rows error in db")
	}

	return userID, nil
}

func (db *DB) GetOrdersProcessing(ctx context.Context) ([]string, error) {

	fmt.Fprintln(os.Stdout, "getOrdersProcessing")

	textQuery := `SELECT "order"
	FROM  accum 
	where "status" = $1 or "status" = $2 or "status" = $3`

	var out []string
	// new, registered, processing
	rows, err := db.Connect.QueryContext(ctx, textQuery, "NEW", "PROCESSING", "REGISTERED")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	err = rows.Err()
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var item string
		err = rows.Scan(&item)
		if err != nil {
			return nil, err
		}

		out = append(out, item)
	}

	return out, err
}

func (db *DB) UpdateOrder(ctx context.Context, userID string, order string, status string, sum float32) (string, error) {

	fmt.Fprintln(os.Stdout, "updateOrder")

	mu := &sync.Mutex{}
	mu.Lock()
	defer mu.Unlock()

	// Начало транзацкции
	tx, err := db.Connect.Begin()
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

	if status == "PROCESSED" {

		textQuery := `UPDATE accum SET "sum" = $1, "status" = $2 WHERE "order" = $3`
		_, err = tx.ExecContext(ctx, textQuery, sum, status, order)

		if err != nil {
			return "", err
		}

		textQuery = `UPDATE users SET "balanse" = "balanse" + $1 WHERE "userID" = $2`
		_, err = tx.ExecContext(ctx, textQuery, sum, userID)
		if err != nil {
			return "", err
		}

	} else {
		textQuery := `UPDATE accum SET "status" = $1 WHERE "order" = $2`
		_, err = tx.ExecContext(ctx, textQuery, status, order)
		if err != nil {
			return "", err
		}
	}

	tx.Commit()

	return status, nil
}
