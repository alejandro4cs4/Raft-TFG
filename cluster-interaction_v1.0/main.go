package main

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

func main() {
	var settings Settings

	settingsFileContent, err := os.ReadFile("./settings.yaml")
	checkError(err)

	yaml.Unmarshal(settingsFileContent, &settings)

	fmt.Printf("%+v\n", settings)

	clusterClient, err := getClusterClient(settings.ClusterType)
	checkError(err)

	fmt.Printf("%+v\n", clusterClient)
}

func checkError(e error) {
	if e != nil {
		panic(e)
	}
}