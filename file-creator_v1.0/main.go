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

	log.Default().Printf("Connected to %s object storage successfully\n", settings.StorageType)

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
	objectName := uuid.New().String()

	metadataCli.StoreKeyValue(newFilePath, objectName)

	storageCli.StoreObject(objectName, newFilePath)

	fmt.Println("File stored successfully")
}

func checkError(e error) {
	if e != nil {
		panic(e)
	}
}
