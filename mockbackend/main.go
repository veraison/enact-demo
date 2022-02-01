package main

import (
	"github.com/gin-gonic/gin"
)

type PostNodePem struct {
	AK_PUB string
	EK_PUB string
}

func main() {
	r := gin.Default()

	r.POST("/node/pem", func(c *gin.Context) {
		// var reqBody PostNodePem

		// if err := c.BindJSON(&reqBody); err != nil {
		// 	log.Println(err.Error())
		// }

		// c.JSON(200, gin.H{
		// 	"ak_pub": reqBody.AK_PUB,
		// 	"ek_pub": reqBody.EK_PUB,
		// })
		ak_pub := c.PostForm("ak_pub")
		ek_pub := c.PostForm("ek_pub")

		c.JSON(200, gin.H{
			"ak_pub":         ak_pub,
			"size_of_ak_pub": len(ak_pub),
			"ek_pub":         ek_pub,
			"size_of_ek_pub": len(ek_pub),
		})
	})
	r.Run(":8000")
}
