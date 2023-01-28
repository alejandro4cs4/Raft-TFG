package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

func clearAllClusterData(cli *clientv3.Client, measureElapsedTime bool) {
	var startTime time.Time
	var elapsedTime time.Duration

	// Start timer
	if measureElapsedTime {
		startTime = time.Now()
	}

	deleteResponse, err := cli.Delete(context.Background(), "", clientv3.WithPrefix())

	if err != nil {
		log.Panicf("Could not delete cluster data: %v\n", err)
	}

	// Stop timer
	if measureElapsedTime {
		elapsedTime = time.Since(startTime)
	}

	if measureElapsedTime {
		fmt.Printf("It took %d ms / %.2f sec to wipe cluster data\n", elapsedTime.Milliseconds(), elapsedTime.Seconds())
	}

	fmt.Printf("Successfully deleted %v keys from etcd cluster\n", deleteResponse.Deleted)
}

func clearClusterDataOneByOne(cli *clientv3.Client) {
	cmd := exec.Command("find", "/usr/")
	stdout, err := cmd.StdoutPipe()

	if err != nil {
		log.Panicf("cmd.StdoutPipe(): %v\n", err)
	}

	cmd.Start()

	scanner := bufio.NewScanner(stdout)

	// Start timer
	startTime := time.Now()

	for scanner.Scan() {
		key := scanner.Text()

		_, err := cli.Delete(context.Background(), key)

		if err != nil {
			log.Panicf("cli.Delete(%v): %v\n", key, err)
		}
	}

	// Stop timer
	elapsedTime := time.Since(startTime)

	cmd.Wait()

	fmt.Printf("It took %d ms / %.2f sec / %.2f min to delete the data one by one from etcd\n", elapsedTime.Milliseconds(), elapsedTime.Seconds(), elapsedTime.Minutes())
}

func printDataAmount() {
	numFilesCmd := "find /usr/ | wc -l"

	out, err := exec.Command("bash", "-c", numFilesCmd).Output()

	if err != nil {
		log.Panicf("exec.Command(find /usr/ | wc -l): %v\n", err)
	}

	fmt.Printf("%v key-value pairs will be stored in etcd cluster\n", string(out[:len(out)-1]))
}

func storeDataInCluster(cli *clientv3.Client) {
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

		_, err := cli.Put(context.Background(), key, value)

		if err != nil {
			log.Panicf("cli.Put(%v, %v): %v\n", key, value, err)
		}
	}

	// Stop timer
	elapsedTime := time.Since(startTime)

	cmd.Wait()

	fmt.Printf("It took %d ms / %.2f sec / %.2f min to store the data in etcd\n", elapsedTime.Milliseconds(), elapsedTime.Seconds(), elapsedTime.Minutes())
}

func listClusterData(cli *clientv3.Client) {
	file, createErr := os.Create("./clusterData.txt")

	if createErr != nil {
		log.Panicf("os.Create(): %v\n", createErr)
	}

	defer file.Close()

	// Start timer
	getDataStartTime := time.Now()

	getResponse, getErr := cli.Get(context.Background(), "", clientv3.WithPrefix())

	if getErr != nil {
		log.Panicf("cli.Get(): %v\n", getErr)
	}

	// Stop timer
	getDataElapsedTime := time.Since(getDataStartTime)

	fmt.Printf("It took %d ms / %.2f sec to get all data stored in etcd (%v entries)\n", getDataElapsedTime.Milliseconds(), getDataElapsedTime.Seconds(), getResponse.Count)

	// Start timer
	writeDataStartTime := time.Now()

	for index, kv := range getResponse.Kvs {
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

func getClient() (*clientv3.Client, error) {
	log.Default().Println("Connecting to etcd cluster...")

	config := clientv3.Config{
		Endpoints:   []string{"localhost:2379", "localhost:22379", "localhost:32379"},
		DialTimeout: 5 * time.Second,
	}

	errChan := make(chan error, 1)
	cliChan := make(chan *clientv3.Client, 1)

	go func() {
		cli, err := clientv3.New(config)

		if err != nil {
			errChan <- err
			return
		}

		statusRes, err := cli.Status(context.Background(), config.Endpoints[0])

		if err != nil {
			errChan <- err
			return
		} else if statusRes == nil {
			errChan <- errors.New("The status response form etcd was nil")
			return
		}

		cliChan <- cli
	}()

	select {
	case err := <-errChan:
		return nil, err
	case cli := <-cliChan:
		return cli, nil
	case <-time.After(6 * time.Second):
		return nil, errors.New("5 seconds timeout exceeded while trying to connect to etcd cluster")
	}
}

const PROGRAM_INFO = "\nType the number of the option you want to execute:\n\t1. Clear all cluster data at once\n\t2. Clear cluster data one by one\n\t3. Store all data in cluster\n\t4. List all data in cluster\n\t5. Exit the program\n\n"

func main() {
	cli, err := getClient()

	if err != nil {
		log.Fatalf("Could not connect to etcd cluster: %v\n", err)
	}

	defer cli.Close()

	log.Default().Println("Connected to etcd cluster successfully")

	// Read user input
	buf := bufio.NewReader(os.Stdin)

	fmt.Print(PROGRAM_INFO)

	for {
		fmt.Print("> ")

		input, err := buf.ReadBytes('\n')

		if err != nil {
			log.Fatalln("Error reading user input")
		}

		if string(input[:]) == "5\n" {
			break
		}

		fmt.Println()

		switch string(input[:]) {
		case "1\n":
			clearAllClusterData(cli, true)
			break
		case "2\n":
			clearClusterDataOneByOne(cli)
			break
		case "3\n":
			printDataAmount()
			storeDataInCluster(cli)
			break
		case "4\n":
			listClusterData(cli)
			break
		default:
			fmt.Printf("Unknown option\n%v", PROGRAM_INFO)
		}

		fmt.Println()
	}

	fmt.Println("\nExiting...")
}
