package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

func clearClusterData(cli *clientv3.Client, ctx context.Context, measureElapsedTime bool) {
	var startTime time.Time
	var elapsedTime time.Duration

	// Start timer
	if measureElapsedTime {
		startTime = time.Now()
	}

	_, err := cli.Delete(ctx, "", clientv3.WithPrefix())

	if err != nil {
		log.Panicf("Could not delete cluster data: %v\n", err)
	}

	// Stop timer
	if measureElapsedTime {
		elapsedTime = time.Since(startTime)
	}

	if measureElapsedTime {
		fmt.Printf("It took %d ms / %.2f sec to wipe cluster data\n", elapsedTime.Milliseconds(), elapsedTime.Seconds())

		return
	}

	log.Default().Println("Deleted cluster data successfully")
}

func printDataAmount() {
	numFilesCmd := "find /usr/ | wc -l"

	out, err := exec.Command("bash", "-c", numFilesCmd).Output()

	if err != nil {
		log.Panicf("exec.Command(find /usr/ | wc -l): %v\n", err)
	}

	fmt.Printf("%v files will be stored in etcd cluster\n", string(out[:len(out)-1]))
}

func storeDataInCluster(cli *clientv3.Client, ctx context.Context) {
	cmd := exec.Command("find", "/usr/", "-printf", "%p:%y:%m:%U:%G:%A@:%s:%i\n")
	stdout, err := cmd.StdoutPipe()

	if err != nil {
		log.Panicf("cmd.StdoutPipe(): %v\n", err)
	}

	cmd.Start()

	scanner := bufio.NewScanner(stdout)

	// Start timer
	startTime := time.Now()

	for scanner.Scan() {
		key, value, _ := strings.Cut(scanner.Text(), ":")

		_, err := cli.Put(ctx, key, value)

		if err != nil {
			log.Panicf("cli.Put(%v, %v): %v\n", key, value, err)
		}
	}

	// Stop timer
	elapsedTime := time.Since(startTime)

	cmd.Wait()

	fmt.Printf("It took %d ms / %.2f sec / %.2f min to store the data in etcd\n", elapsedTime.Milliseconds(), elapsedTime.Seconds(), elapsedTime.Minutes())
}

func listClusterData(cli *clientv3.Client, ctx context.Context) {
	file, createErr := os.Create("./clusterData.txt")

	if createErr != nil {
		log.Panicf("os.Create(): %v\n", createErr)
	}

	defer file.Close()

	// Start timer
	getDataStartTime := time.Now()

	resp, getErr := cli.Get(ctx, "", clientv3.WithPrefix())

	if getErr != nil {
		log.Panicf("cli.Get(): %v\n", getErr)
	}

	// Stop timer
	getDataElapsedTime := time.Since(getDataStartTime)

	fmt.Printf("It took %d ms / %.2f sec to get all data stored in etcd\n", getDataElapsedTime.Milliseconds(), getDataElapsedTime.Seconds())

	// Start timer
	writeDataStartTime := time.Now()

	for index, kv := range resp.Kvs {
		line := fmt.Sprintf("%v. %v - %v\n", index, string(kv.Key[:]), string(kv.Value[:]))

		_, writeErr := file.WriteString(line)

		if writeErr != nil {
			log.Default().Printf("Error writing to clusterData.txt: %v\n", writeErr)
		}
	}

	// Stop timer
	writeDataElapsedTime := time.Since(writeDataStartTime)

	fmt.Printf("It took %d ms / %.2f sec / %.2f min to write all data stored in etcd to clusterData.txt\n", writeDataElapsedTime.Milliseconds(), writeDataElapsedTime.Seconds(), writeDataElapsedTime.Minutes())
}

func main() {
	// Connect to etcd cluster
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"localhost:2379", "localhost:22379", "localhost:32379"},
		DialTimeout: 5 * time.Second,
	})

	if err != nil {
		log.Panicf("Could not connect to etcd cluster: %v\n", err)
	}

	defer cli.Close()

	log.Default().Println("Connected to etcd cluster successfully")

	ctx := context.Background()

	clearClusterData(cli, ctx, false)

	printDataAmount()

	storeDataInCluster(cli, ctx)

	listClusterData(cli, ctx)

	clearClusterData(cli, ctx, true)
}
