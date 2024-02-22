package main

import (
	"github.com/gin-gonic/gin"
	"golang.org/x/exp/maps"
	"log"
	"net/http"
)

var messages = make(map[string]string)

func getLogs(c *gin.Context) {
	c.JSON(http.StatusOK, maps.Values(messages))
}

func addLog(c *gin.Context) {
	var newMsg map[string]string

	if err := c.BindJSON(&newMsg); err != nil {
		log.Fatal(err)
		return
	}

	log.Printf("New message with id=%s and value=%s",
		maps.Keys(newMsg)[0], maps.Values(newMsg)[0])
	maps.Copy(messages, newMsg)

	c.Status(http.StatusOK)
}

func main() {
	router := gin.Default()

	router.GET("/log", getLogs)
	router.POST("/log", addLog)

	if err := router.Run(":8081"); err != nil {
		log.Fatal(err)
	}
}
