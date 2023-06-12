package pfslib

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	clientv3 "go.etcd.io/etcd/client/v3"

	"raft-tfg.com/alejandroc/pfslib/globals"
	"raft-tfg.com/alejandroc/pfslib/metaclient"
	"raft-tfg.com/alejandroc/pfslib/storeclient"
	"raft-tfg.com/alejandroc/pfslib/utils"
)

const (
	RootDirectoryUuid string = "nil"
	RootDirectoryKey         = "nil_nil"
)

const (
	TypeDirectory string = "D"
	TypeRegular          = "F"
)

var metaClient *clientv3.Client
var storeClient *minio.Client

func PfsInit() {
	utils.ReadSettings()

	// Init etcd client
	log.Default().Printf("[pfslib]: Connecting to etcd metadata storage...")
	metaClient = metaclient.GetEtcdClient()
	log.Default().Printf("[pfslib]: Connected to etcd metadata storage successfully")

	// Init MinIO client
	log.Default().Printf("[pfslib]: Connecting to MinIO object storage...")
	storeClient = storeclient.GetMinioClient()
	log.Default().Printf("[pfslib]: Connected to MinIO object storage successfully")

	// Init root directory
	initRoot()
}

func initRoot() {
	getResponse, err := metaClient.Get(context.Background(), RootDirectoryKey)

	utils.CheckError(err)

	if getResponse.Count == 1 {
		return
	}

	rootUuid := uuid.New().String()
	rootValue := strings.Join([]string{TypeDirectory, rootUuid}, "_")

	_, err = metaClient.Put(context.Background(), RootDirectoryKey, rootValue)

	utils.CheckError(err)
}

// Tries to open the pathname file and return its associated file descriptor
func PfsOpen(pathname string) (int, error) {
	// Solve file pathname contacting etcd
	// - start from root (/) if absolute pathname
	// - start from current directory if relative pathname
	absolutePath := utils.GetAbsolutePath(pathname)
	pathComponents := strings.Split(absolutePath, "/")
	parentDirectoryUuid := RootDirectoryUuid
	var currentComponentValue string
	newFd := -1

	for index, pathComponent := range pathComponents {
		mappedName := utils.MapRouteComponentName(pathComponent)
		queryKey := strings.Join([]string{parentDirectoryUuid, mappedName}, "_")

		getResponse, err := metaClient.Get(context.Background(), queryKey)

		utils.CheckError(err)

		if getResponse.Count == 0 {
			printOpenNotFound(pathComponents[:index+1])

			return -1, errors.New("No such file or directory")
		}

		currentComponentValue = string(getResponse.Kvs[0].Value)
		parentDirectoryUuid = strings.Split(currentComponentValue, "_")[1]
	}

	// If file is not regular file (F) -> return error
	lastComponentType := strings.Split(currentComponentValue, "_")[0]

	if lastComponentType != TypeRegular {
		printOpenNotRegular(absolutePath)

		return -1, errors.New("Not a regular file")
	}

	// Get file's UUID from its etcd value
	requestedFileUuid := strings.Split(currentComponentValue, "_")[1]

	// Look for a free entry in openfds table and store the file UUID
	for index, openfd := range globals.Openfds {
		if openfd == nil {
			newFd = index
			break
		}
	}

	if newFd == -1 {
		printOpenOutOfMemory()
		return -1, errors.New("Out of memory")
	}

	globals.Openfds[newFd] = &globals.Openfd{
		Name:   pathComponents[len(pathComponents)-1],
		Uuid:   requestedFileUuid,
		Offset: 0,
	}

	fmt.Printf("File opened successfully\nOpenfds: %+v\n", globals.Openfds)

	// Return openfds index for the recently opened file
	return newFd, nil
}

// Tries to read
func PfsRead(fileDescriptor int, buffer []byte) (int, error) {
	if fileDescriptor < 0 || fileDescriptor > globals.OpenfdsMaxSize {
		printReadInvalidDescriptor(fileDescriptor)
		return 0, errors.New("Invalid descriptor")
	}

	openDescriptor := globals.Openfds[fileDescriptor]

	if openDescriptor == nil {
		printReadNotOpenedDescriptor(fileDescriptor)
		return 0, errors.New("No file associated to the specified descriptor")
	}

	objectToReadUuid := openDescriptor.Uuid

	minioObject, err := storeClient.GetObject(context.Background(), globals.MinioBucket, objectToReadUuid, minio.GetObjectOptions{})

	utils.CheckError(err)

	defer minioObject.Close()

	bytesRead, err := minioObject.ReadAt(buffer, openDescriptor.Offset)

	utils.CheckError(err)

	return bytesRead, nil
}
