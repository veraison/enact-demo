// Copyright 2023 EnactTrust LTD All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and

package db

import "github.com/jmoiron/sqlx"

const create_schema string = `
	CREATE TABLE IF NOT EXISTS nodes (
		id STRING NOT NULL PRIMARY KEY,
		ak_pub STRING,
		ek_pub STRING,
		created_at STRING,
		in_good_state INTEGER
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
