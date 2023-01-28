package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"time"

	grpcprom "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"
)

func main() {
	cli, err := getClient()

	if err != nil {
		log.Fatalf("Could not connect to etcd cluster: %v\n", err)
	}

	defer cli.Close()

	log.Default().Println("Connected to etcd cluster successfully")

	readGetMetrics(cli)
}

func getClient() (*clientv3.Client, error) {
	log.Default().Println("Connecting to etcd cluster...")

	config := clientv3.Config{
		Endpoints:   []string{"localhost:2379", "localhost:22379", "localhost:32379"},
		DialTimeout: 5 * time.Second,
		DialOptions: []grpc.DialOption{
			grpc.WithUnaryInterceptor(grpcprom.UnaryClientInterceptor),
			grpc.WithStreamInterceptor(grpcprom.StreamClientInterceptor),
		},
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

func readGetMetrics(cli *clientv3.Client) {
	cli.Get(context.Background(), "test_key");

	// Listen for all Prometheus metrics
	ln, err := net.Listen("tcp", ":0")
	checkErr(err)

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

	resp, err := http.Get(url)
	checkErr(err)

	b, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	checkErr(err)

		fmt.Printf("Result:\n%v\n", string(b))
}

func checkErr(err error) {
		if err != nil {
			log.Fatal(err)
		}
}