package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"log"
	"os/exec"
	"strings"
	"time"

	"github.com/tikv/client-go/v2/config"
	"github.com/tikv/client-go/v2/rawkv"
)

type TikvClient struct {
	client *rawkv.Client
	settings *Settings
}

/*******************/
/*** CONSTRUCTOR ***/
/*******************/

func newTikvClient(settings *Settings) IClusterClient {
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
		panic(err)
	case cli := <-cliChan:
		return &TikvClient{
			client: cli,
			settings: settings,
		}
	case <-time.After(6 * time.Second):
		panic(errors.New("5 seconds timeout exceeded while trying to connect to TiKV cluster"))
	}
}

/*************************/
/*** INTERFACE METHODS ***/
/*************************/

func (tikvCli *TikvClient) closeClient() {
	tikvCli.client.Close()
}

func (tikvCli *TikvClient) clearAllClusterData(measureElapsedTime bool) {
	var startTime time.Time
	var elapsedTime time.Duration

	// Start timer
	if measureElapsedTime {
		startTime = time.Now()
	}

	err := tikvCli.client.DeleteRange(context.Background(), []byte("/"), nil)

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
}

func (tikvCli *TikvClient) clearClusterDataOneByOne() {
	cmd := exec.Command("find", tikvCli.settings.ExploreDirectory)
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

		err := tikvCli.client.Delete(context.Background(), []byte(key))

		if err != nil {
			log.Panicf("cli.Delete(%v): %v\n", key, err)
		}
	}

	// Stop timer
	elapsedTime := time.Since(startTime)

	cmd.Wait()

	fmt.Printf("It took %d ms / %.2f sec / %.2f min to delete the data one by one from TiKV\n", elapsedTime.Milliseconds(), elapsedTime.Seconds(), elapsedTime.Minutes())
}

func (tikvCli *TikvClient) storeDataInCluster() {
	cmd := exec.Command("find", tikvCli.settings.ExploreDirectory, "-printf", "%p:%y:%m:%U:%G:%A@:%s:%i\n")
	stdout, err := cmd.StdoutPipe()

	if err != nil {
		log.Panicf("cmd.StdoutPipe(): %v\n", err)
	}

	cmd.Start()

	scanner := bufio.NewScanner(stdout)

	// Start timer
	startTime := time.Now()

	var numStoredItems int64

	for scanner.Scan() {
		key, value, _ := strings.Cut(scanner.Text(), ":")

		err := tikvCli.client.Put(context.Background(), []byte(key), []byte(value))

		if err != nil {
			log.Panicf("cli.Put(%v, %v): %v\n", key, value, err)
		}

		numStoredItems++

		if tikvCli.settings.MaxItemsStore > 0 && numStoredItems == tikvCli.settings.MaxItemsStore {
			break
		}
	}

	// Stop timer
	elapsedTime := time.Since(startTime)

	cmd.Wait()

	fmt.Printf("It took %d ms / %.2f sec / %.2f min to store the data in TiKV\n", elapsedTime.Milliseconds(), elapsedTime.Seconds(), elapsedTime.Minutes())
}

func (tikvCli *TikvClient) listClusterData() {
	// Start timer
	getDataStartTime := time.Now()

	keysCount := 0

	keys, _, getErr := tikvCli.client.Scan(context.Background(), nil, nil, rawkv.MaxRawKVScanLimit)

	if getErr != nil {
		log.Panicf("cli.Scan(): %v\n", getErr)
	}

	retrievedKeys := len(keys)
	keysCount += len(keys)

	for retrievedKeys == rawkv.MaxRawKVScanLimit {
		keys, _, getErr = tikvCli.client.Scan(context.Background(), keys[len(keys)-1], nil, rawkv.MaxRawKVScanLimit)

		if getErr != nil {
			log.Panicf("cli.Scan(): %v\n", getErr)
		}

		retrievedKeys = len(keys)
		keysCount += retrievedKeys - 1
	}

	// Stop timer
	getDataElapsedTime := time.Since(getDataStartTime)

	fmt.Printf("It took %d ms / %.2f sec to get all data stored in TiKV (%v entries)\n", getDataElapsedTime.Milliseconds(), getDataElapsedTime.Seconds(), keysCount)
}