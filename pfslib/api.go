package pfslib

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
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

type openfd struct {
	name   string
	uuid   string
	offset int64
}

type PfsFile struct {
	*openfd
}

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

// Tries to open the pathname file and return a PfsFile that represents it
func PfsOpen(pathname string) (*PfsFile, error) {
	// Solve file pathname contacting etcd
	// - start from root (/) if absolute pathname
	// - start from current directory if relative pathname
	absolutePath := utils.GetAbsolutePath(pathname)
	pathComponents := strings.Split(absolutePath, "/")
	parentDirectoryUuid := RootDirectoryUuid
	var currentComponentValue string

	for index, pathComponent := range pathComponents {
		mappedName := utils.MapRouteComponentName(pathComponent)
		queryKey := strings.Join([]string{parentDirectoryUuid, mappedName}, "_")

		getResponse, err := metaClient.Get(context.Background(), queryKey)
		utils.CheckError(err)

		if getResponse.Count == 0 {
			printOpenNotFound(pathComponents[:index+1])

			return nil, errors.New("No such file or directory")
		}

		currentComponentValue = string(getResponse.Kvs[0].Value)
		parentDirectoryUuid = strings.Split(currentComponentValue, "_")[1]
	}

	// If file is not regular file (F) -> return error
	lastComponentType := strings.Split(currentComponentValue, "_")[0]

	if lastComponentType != TypeRegular {
		printOpenNotRegular(absolutePath)

		return nil, errors.New("Not a regular file")
	}

	// Get file's UUID from its etcd value
	requestedFileUuid := strings.Split(currentComponentValue, "_")[1]

	newPfsFile := &PfsFile{
		&openfd{
			name:   pathComponents[len(pathComponents)-1],
			uuid:   requestedFileUuid,
			offset: 0,
		},
	}

	log.Default().Printf("[pfslib]: PfsOpen(): file \"%s\" opened successfully\n", pathname)

	// Return openfds index for the recently opened file
	return newPfsFile, nil
}

// Closes the PfsFile making it invalid
func (pfsf *PfsFile) Close() error {
	pfsf.openfd = nil

	return nil
}

// Tries to read len(buffer) bytes from the PfsFile
func (pfsf *PfsFile) Read(buffer []byte) (int, error) {
	if pfsf.openfd == nil {
		printClosedPfsFile("Read")
		return 0, errors.New("PfsFile already closed")
	}

	if buffer == nil {
		printInvalidBuffer("Read")
		return 0, errors.New("Invalid buffer")
	}

	minioObject, err := storeClient.GetObject(context.Background(), globals.MinioBucket, pfsf.uuid, minio.GetObjectOptions{})
	utils.CheckError(err)

	defer minioObject.Close()

	bytesRead, err := minioObject.ReadAt(buffer, pfsf.offset)
	utils.CheckError(err)

	pfsf.offset += int64(bytesRead)

	log.Default().Printf("[pfslib]: Read(): %d bytes successfully read from the PfsFile\n", bytesRead)

	return bytesRead, nil
}

// Tries to write len(buffer) bytes to the PfsFile
func (pfsf *PfsFile) Write(buffer []byte) (int, error) {
	if pfsf.openfd == nil {
		printClosedPfsFile("Write")
		return 0, errors.New("PfsFile already closed")
	}

	if buffer == nil {
		printInvalidBuffer("Write")
		return 0, errors.New("Invalid buffer")
	}

	minioObject, err := storeClient.GetObject(context.Background(), globals.MinioBucket, pfsf.uuid, minio.GetObjectOptions{})
	utils.CheckError(err)

	defer minioObject.Close()

	minioObjectInfo, err := minioObject.Stat()
	utils.CheckError(err)

	var preContentBufferSize int64
	if pfsf.offset == minioObjectInfo.Size {
		preContentBufferSize = minioObjectInfo.Size - 1
	} else {
		preContentBufferSize = pfsf.offset
	}

	preContentBuffer := make([]byte, preContentBufferSize)
	if preContentBufferSize > 0 {
		_, err = minioObject.Read(preContentBuffer)
		utils.CheckError(err)
	}

	var postContentBufferSize int64
	if pfsf.offset == minioObjectInfo.Size {
		postContentBufferSize = 0
	} else {
		postContentBufferSize = minioObjectInfo.Size - (pfsf.offset + int64(len(buffer))) - 1
	}

	postContentBuffer := make([]byte, postContentBufferSize)
	if postContentBufferSize > 0 {
		_, err = minioObject.ReadAt(postContentBuffer, pfsf.offset+int64(len(buffer)))
		utils.CheckError(err)
	}

	newFileContent := append(preContentBuffer[:], buffer[:]...)
	newFileContent = append(newFileContent[:], postContentBuffer[:]...)
	reader := bytes.NewReader(newFileContent)

	_, err = storeClient.PutObject(context.Background(), globals.MinioBucket, pfsf.uuid, reader, int64(len(newFileContent)), minio.PutObjectOptions{})
	utils.CheckError(err)

	pfsf.offset += int64(len(buffer))

	return len(buffer), nil
}

// Changes the offset of the PfsFile for the next read/write and returns the new offset
func (pfsf *PfsFile) Lseek(offset int64, whence int) (int64, error) {
	if pfsf.openfd == nil {
		printClosedPfsFile("Lseek")
		return 0, errors.New("PfsFile already closed")
	}

	if offset < 0 {
		printInvalidOffset(offset)
		return pfsf.offset, errors.New("Invalid offset")
	}

	if whence < 0 || whence > 2 {
		printInvalidWhence(whence)
		return pfsf.offset, errors.New("Invalid whence")
	}

	switch whence {
	case io.SeekStart:
		pfsf.offset = offset
		break
	case io.SeekCurrent:
		pfsf.offset += offset
		break
	case io.SeekEnd:
		pfsf.offset = pfsf.getSize() + offset
		break
	}

	log.Default().Printf("[pfslib]: PfsLseek(): offset is now set to \"%d\"\n", pfsf.offset)

	return pfsf.offset, nil
}

func (pfsf *PfsFile) getSize() int64 {
	minioObjectInfo, err := storeClient.StatObject(context.Background(), globals.MinioBucket, pfsf.uuid, minio.StatObjectOptions{})
	utils.CheckError(err)

	return minioObjectInfo.Size
}

// Tries to create the directory specified by pathname
func PfsMkdir(pathname string, perms os.FileMode) error {
	if pathname == "" {
		printInvalidEmptyPathname()
		return errors.New("Empty pathname not allowed")
	}

	absolutePath := utils.GetAbsolutePath(pathname)
	pathComponents := strings.Split(absolutePath, "/")

	if len(pathComponents) == 0 {
		printInvalidPathname()
		return errors.New("Invalid pathname")
	}

	parentDirectoryUuid := RootDirectoryUuid

	for index, pathComponent := range pathComponents[:len(pathComponents)-1] {
		mappedName := utils.MapRouteComponentName(pathComponent)
		queryKey := strings.Join([]string{parentDirectoryUuid, mappedName}, "_")

		getResponse, err := metaClient.Get(context.Background(), queryKey)
		utils.CheckError(err)

		// The directory does not exist
		if getResponse.Count == 0 {
			nonCreatedDirectory := strings.Join(pathComponents[:index+1], "/")
			errorMsg := fmt.Sprintf("The directory \"%s\" does not exist\n", nonCreatedDirectory)

			printDirectoryNotCreated(errorMsg)
			return errors.New(errorMsg)
		}

		currentDirectoryUuid := strings.Split(string(getResponse.Kvs[0].Value), "_")[1]
		parentDirectoryUuid = currentDirectoryUuid
	}

	newDirectoryName := utils.MapRouteComponentName(pathComponents[len(pathComponents)-1])
	newDirectoryKey := strings.Join([]string{parentDirectoryUuid, newDirectoryName}, "_")

	// Check if directory already exists
	getResponse, err := metaClient.Get(context.Background(), newDirectoryKey)
	utils.CheckError(err)

	// Directory already exists
	if getResponse.Count == 1 {
		return nil
	}

	newDirectoryUuid := uuid.New().String()
	newDirectoryValue := strings.Join([]string{TypeDirectory, newDirectoryUuid}, "_")

	_, err = metaClient.Put(context.Background(), newDirectoryKey, newDirectoryValue)
	utils.CheckError(err)

	return nil
}

// Tries to create the directory specified by pathname and all its parent directories
func PfsMkdirAll(pathname string, perms os.FileMode) error {
	if pathname == "" {
		printInvalidEmptyPathname()
		return errors.New("Empty pathname not allowed")
	}

	absolutePath := utils.GetAbsolutePath(pathname)
	pathComponents := strings.Split(absolutePath, "/")

	if len(pathComponents) == 0 {
		printInvalidPathname()
		return errors.New("Invalid pathname")
	}

	parentDirectoryUuid := RootDirectoryUuid

	for index, pathComponent := range pathComponents {
		mappedName := utils.MapRouteComponentName(pathComponent)
		queryKey := strings.Join([]string{parentDirectoryUuid, mappedName}, "_")
		var currentDirectoryUuid string

		getResponse, err := metaClient.Get(context.Background(), queryKey)
		utils.CheckError(err)

		// The directory does not exist
		if getResponse.Count == 0 {
			newDirectoryName := utils.MapRouteComponentName(pathComponents[index])
			newDirectoryKey := strings.Join([]string{parentDirectoryUuid, newDirectoryName}, "_")
			newDirectoryUuid := uuid.New().String()
			newDirectoryValue := strings.Join([]string{TypeDirectory, newDirectoryUuid}, "_")

			_, err := metaClient.Put(context.Background(), newDirectoryKey, newDirectoryValue)
			utils.CheckError(err)

			currentDirectoryUuid = newDirectoryUuid
		} else {
			currentDirectoryUuid = strings.Split(string(getResponse.Kvs[0].Value), "_")[1]
		}

		parentDirectoryUuid = currentDirectoryUuid
	}

	return nil
}

// Creates or truncates the file pointed by pathname
func PfsCreate(pathname string) (*PfsFile, error) {
	if pathname == "" {
		printInvalidEmptyPathname()
		return nil, errors.New("Empty pathname not allowed")
	}

	absolutePath := utils.GetAbsolutePath(pathname)
	pathComponents := strings.Split(absolutePath, "/")

	if len(pathComponents) == 0 {
		printInvalidPathname()
		return nil, errors.New("Invalid pathname")
	}

	parentDirectoryUuid := RootDirectoryUuid

	for index, pathComponent := range pathComponents[:len(pathComponents)-1] {
		mappedName := utils.MapRouteComponentName(pathComponent)
		queryKey := strings.Join([]string{parentDirectoryUuid, mappedName}, "_")

		getResponse, err := metaClient.Get(context.Background(), queryKey)
		utils.CheckError(err)

		// The directory does not exist
		if getResponse.Count == 0 {
			nonCreatedDirectory := strings.Join(pathComponents[:index+1], "/")
			errorMsg := fmt.Sprintf("The directory \"%s\" does not exist\n", nonCreatedDirectory)

			printDirectoryNotCreated(errorMsg)
			return nil, errors.New(errorMsg)
		}

		currentDirectoryUuid := strings.Split(string(getResponse.Kvs[0].Value), "_")[1]
		parentDirectoryUuid = currentDirectoryUuid
	}

	newFileName := utils.MapRouteComponentName(pathComponents[len(pathComponents)-1])
	newFileKey := strings.Join([]string{parentDirectoryUuid, newFileName}, "_")

	// Check if the file already exists
	getResponse, err := metaClient.Get(context.Background(), newFileKey)
	utils.CheckError(err)

	// Directory already exists
	if getResponse.Count == 1 {
		fileValue := string(getResponse.Kvs[0].Value)
		fileUuid := strings.Split(fileValue, "_")[1]

		return &PfsFile{
			&openfd{
				name:   newFileName,
				uuid:   fileUuid,
				offset: 0,
			},
		}, nil
	}

	newFileUuid := uuid.New().String()
	newFileValue := strings.Join([]string{TypeRegular, newFileUuid}, "_")

	_, err = metaClient.Put(context.Background(), newFileKey, newFileValue)
	utils.CheckError(err)

	return &PfsFile{
		&openfd{
			name:   newFileName,
			uuid:   newFileUuid,
			offset: 0,
		},
	}, nil

}
