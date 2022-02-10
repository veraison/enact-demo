package node

import (
	// "backend/pkg/enactcorim"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/veraison/enact-demo/pkg/enactcorim"
	"github.com/veraison/enact-demo/pkg/veraison"
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
		return err
	}

	log.Println(nodeID)

	// 3. Init node entity and store it in the db
	node := Node{
		ID:         nodeID,
		AK_Pub:     akPub,
		EK_Pub:     ekPub,
		Created_At: time.Now().UTC().String(),
	}

	log.Println(nodeID)

	err = n.repo.CreateNewNode(&node)
	if err != nil {
		log.Println(err.Error())
		return err
	}

	// 4. Repackage node_id and AK pub as CoRIM
	corim, err := enactcorim.RepackageNodePEM(akPub, nodeID)
	if err != nil {
		log.Fatal(err)
		return err
	}

	log.Println(corim)

	// out, err := c.ToCBOR()
	// _ = out
	// if err != nil {
	// 	log.Fatal(err)
	// 	return err
	// }

	// 5. `POST /submit, Body: { CoRIM }` to veraison backend and forward response to agent
	err = veraison.SendPEMToVeraison(corim)
	if err != nil {
		log.Fatal(err)
		return err
	}

	return nil
}
