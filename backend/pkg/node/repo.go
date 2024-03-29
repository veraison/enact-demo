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

package node

import (
	"database/sql"
	"errors"
	"log"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

type NodeRepository interface {
	InsertNode(node Node) error
	ListNodes() ([]Node, error)
	GetNodeById(node_id string) (*Node, error)
}

type SQLiteNodeRepo struct {
	db *sqlx.DB
}

var (
	ErrNotFound = errors.New("not found")
)

func NewNodeRepo(db *sqlx.DB) NodeRepository {
	return &SQLiteNodeRepo{
		db: db,
	}
}

func (repo SQLiteNodeRepo) InsertNode(node Node) error {
	const query = `
		INSERT INTO nodes (
			id,
			ak_pub,
			ek_pub,
			created_at
		)
		VALUES (
			:id,
			:ak_pub,
			:ek_pub,
			:created_at
		);`

	log.Println(`db`, repo.db)

	statement, err := repo.db.PrepareNamed(query)
	if err != nil {
		log.Println(err.Error())
		return err
	}

	response, err := statement.Exec(&node)
	if err != nil {
		log.Println(err.Error())
		return err
	}

	if count, err := response.RowsAffected(); err != nil {
		log.Println("Could not check RowsAffected", err)
		return err
	} else if count == 0 {
		log.Println("Nothing inserted", err)
		return err
	}

	return nil
}

func (repo SQLiteNodeRepo) ListNodes() ([]Node, error) {
	var nodes_list []Node = []Node{}

	return nodes_list, nil
}

func (repo SQLiteNodeRepo) GetNodeById(node_id string) (*Node, error) {
	node := Node{}

	const query = `
		SELECT
			id,
			ak_pub,
			ek_pub,
			created_at
		FROM nodes
		WHERE id = $1;
		);`

	statement, err := repo.db.Preparex(query)
	if err != nil {
		return nil, err
	}

	err = statement.Get(&node, node_id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}

		return nil, err
	}

	return &node, nil
}
