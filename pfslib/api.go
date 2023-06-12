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

type PfsFile struct {
	*globals.Openfd
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

	log.Default().Printf("[pfslib]: PfsOpen(): file \"%s\" opened successfully, file descriptor is \"%d\"\n", pathname, newFd)

	// Return openfds index for the recently opened file
	return newFd, nil
}

func (pfsf *PfsFile) Close() error {
	return nil
}

// func PfsClose(fileDescriptor int) error {
// 	if fileDescriptor < 0 || fileDescriptor > globals.OpenfdsMaxSize {
// 		printInvalidDescriptor(fileDescriptor, "Close")
// 		return errors.New("Invalid descriptor")
// 	}

// 	globals.Openfds[fileDescriptor] = nil

// 	log.Default().Printf("[pfslib]: PfsClose(): file associated to descriptor \"%d\" closed successfully\n", fileDescriptor)

// 	return nil
// }

// Tries to read len(buffer) bytes from the file pointed by fileDescriptor
func PfsRead(fileDescriptor int, buffer []byte) (int, error) {
	if fileDescriptor < 0 || fileDescriptor > globals.OpenfdsMaxSize {
		printInvalidDescriptor(fileDescriptor, "Read")
		return 0, errors.New("Invalid descriptor")
	}

	if buffer == nil {
		printInvalidBuffer("Read")
		return -1, errors.New("Invalid buffer")
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

	globals.Openfds[fileDescriptor].Offset += int64(bytesRead)

	log.Default().Printf("[pfslib]: PfsRead(): %d bytes successfully read from file associated to descriptor \"%d\"\n", bytesRead, fileDescriptor)

	return bytesRead, nil
}

// Tries to write len(buffer) bytes to the file pointed by fileDescriptor
func PfsWrite(fileDescriptor int, buffer []byte) (int, error) {
	if fileDescriptor < 0 || fileDescriptor > globals.OpenfdsMaxSize {
		printInvalidDescriptor(fileDescriptor, "Write")
		return -1, errors.New("Invalid descriptor")
	}

	if buffer == nil {
		printInvalidBuffer("Write")
		return -1, errors.New("Invalid buffer")
	}

	fileOffset := globals.Openfds[fileDescriptor].Offset

	minioObject, err := storeClient.GetObject(context.Background(), globals.MinioBucket, globals.Openfds[fileDescriptor].Uuid, minio.GetObjectOptions{})
	utils.CheckError(err)

	defer minioObject.Close()

	minioObjectInfo, err := minioObject.Stat()
	utils.CheckError(err)

	preContentBuffer := make([]byte, fileOffset)
	_, err = minioObject.Read(preContentBuffer)
	utils.CheckError(err)

	postContentBuffer := make([]byte, minioObjectInfo.Size-fileOffset)
	_, err = minioObject.ReadAt(postContentBuffer, fileOffset)
	utils.CheckError(err)

	newFileContent := append(preContentBuffer[:], buffer[:]...)
	newFileContent = append(newFileContent[:], postContentBuffer[:]...)

	fmt.Printf("The new contents are:\n%s\n", string(newFileContent))

	return len(buffer), nil
}

// Changes the offset of the file for the next read/write and returns the new offset
func (pfsf *PfsFile) Lseek(offset int64, whence int) (int64, error) {
	if offset < 0 {
		printInvalidOffset(offset)
		return -1, errors.New("Invalid offset")
	}

	if whence < 0 || whence > 2 {
		printInvalidWhence(whence)
		return -1, errors.New("Invalid whence")
	}

	switch whence {
	case 0:
		pfsf.Offset = offset
		break
	case 1:
		pfsf.Offset += offset
		break
	case 2:
		pfsf.Offset = pfsf.getSize() + offset
		break
	}

	log.Default().Printf("[pfslib]: PfsLseek(): offset is now set to \"%d\"\n", pfsf.Offset)

	return pfsf.Offset, nil
}

func (pfsf *PfsFile) getSize() int64 {
	minioObjectInfo, err := storeClient.StatObject(context.Background(), globals.MinioBucket, pfsf.Uuid, minio.StatObjectOptions{})
	utils.CheckError(err)

	return minioObjectInfo.Size
}

// func PfsLseek(fileDescriptor int, offset int64, whence int) (int64, error) {
// 	if fileDescriptor < 0 || fileDescriptor > globals.OpenfdsMaxSize {
// 		printInvalidDescriptor(fileDescriptor, "Lseek")
// 		return -1, errors.New("Invalid descriptor")
// 	}

// 	if offset < 0 {
// 		printInvalidOffset(offset)
// 		return -1, errors.New("Invalid offset")
// 	}

// 	if whence < 0 || whence > 2 {
// 		printInvalidWhence(whence)
// 		return -1, errors.New("Invalid whence")
// 	}

// 	openfd := globals.Openfds[fileDescriptor]

// 	switch whence {
// 	case 0:
// 		openfd.Offset = offset
// 		break
// 	case 1:
// 		openfd.Offset += offset
// 		break
// 	case 2:
// 		globals.Openfds[fileDescriptor].Offset = getFileSize(fileDescriptor) + offset
// 		break
// 	}

// 	log.Default().Printf("[pfslib]: PfsLseek(): offset of file associated to descriptor \"%d\" is now set to \"%d\"\n", fileDescriptor, openfd.Offset)

// 	return openfd.Offset, nil
// }
