package main

import (
	"fmt"
	"log"
	"os/exec"
	"strconv"
)

func printDataAmount(settings *Settings) {
	numItemsCmd := fmt.Sprintf("find %s | wc -l", settings.ExploreDirectory)
	
	numItemsAvailableStr, err := exec.Command("bash", "-c", numItemsCmd).Output()
	if err != nil {
		log.Panicf("exec.Command(find %s | wc -l): %v\n", settings.ExploreDirectory, err)
	}

	numItemsAvailable, err := strconv.Atoi(string(numItemsAvailableStr[:len(numItemsAvailableStr)-1]))
	if err != nil {
		log.Panicf("strconv.Atoi(%v): %v\n", string(numItemsAvailableStr[:len(numItemsAvailableStr)-1]), err)
	}

	if settings.MaxItemsStore > 0 && numItemsAvailable > int(settings.MaxItemsStore) {
		fmt.Printf("%d key-value pairs will be stored in %s cluster\n", settings.MaxItemsStore, settings.ClusterType)
		return
	}

	fmt.Printf("%v key-value pairs will be stored in %s cluster\n", numItemsAvailable, settings.ClusterType)
}