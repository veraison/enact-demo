package db

import "github.com/jmoiron/sqlx"

const create_schema string = `
	CREATE TABLE IF NOT EXISTS nodes (
		id STRING NOT NULL PRIMARY KEY,
		ak_pub STRING,
		ek_pub STRING,
		created_at STRING;
	);`

func InitDatabaseConnection() (*sqlx.DB, error) {
	db, err := sqlx.Open("sqlite3", "./enact.db")
	if err != nil {
		return nil, err
	}
	if _, err := db.Exec(create_schema); err != nil {
		return nil, err
	}

	return db, nil
}
