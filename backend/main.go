package main

import (
	"bytes"
	"io"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/veraison/enact-demo/pkg/db"
	"github.com/veraison/enact-demo/pkg/node"
)

func setupServices() *node.NodeService {
	// TODO: Load env vars from config

	// DB setup
	db, err := db.InitDatabaseConnection()
	if err != nil {
		log.Println(err.Error())
	}

	// Init repos
	nodeRepo := node.NewNodeRepo(db)

	// Init services (domains) and pass repos to them
	nodeService := node.NewService(nodeRepo)

	return nodeService
}

// TODO: Move into controller
func setupRoutes(nodeService *node.NodeService) *gin.Engine {
	// Init with the Logger and Recovery middleware already attached
	r := gin.Default()

	r.POST("/node/pem", func(c *gin.Context) {
		// Read POST submitted files - this gets their file headers
		ak_pub, err := c.FormFile("ak_pub")
		if err != nil {
			log.Println(err.Error())
		}

		ek_pub, err := c.FormFile("ek_pub")
		if err != nil {
			log.Println(err.Error())
		}

		// Allocate buffers, so we can read the files
		ak_pub_buf := bytes.NewBuffer(nil)
		ek_pub_buf := bytes.NewBuffer(nil)

		// Get multipart.File from FileHeader
		ak_file, err := ak_pub.Open()
		if err != nil {
			log.Println(err.Error())
		}
		ek_file, err := ek_pub.Open()
		if err != nil {
			log.Println(err.Error())
		}

		// Read the file into a buffer
		_, err = io.Copy(ak_pub_buf, ak_file)
		if err != nil {
			log.Println(err.Error())
		}
		_, err = io.Copy(ek_pub_buf, ek_file)
		if err != nil {
			log.Println(err.Error())
		}

		// Handle first step of node onboarding
		nodeID, err := nodeService.HandleReceivePEM(ak_pub_buf.String(), ek_pub_buf.String())
		if err != nil {
			log.Println(err.Error())
			c.JSON(500, gin.H{
				"error": err.Error(),
			})
		} else {
			c.JSON(201, gin.H{
				"node_id": nodeID,
			})
		}
	})

	return r
}

func main() {
	nodeService := setupServices()

	gin := setupRoutes(nodeService)

	gin.Run(":8000")
}
