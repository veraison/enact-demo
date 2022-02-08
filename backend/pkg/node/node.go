package node

import (
	// "backend/pkg/enactcorim"
	"log"

	"enactcorim"

	"github.com/google/uuid"
)

type NodeService struct {
	repo NodeRepository
}

type Node struct {
	ID         uuid.UUID `db:"id"`
	AK_Pub     string    `db:"ak_pub"`
	EK_Pub     string    `db:"ek_pub"`
	Created_At string    `db:"created_at"`
}

func NewService(repo NodeRepository) *NodeService {
	return &NodeService{
		repo: repo,
	}
}

func (n *NodeService) HandleReceivePEM(akPub string, ekPub string) error {
	// 1. From the agent: `POST /node/pem, Body: { AK_pub, EK_pub }`
	// 2. Generate node_id (UUID v4)
	nodeID, err := uuid.NewUUID()
	if err != nil {
		log.Println("Error generating node UUID")
	}

	_ = nodeID

	// 3. Store node_id, AK_pub, EK_pub in the DB

	// 4. Repackage node_id and AK pub as CoRIM
	c, err := enactcorim.RepackageNodePEM(akPub, nodeID)
	if err != nil {
		log.Fatal(err)
		return err
	}

	out, err := c.ToCBOR()
	_ = out
	if err != nil {
		log.Fatal(err)
		return err
	}

	// 5. `POST /submit, Body: { CoRIM }` to veraison backend and forward response to agent

	return nil
}
