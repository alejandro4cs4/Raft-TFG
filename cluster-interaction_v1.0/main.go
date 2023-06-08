package main

import (
	"bufio"
	"fmt"
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

const PROGRAM_INFO = "\nType the number of the option you want to execute:\n\t1. Clear all cluster data at once\n\t2. Clear cluster data one by one\n\t3. Store all data in cluster\n\t4. List all data in cluster\n\t5. Generate metrics\n\t6. Exit the program\n\n"

func main() {
	var settings Settings

	settingsFileContent, err := os.ReadFile("./settings.yaml")
	checkError(err)

	yaml.Unmarshal(settingsFileContent, &settings)

	clusterClient, err := getClusterClient(&settings)
	checkError(err)

	defer clusterClient.closeClient()

	log.Default().Printf("Connected to %s cluster successfully\n", settings.ClusterType)

	handleCommandInteraction(clusterClient, &settings)
}

func handleCommandInteraction(clusterClient IClusterClient, settings *Settings) {
	// Read user input
	buf := bufio.NewReader(os.Stdin)

	fmt.Print(PROGRAM_INFO)

	for {
		fmt.Print("> ")

		input, err := buf.ReadBytes('\n')

		if err != nil {
			log.Fatalln("Error reading user input")
		}

		if string(input[:]) == "6\n" {
			break
		}

		fmt.Println()

		switch string(input[:]) {
		case "1\n":
			clusterClient.clearAllClusterData(true)
			break
		case "2\n":
			clusterClient.clearClusterDataOneByOne()
			break
		case "3\n":
			printDataAmount(settings)
			clusterClient.storeDataInCluster()
			break
		case "4\n":
			clusterClient.listClusterData()
			break
		case "5\n":
			clusterClient.getMetrics()
			break
		default:
			fmt.Printf("Unknown option\n%v", PROGRAM_INFO)
		}

		fmt.Println()
	}

	fmt.Println("\nExiting...")
}

func checkError(e error) {
	if e != nil {
		panic(e)
	}
}
