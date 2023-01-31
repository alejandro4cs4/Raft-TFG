package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	grpcprom "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"
)

type EtcdClient struct {
	client *clientv3.Client
	settings *Settings
}

/*******************/
/*** CONSTRUCTOR ***/
/*******************/

func newEtcdClient(settings *Settings) IClusterClient {
	config := clientv3.Config{
		Endpoints:   []string{"localhost:2379", "localhost:22379", "localhost:32379"},
		DialTimeout: 5 * time.Second,
	}

	if settings.EtcdMetrics {
		config.DialOptions = []grpc.DialOption{
			grpc.WithUnaryInterceptor(grpcprom.UnaryClientInterceptor),
			grpc.WithStreamInterceptor(grpcprom.StreamClientInterceptor),
		}
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
		panic(err)
	case cli := <-cliChan:
		return &EtcdClient{
			client: cli,
			settings: settings,
		}
	case <-time.After(6 * time.Second):
		panic(errors.New("5 seconds timeout exceeded while trying to connect to etcd cluster"))
	}
}

/*************************/
/*** INTERFACE METHODS ***/
/*************************/

func (etcdCli *EtcdClient) closeClient() {
	etcdCli.client.Close()
}

func (etcdCli *EtcdClient) clearAllClusterData(measureElapsedTime bool) {
	var startTime time.Time
	var elapsedTime time.Duration

	// Start timer
	if measureElapsedTime {
		startTime = time.Now()
	}

	deleteResponse, err := etcdCli.client.Delete(context.Background(), "", clientv3.WithFromKey())
	// deleteResponse, err := etcdCli.client.Delete(context.Background(), "", clientv3.WithPrefix())

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

func (etcdCli *EtcdClient) clearClusterDataOneByOne() {
	cmd := exec.Command("find", etcdCli.settings.ExploreDirectory)
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

		_, err := etcdCli.client.Delete(context.Background(), key)

		if err != nil {
			log.Panicf("cli.Delete(%v): %v\n", key, err)
		}
	}

	// Stop timer
	elapsedTime := time.Since(startTime)

	cmd.Wait()

	fmt.Printf("It took %d ms / %.2f sec / %.2f min to delete the data one by one from etcd\n", elapsedTime.Milliseconds(), elapsedTime.Seconds(), elapsedTime.Minutes())
}

func (etcdCli *EtcdClient) storeDataInCluster() {
	cmd := exec.Command("find", etcdCli.settings.ExploreDirectory, "-printf", "%p:%y:%m:%U:%G:%A@:%s:%i\n")
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

		_, err := etcdCli.client.Put(context.Background(), key, value)

		if err != nil {
			log.Panicf("cli.Put(%v, %v): %v\n", key, value, err)
		}

		numStoredItems++

		if etcdCli.settings.MaxItemsStore > 0 && numStoredItems == etcdCli.settings.MaxItemsStore {
			break
		}
	}

	// Stop timer
	elapsedTime := time.Since(startTime)

	fmt.Printf("It took %d ms / %.2f sec / %.2f min to store the data in etcd\n", elapsedTime.Milliseconds(), elapsedTime.Seconds(), elapsedTime.Minutes())
}

func (etcdCli *EtcdClient) listClusterData() {
	// Start timer
	getDataStartTime := time.Now()

	getResponse, getErr := etcdCli.client.Get(context.Background(), "", clientv3.WithPrefix())

	if getErr != nil {
		log.Panicf("cli.Get(): %v\n", getErr)
	}

	// Stop timer
	getDataElapsedTime := time.Since(getDataStartTime)

	for index, kv := range getResponse.Kvs {
		line := fmt.Sprintf("%v. %v - %v\n", index, string(kv.Key[:]), string(kv.Value[:]))

		fmt.Print(line)
	}

	fmt.Printf("It took %d ms / %.2f sec to get all data stored in etcd (%v entries)\n", getDataElapsedTime.Milliseconds(), getDataElapsedTime.Seconds(), getResponse.Count)
}

func (etcdCli *EtcdClient) getMetrics() {
	// Listen for all Prometheus metrics
	ln, err := net.Listen("tcp", ":0")
	if err != nil {
		log.Panicf("net.Listen(): %v\n", err)
	}

	donec := make(chan struct{})

	go func() {
		defer close(donec)
		http.Serve(ln, promhttp.Handler())
	}()

	defer func() {
		ln.Close()
		<-donec
	}()

	// Make an HTTP request to fetch all Prometheus metrics
	url := "http://" + ln.Addr().String() + "/metrics"
	fmt.Printf("URL: %s\n", url)

	resp, err := http.Get(url)
	if err != nil {
		log.Panicf("http.Get(): %v\n", err)
	}

	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		log.Panicf("io.ReadAll(): %v\n", err)
	}

	err = os.WriteFile("./metrics.log", body, 0644)
	if err != nil {
		log.Panicf("os.WriteFile(): %v\n", err)
	}

	fmt.Printf("Generated metrics successfully in metrics.log\n")
}