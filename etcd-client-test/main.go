package main

import (
	"context"
	"fmt"
	"log"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

func registerKeyValue(cli *clientv3.Client) {
	ctx := context.Background()

	_, err := cli.Put(ctx, "test_key", "test_value")

	if err != nil {
		log.Fatalln("Put():", err)
	}

	// Get registered key - value
	resp, err := cli.Get(ctx, "test_key")

	if err != nil {
		log.Fatalln("Get():", err)
	}

	if len(resp.Kvs) == 0 {
		fmt.Println("There are no values associated to the provided key")
	} else {
		for _, item := range resp.Kvs {
			fmt.Printf("Key: %v - Value: %v\n", string(item.Key), string(item.Value))
		}
	}
}

func main() {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints: []string{"localhost:2379", "localhost:22379", "localhost:32379"},
		DialTimeout: 5 * time.Second,
	})

	if err != nil {
		log.Fatalln("Could not connect to etcd cluster")
	}

	log.Default().Println("Connected to etcd cluster successfully")

	registerKeyValue(cli)

	defer cli.Close()
}