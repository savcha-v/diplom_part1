package store

import (
	"database/sql"
	"diplom_part1/internal/config"
)

type DB struct {
	Connect *sql.DB
}

func DBInit(config config.Config) (db *DB, err error) {

	connect, err := sql.Open("pgx", config.DataBase)
	if err != nil {
		return nil, err
	}

	// users
	textCreate := `CREATE TABLE IF NOT EXISTS users(
		"userID" TEXT,
		"login" TEXT PRIMARY KEY,
		"hash" TEXT,
		"balanse" FLOAT 
		 );`
	if _, err := connect.Exec(textCreate); err != nil {
		return nil, err
	}
	// accumulation
	textCreate = `CREATE TABLE IF NOT EXISTS accum(
		"userID" TEXT,
		"order" TEXT PRIMARY KEY,
		"sum" FLOAT,
		"date" DATE,
		"status" TEXT
		 );`
	if _, err := connect.Exec(textCreate); err != nil {
		return nil, err
	}
	// subtraction
	textCreate = `CREATE TABLE IF NOT EXISTS subtract(
		"userID" TEXT,
		"order" TEXT PRIMARY KEY,
		"sum" FLOAT,
		"date" DATE
		 );`
	if _, err := connect.Exec(textCreate); err != nil {
		return nil, err
	}
	db = &DB{Connect: connect}
	return db, nil
}
