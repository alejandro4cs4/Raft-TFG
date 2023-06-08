package storeclient

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"raft-tfg.com/alejandroc/pfslib/globals"
)

func GetMinioClient() *minio.Client {
	minioIp := globals.PfsSettings.MinioSettings.Endpoint.Ip
	minioPort := strconv.Itoa(globals.PfsSettings.MinioSettings.Endpoint.Port)
	endpoint := strings.Join([]string{minioIp, minioPort}, ":")
	accessKeyID := globals.PfsSettings.MinioSettings.Credentials.AccessKeyId
	secretAccessKey := globals.PfsSettings.MinioSettings.Credentials.SecretAccessKey
	useSSL := false

	config := minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
	}

	errChan := make(chan error, 1)
	cliChan := make(chan *minio.Client, 1)
	ticker := time.NewTicker(time.Second)
	quit := make(chan struct{})

	go func() {
		cli, err := minio.New(endpoint, &config)

		if err != nil {
			errChan <- err
			return
		}

		for {
			select {
			case <-ticker.C:
				_, err := cli.GetBucketLocation(context.Background(), "testbucket")
				if err == nil {
					cliChan <- cli
					close(quit)
				}
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()

	select {
	case err := <-errChan:
		panicMsg := fmt.Sprintf("[pfslib]: %v\n", err)
		panic(errors.New(panicMsg))
	case cli := <-cliChan:
		return cli
	case <-time.After(6 * time.Second):
		panic(errors.New("[pfslib]: 5 seconds timeout exceeded while trying to connect to MinIO object storage"))
	}
}
