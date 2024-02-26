package main

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/hazelcast/hazelcast-go-client"
	"os"
)

func getInsertArgs() (string, string) {
	switch len(os.Args) {
	case 2:
		return os.Args[1], "map"
	case 3:
		return os.Args[1], os.Args[2]
	default:
		return "hw2_shevchenko_hazelcast", "map"
	}
}

func main() {
	ctx := context.TODO()
	cfg := hazelcast.Config{}
	clusterName, mapName := getInsertArgs()
	cfg.Cluster.Name = clusterName

	hzClient, _ := hazelcast.StartNewClientWithConfig(ctx, cfg)
	defer hzClient.Shutdown(ctx)

	myMap, _ := hzClient.GetMap(ctx, mapName)
	for i := range 1000 {
		myMap.Put(ctx, i, uuid.New())
	}

	fmt.Println("Successfully inserted 1000 key-value pairs in the map!")
}
