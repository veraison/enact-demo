package main

import (
	"bytes"
	"io"
	"log"

	"github.com/gin-gonic/gin"
)

type PostNodePem struct {
	AK_PUB string
	EK_PUB string
}

func main() {
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

		// Return submitted files and their length in bytes
		c.JSON(200, gin.H{
			"ak_pub":                 ak_pub_buf.String(),
			"size_of_ak_pub (bytes)": ak_pub.Size,
			"ek_pub":                 ek_pub_buf.String(),
			"size_of_ek_pub (bytes)": ek_pub.Size,
		})
	})
	r.Run(":8000")
}
