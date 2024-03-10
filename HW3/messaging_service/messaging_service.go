package main

import (
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

func getMessages(c *gin.Context) {
	c.JSON(http.StatusOK, "Messaging service is not implemented yet")
}

func main() {
	router := gin.Default()

	router.GET("/msg", getMessages)

	if err := router.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
