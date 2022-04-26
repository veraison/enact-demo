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
			c.JSON(400, gin.H{
				"error": err.Error(),
			})
		}

		ek_pub, err := c.FormFile("ek_pub")
		if err != nil {
			log.Println(err.Error())
			c.JSON(400, gin.H{
				"error": err.Error(),
			})
		}

		// Allocate buffers, so we can read the files
		ak_pub_buf := bytes.NewBuffer(nil)
		ek_pub_buf := bytes.NewBuffer(nil)

		// Get multipart.File from FileHeader
		ak_file, err := ak_pub.Open()
		if err != nil {
			log.Println(err.Error())
			c.JSON(400, gin.H{
				"error": err.Error(),
			})
		}
		ek_file, err := ek_pub.Open()
		if err != nil {
			log.Println(err.Error())
			c.JSON(400, gin.H{
				"error": err.Error(),
			})
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
			c.String(201, nodeID.String())
		}
	})

	// Note: ./agent onboard -> sends PEM, then sends GOLDEN
	// ./agent -> sends EVIDENCE

	r.POST("/node/golden", func(c *gin.Context) {
		// body param
		nodeID := c.PostForm("node_id")

		// Read POST submitted files - this gets their file headers
		golden_blob, err := c.FormFile("golden_blob")
		if err != nil {
			log.Println(err.Error())
			c.JSON(400, gin.H{
				"error": err.Error(),
			})
		}
		signature_blob, err := c.FormFile("signature_blob")
		if err != nil {
			log.Println(err.Error())
			c.JSON(400, gin.H{
				"error": err.Error(),
			})
		}

		// Allocate buffers, so we can read the files
		golden_blob_buf := bytes.NewBuffer(nil)

		// Get multipart.File from FileHeader
		golden_blob_file, err := golden_blob.Open()
		if err != nil {
			log.Println(err.Error())
		}

		// Read the file into a buffer
		_, err = io.Copy(golden_blob_buf, golden_blob_file)
		if err != nil {
			log.Println(err.Error())
		}

		signature_blob_buf := bytes.NewBuffer(nil)
		signature_blob_file, err := signature_blob.Open()
		if err != nil {
			log.Println(err.Error())
			c.JSON(400, gin.H{
				"error": err.Error(),
			})
		}
		_, err = io.Copy(signature_blob_buf, signature_blob_file)
		if err != nil {
			log.Println(err.Error())
			c.JSON(400, gin.H{
				"error": err.Error(),
			})
		}

		err = nodeService.HandleGoldenValue(nodeID, golden_blob_buf, signature_blob_buf)
		if err != nil {
			log.Println(err.Error())
			c.JSON(500, gin.H{
				"error": err.Error(),
			})
		} else {
			c.Status(201)
		}
	})

	r.POST("/node/evidence", func(c *gin.Context) {
		// Read POST submitted files - this gets their file headers
		nodeID := c.PostForm("node_id")

		evidence_blob, err := c.FormFile("evidence_blob")
		if err != nil {
			log.Println(err.Error())
			c.JSON(400, gin.H{
				"error": err.Error(),
			})
		}
		signature_blob, err := c.FormFile("signature_blob")
		if err != nil {
			log.Println(err.Error())
			c.JSON(400, gin.H{
				"error": err.Error(),
			})
		}

		// Allocate buffers, so we can read the files
		evidence_blob_buf := bytes.NewBuffer(nil)

		// Get multipart.File from FileHeader
		evidence_blob_file, err := evidence_blob.Open()
		if err != nil {
			log.Println(err.Error())
		}

		// Read the file into a buffer
		_, err = io.Copy(evidence_blob_buf, evidence_blob_file)
		if err != nil {
			log.Println(err.Error())
		}

		signature_blob_buf := bytes.NewBuffer(nil)
		signature_blob_file, err := signature_blob.Open()
		if err != nil {
			log.Println(err.Error())
			c.JSON(400, gin.H{
				"error": err.Error(),
			})
		}
		_, err = io.Copy(signature_blob_buf, signature_blob_file)
		if err != nil {
			log.Println(err.Error())
			c.JSON(400, gin.H{
				"error": err.Error(),
			})
		}

		err = nodeService.HandleEvidence(nodeID, evidence_blob_buf, signature_blob_buf)
		if err != nil {
			log.Println(err.Error())
			c.JSON(500, gin.H{
				"error": err.Error(),
			})
		} else {
			c.Status(201)
		}
	})

	return r
}

func main() {
	nodeService := setupServices()

	gin := setupRoutes(nodeService)

	gin.Run(":8000")
}
