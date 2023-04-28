package main

import (
	"bytes"
	"io"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/veraison/apiclient/verification"
	"github.com/veraison/enact-demo/pkg/db"
	"github.com/veraison/enact-demo/pkg/node"
	"github.com/veraison/enact-demo/pkg/veraison"
)

var (
	EntryPoint           = "http://localhost:8080/challenge-response/v1/newSession"
	VeraisonSessionTable = map[string]string{}
	FakeNodeID           = "7dd5db06-d2f5-4e0d-8a9c-9baaa5a446ef"
	FakeGolden           = []byte{0x00, 0x01, 0x02, 0x03}
	TPMEvidenceMediaType = "application/vnd.enacttrust.tpm-evidence"
)

var globalCfg *verification.ChallengeResponseConfig

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

		ak_name := c.PostForm("ak_name")
		log.Println(ak_name)

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
		// TODO: SESSION - store ak_name and ek_pub, so we can use them in /node/secret
		// Handle first step of node onboarding
		nodeID, err := nodeService.HandleReceivePEM(ak_pub_buf.String(), ek_pub_buf.String())
		if err != nil {
			log.Println(err.Error())
			c.JSON(500, gin.H{
				"error": err.Error(),
			})
		} else {
			log.Println(nodeID.String())
			c.String(201, nodeID.String())
		}
	})

	r.POST("/node/secret", func(c *gin.Context) {
		nodeID := c.PostForm("node_id")

		// 1. call Veraison frontend
		// 2. TODO: encrypt the nonce (the challenge) returned by Veraison
		// 3. store the session_id (regenerated on every call to /session) to make calls later
		cfg, sessionCtx, sessionURI, err := veraison.CreateVeraisonSession()

		globalCfg = cfg

		if err != nil {
			log.Println(err.Error())
			// 500 = Session or challenge creation failed
			c.JSON(500, gin.H{
				"error": err.Error(),
			})
		} else {
			log.Println("nonce:", sessionCtx.Nonce)
			log.Println(`sessionURI: `, sessionURI)

			// store session_id and associate it with node_id, so we can use it later to call Veraison
			VeraisonSessionTable[nodeID] = sessionURI

			// On success -> write 204 and return the challenge []byte to the agent
			c.Writer.WriteHeader(204)
			c.Writer.Write(sessionCtx.Nonce)
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

		evidenceDigest, err := nodeService.HandleGoldenValue(nodeID, golden_blob_buf, signature_blob_buf)
		if err != nil {
			log.Println(err.Error())
			c.JSON(500, gin.H{
				"error": err.Error(),
			})
		} else {
			err = nodeService.RouteEvidenceToVeraison(globalCfg, VeraisonSessionTable[nodeID], nodeID, golden_blob_buf, signature_blob_buf, evidenceDigest)
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
