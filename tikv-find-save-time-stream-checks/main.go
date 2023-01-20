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

	"github.com/tikv/client-go/v2/config"
	"github.com/tikv/client-go/v2/rawkv"
)

func clearAllClusterData(cli *rawkv.Client, measureElapsedTime bool) {
	var startTime time.Time
	var elapsedTime time.Duration

	// Start timer
	if measureElapsedTime {
		startTime = time.Now()
	}

	err := cli.DeleteRange(context.Background(), []byte("/"), nil)

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

	// fmt.Printf("Successfully deleted %v keys from TiKV cluster\n", deleteResponse.Deleted)
}

func clearClusterDataOneByOne(cli *rawkv.Client) {
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

		err := cli.Delete(context.Background(), []byte(key))

		if err != nil {
			log.Panicf("cli.Delete(%v): %v\n", key, err)
		}
	}

	// Stop timer
	elapsedTime := time.Since(startTime)

	cmd.Wait()

	fmt.Printf("It took %d ms / %.2f sec / %.2f min to delete the data one by one from TiKV\n", elapsedTime.Milliseconds(), elapsedTime.Seconds(), elapsedTime.Minutes())
}

func printKVS(keys [][]byte, values [][]byte) {
	for index, key := range keys {
		fmt.Printf("%v - %v\n", string(key[:]), string(values[index][:]))
	}
}

func printFirstLastKeys(keys [][]byte) {
	if len(keys) == 0 {
		return
	}

	fmt.Printf("First: %v\n", string(keys[0][:]))

	if len(keys) > 1 {
		fmt.Printf("Last: %v\n\n", string(keys[len(keys)-1][:]))
	} else {
		fmt.Println()
	}
}

func printDataAmount() {
	numFilesCmd := "find /usr/ | wc -l"

	out, err := exec.Command("bash", "-c", numFilesCmd).Output()

	if err != nil {
		log.Panicf("exec.Command(find /usr/ | wc -l): %v\n", err)
	}

	fmt.Printf("%v key-value pairs will be stored in TiKV cluster\n", string(out[:len(out)-1]))
}

func storeDataInCluster(cli *rawkv.Client) {
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

		err := cli.Put(context.Background(), []byte(key), []byte(value))

		if err != nil {
			log.Panicf("cli.Put(%v, %v): %v\n", key, value, err)
		}
	}

	// Stop timer
	elapsedTime := time.Since(startTime)

	cmd.Wait()

	fmt.Printf("It took %d ms / %.2f sec / %.2f min to store the data in TiKV\n", elapsedTime.Milliseconds(), elapsedTime.Seconds(), elapsedTime.Minutes())
}

func listClusterData(cli *rawkv.Client) {
	// file, createErr := os.Create("./clusterData.txt")

	// if createErr != nil {
	// 	log.Panicf("os.Create(): %v\n", createErr)
	// }

	// defer file.Close()

	// Start timer
	getDataStartTime := time.Now()

	keysCount := 0

	// keys, values, getErr := cli.Scan(context.Background(), nil, nil, 60)
	keys, _, getErr := cli.Scan(context.Background(), nil, nil, rawkv.MaxRawKVScanLimit)

	if getErr != nil {
		log.Panicf("cli.Scan(): %v\n", getErr)
	}

	// fmt.Printf("Retrieved %v entries\n", len(keys))
	// printFirstLastKeys(keys)

	retrievedKeys := len(keys)
	keysCount += len(keys)

	for retrievedKeys == rawkv.MaxRawKVScanLimit {
		keys, _, getErr = cli.Scan(context.Background(), keys[len(keys)-1], nil, rawkv.MaxRawKVScanLimit)

		if getErr != nil {
			log.Panicf("cli.Scan(): %v\n", getErr)
		}

		// fmt.Printf("Retrieved %v entries\n", len(keys))
		// printFirstLastKeys(keys)

		retrievedKeys = len(keys)
		keysCount += retrievedKeys - 1
	}

	// Stop timer
	getDataElapsedTime := time.Since(getDataStartTime)

	fmt.Printf("It took %d ms / %.2f sec to get all data stored in TiKV (%v entries)\n", getDataElapsedTime.Milliseconds(), getDataElapsedTime.Seconds(), keysCount)

	// Start timer
	// writeDataStartTime := time.Now()

	// for index, kv := range getResponse.Kvs {
	// 	line := fmt.Sprintf("%v. %v - %v\n", index, string(kv.Key[:]), string(kv.Value[:]))

	// 	_, writeErr := file.WriteString(line)

	// 	if writeErr != nil {
	// 		log.Default().Printf("Error writing to clusterData.txt: %v\n", writeErr)
	// 	}
	// }

	// Stop timer
	// writeDataElapsedTime := time.Since(writeDataStartTime)

	// fmt.Printf("It took %d ms / %.2f sec / %.2f min to write all data stored in TiKV to clusterData.txt\n", writeDataElapsedTime.Milliseconds(), writeDataElapsedTime.Seconds(), writeDataElapsedTime.Minutes())
}

func getClient() (*rawkv.Client, error) {
	log.Default().Println("Connecting to TiKV cluster...")

	errChan := make(chan error, 1)
	cliChan := make(chan *rawkv.Client, 1)

	go func() {
		cli, err := rawkv.NewClient(context.Background(), []string{"127.0.0.1:2379"}, config.DefaultConfig().Security)

		if err != nil {
			errChan <- err
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
		return nil, errors.New("5 seconds timeout exceeded while trying to connect to TiKV cluster")
	}
}

const PROGRAM_INFO = "\nType the number of the option you want to execute:\n\t1. Clear all cluster data at once\n\t2. Clear cluster data one by one\n\t3. Store all data in cluster\n\t4. List all data in cluster\n\t5. Exit the program\n\n"

func main() {
	cli, err := getClient()

	if err != nil {
		log.Fatalf("Could not connect to TiKV cluster: %v\n", err)
	}

	defer cli.Close()

	log.Default().Println("Connected to TiKV cluster successfully")

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
