package main

import (
	"context"
	"fmt"
	"log"
	"os/exec"
	"strings"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

type KeyValue struct {
	Key string
	Value string
}

var systemData []KeyValue

func getDataInMemory() {
		out, err := exec.Command("find", "/usr/", "-printf", "%p:%y:%m:%U:%G:%A@:%s:%i\n").Output()

		if err != nil {
			log.Fatalf("exec.Command(): %s\n", err)
		}

		output := string(out[:])

		lines := strings.Split(output, "\n")

		for _, line := range lines {
			key, value, _ := strings.Cut(line, ":")

			newData := KeyValue{
				Key: key,
				Value: value,
			}

			systemData = append(systemData, newData)
		}

		// Remove last element as it is empty
		systemData = systemData[:len(systemData) - 1]
}

func saveDataInEtcd(cli *clientv3.Client) (elapsedTime time.Duration) {
	ctx := context.Background()

	startTime := time.Now()

	for index, data := range systemData {
		_, err := cli.Put(ctx, data.Key, data.Value)

		if err != nil {
			log.Fatalf("cli.Put() on index %v: %v\n", index, err)
		}
	}

	elapsedTime = time.Since(startTime)

	return
}

func main() {
	getDataInMemory()

	cli, err := clientv3.New(clientv3.Config{
		Endpoints: []string{"localhost:2379", "localhost:22379", "localhost:32379"},
		DialTimeout: 5 * time.Second,
	})

	if err != nil {
		log.Fatalln("Could not connect to etcd cluster")
	}

	log.Default().Println("Connected to etcd cluster successfully")

	elapsedTime := saveDataInEtcd(cli)

	fmt.Printf("It took %vms (%vs) to store the data in etcd\n", elapsedTime.Milliseconds(), elapsedTime.Seconds())

	defer cli.Close()
}