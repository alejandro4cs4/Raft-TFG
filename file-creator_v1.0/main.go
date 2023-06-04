package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	uuid "github.com/google/uuid"
	"gopkg.in/yaml.v3"

	metadataclient "Raft-TFG/file-creator_v1.0/metadata-client"
	"Raft-TFG/file-creator_v1.0/settings"
	storageclient "Raft-TFG/file-creator_v1.0/storage-client"
)

const (
	enterPathMessage string = "Enter absolute path for the new file:\n> "
	retryPathMessage        = "\nThe file already exists, please try with another path:\n> "
	rootDirectoryKey        = "nil_nil"
)

func main() {
	settings := getSettings()

	// Connect to metadata cluster
	metadataClient, err := metadataclient.GetMetadataClient(&settings)
	checkError(err)

	defer metadataClient.CloseClient()

	log.Default().Printf("Connected to %s metadata storage successfully\n", settings.MetadataType)

	// Connect to object storage
	storageClient, err := storageclient.GetStorageClient(&settings)
	checkError(err)

	log.Default().Printf("Connected to %s object storage successfully\n\n", settings.StorageType)

	// Get new file path
	newFilePath := handleNewFilePath()

	// Create file
	createFile(newFilePath)

	// Store file
	storeFile(metadataClient, storageClient, newFilePath)
}

func getSettings() (settings settings.Settings) {
	settingsFileContent, err := os.ReadFile("./settings.yaml")
	checkError(err)

	yaml.Unmarshal(settingsFileContent, &settings)

	return
}

func handleNewFilePath() (newFilePath string) {
	newFilePath = getNewFilePath(enterPathMessage)

	for {
		doesFileExist := fileExists(newFilePath)

		if doesFileExist {
			newFilePath = getNewFilePath(retryPathMessage)
			continue
		}

		break
	}

	return
}

func getNewFilePath(consoleMessage string) (spaceTrimmedNewFilePath string) {
	fmt.Printf(consoleMessage)

	reader := bufio.NewReader(os.Stdin)

	newFilePath, err := reader.ReadString('\n')

	if err != nil {
		log.Fatalln("An error occurred while reading input")
	}

	newfilePathWithoutNewline := strings.TrimSuffix(newFilePath, "\n")
	spaceTrimmedNewFilePath = strings.Trim(newfilePathWithoutNewline, " ")

	return
}

func fileExists(filePath string) bool {
	_, err := os.Open(filePath)

	if err != nil {
		return os.IsExist(err)
	}

	return true
}

func createFile(newFilePath string) {
	filePtr, err := os.Create(newFilePath)
	checkError(err)

	defer filePtr.Close()
}

func storeFile(metadataCli metadataclient.IMetadataClient, storageCli storageclient.IStorageClient, newFilePath string) {
	routeComponents := strings.Split(newFilePath, "/")

	lastParentDirectoryUUID := storeDirectoryHierarchy(metadataCli, routeComponents)

	// Store file itself
	objectName := uuid.New().String()
	metadataFileKey := strings.Join([]string{lastParentDirectoryUUID, routeComponents[len(routeComponents)-1]}, "_")
	metadataFileValue := strings.Join([]string{getFileType(len(routeComponents)-1, len(routeComponents)), objectName}, "_")

	metadataCli.StoreKeyValue(metadataFileKey, metadataFileValue)

	storageCli.StoreObject(objectName, newFilePath)

	fmt.Println("File stored successfully")
}

func storeDirectoryHierarchy(metadataCli metadataclient.IMetadataClient, routeComponents []string) string {
	var parentDirectoryUUID string

	for index, routeComponentName := range routeComponents[0 : len(routeComponents)-1] {
		metadataKey := getMetadataKey(routeComponentName, parentDirectoryUUID)

		getResponse := metadataCli.GetByKey(metadataKey)

		if getResponse.Count == 0 {
			objectName := uuid.New().String()

			metadataCli.StoreKeyValue(metadataKey, strings.Join([]string{getFileType(index, len(routeComponents)), objectName}, "_"))

			parentDirectoryUUID = objectName

			continue
		}

		parentDirectoryUUID = strings.Split(getResponse.Value, "_")[1]
	}

	return parentDirectoryUUID
}

func getMetadataKey(rawName string, parentUUID ...string) string {
	if rawName == "" {
		return rootDirectoryKey
	}

	return strings.Join([]string{parentUUID[0], rawName}, "_")
}

func getFileType(routeComponentIndex, routeComponentCount int) string {
	if routeComponentIndex < routeComponentCount {
		return "D"
	}

	return "F"
}

func checkError(e error) {
	if e != nil {
		panic(e)
	}
}
