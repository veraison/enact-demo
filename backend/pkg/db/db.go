package db

import "database/sql"

const create_schema string = `
	CREATE TABLE IF NOT EXISTS nodes (
		id STRING NOT NULL PRIMARY KEY,
		ak_pub STRING,
		ek_pub STRING
		created_at STRING
	);`

func InitDatabaseConnection() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "./enact.db")
	if err != nil {
		return nil, err
	}
	if _, err := db.Exec(create_schema); err != nil {
		return nil, err
	}

	return db, nil
}
