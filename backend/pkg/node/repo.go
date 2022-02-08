package node

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

type NodeRepository interface {
	CreateNewNode(node *Node) (string, error)
	ListNodes() ([]*Node, error)
	GetNodeById(node_id string) (*Node, error)
}

type SQLiteNodeRepo struct {
	db *sql.DB
}

func NewNodeRepo(db *sql.DB) *SQLiteNodeRepo {
	return &SQLiteNodeRepo{
		db: db,
	}
}

func (repo *SQLiteNodeRepo) CreateNewNode(node *Node) (string, error) {
}

func (repo *SQLiteNodeRepo) ListNodes() ([]*Node, error) {
}

func (repo *SQLiteNodeRepo) GetNodeById(node_id string) (*Node, error) {

}
