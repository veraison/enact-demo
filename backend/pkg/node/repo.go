package node

import (
	"database/sql"
	"errors"
	"log"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

type NodeRepository interface {
	CreateNewNode(node *Node) error
	ListNodes() ([]Node, error)
	GetNodeById(node_id string) (*Node, error)
}

type SQLiteNodeRepo struct {
	db *sqlx.DB
}

var (
	ErrNotFound = errors.New("Not Found")
)

func NewNodeRepo(db *sqlx.DB) NodeRepository {
	return &SQLiteNodeRepo{
		db: db,
	}
}

func (repo *SQLiteNodeRepo) CreateNewNode(node *Node) error {
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

	statement, err := repo.db.PrepareNamed(query)
	if err != nil {
		return err
	}

	response, err := statement.Exec(&node)
	if err != nil {
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

func (repo *SQLiteNodeRepo) ListNodes() ([]Node, error) {
	var nodes_list []Node = []Node{}

	return nodes_list, nil
}

func (repo *SQLiteNodeRepo) GetNodeById(node_id string) (*Node, error) {
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

	statement, err := repo.db.PrepareNamed(query)
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
