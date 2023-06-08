package metaclient

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	etcdClient "go.etcd.io/etcd/client/v3"

	"raft-tfg.com/alejandroc/pfslib/globals"
)

func GetEtcdClient() *etcdClient.Client {
	etcdEndpoints := []string{}

	for _, etcdEndpoint := range globals.PfsSettings.EtcdSettings.Endpoints {
		etcdIp := etcdEndpoint.Ip
		etcdPort := strconv.Itoa(etcdEndpoint.Port)
		etcdEndpoint := strings.Join([]string{etcdIp, etcdPort}, ":")

		etcdEndpoints = append(etcdEndpoints, etcdEndpoint)
	}

	config := etcdClient.Config{
		Endpoints:   etcdEndpoints,
		DialTimeout: 5 * time.Second,
	}

	errChan := make(chan error, 1)
	cliChan := make(chan *etcdClient.Client, 1)

	go func() {
		cli, err := etcdClient.New(config)

		if err != nil {
			errChan <- err
			return
		}

		statusRes, err := cli.Status(context.Background(), config.Endpoints[0])

		if err != nil {
			errChan <- err
			return
		} else if statusRes == nil {
			errChan <- errors.New("[pfslib]: The status response form etcd was nil")
			return
		}

		cliChan <- cli
	}()

	select {
	case err := <-errChan:
		panicMsg := fmt.Sprintf("[pfslib]: %v\n", err)
		panic(errors.New(panicMsg))
	case cli := <-cliChan:
		return cli
	case <-time.After(5 * time.Second):
		panic(errors.New("[pfslib]: 5 seconds timeout exceeded while trying to connect to etcd metadata storage"))
	}
}
