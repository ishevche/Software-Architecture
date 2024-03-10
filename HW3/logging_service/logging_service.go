package main

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/hazelcast/hazelcast-go-client"
	"golang.org/x/exp/maps"
	"log"
	"net/http"
	"os"
)

var hzMap *hazelcast.Map
var ctx context.Context

func getLogs(c *gin.Context) {
	data, err := hzMap.GetValues(ctx)
	if err != nil {
		log.Printf("Error occurred: %e\n", err)
		c.String(http.StatusInternalServerError, "Error occurred: %e\n", err)
		return
	}
	c.JSON(http.StatusOK, data)
}

func addLog(c *gin.Context) {
	var newMsg map[string]string

	if err := c.BindJSON(&newMsg); err != nil {
		log.Printf("Error occurred: %e\n", err)
		c.String(http.StatusInternalServerError, "Error occurred: %e\n", err)
		return
	}
	key := maps.Keys(newMsg)[0]
	value := maps.Values(newMsg)[0]
	log.Printf("New message with id=%s and value=%s", key, value)

	if _, err := hzMap.Put(ctx, key, value); err != nil {
		log.Printf("Error occurred: %e\n", err)
		c.String(http.StatusInternalServerError, "Error occurred: %e\n", err)
		return
	}

	c.Status(http.StatusOK)
}

func main() {
	ctx = context.TODO()

	hzAddr, clusterName, mapName := getArgs()

	cfg := hazelcast.Config{}
	cfg.Cluster.Network.SetAddresses(hzAddr)
	cfg.Cluster.Name = clusterName
	hzClient, err := hazelcast.StartNewClientWithConfig(ctx, cfg)
	defer shutdownClient(hzClient, ctx)
	if err != nil {
		log.Fatalf("Error occurred: %e\n", err)
	}

	hzMap, err = hzClient.GetMap(ctx, mapName)
	if err != nil {
		log.Fatalf("Error occurred: %e\n", err)
	}

	router := gin.Default()

	router.GET("/log", getLogs)
	router.POST("/log", addLog)

	if err := router.Run(":8080"); err != nil {
		log.Fatalf("Error occurred: %e\n", err)
	}
}

func shutdownClient(hzClient *hazelcast.Client, ctx context.Context) {
	err := hzClient.Shutdown(ctx)
	if err != nil {
		log.Fatalf("Failed to shutdown hazelcast client: %e\n", err)
	}
}

func getArgs() (string, string, string) {
	hzAddr := os.Getenv("HZ_ADDRESS")
	if hzAddr == "" {
		panic("Hazelcast address is not set up")
	}
	clsName := os.Getenv("HZ_CLUSTER_NAME")
	if clsName == "" {
		clsName = "hw3_shevchenko_hazelcast"
	}
	mapName := os.Getenv("HZ_MAP_NAME")
	if mapName == "" {
		mapName = "map"
	}
	return hzAddr, clsName, mapName
}
