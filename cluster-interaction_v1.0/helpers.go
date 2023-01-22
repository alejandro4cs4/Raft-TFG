package main

import (
	"fmt"
	"log"
	"os/exec"
)

func printDataAmount(settings *Settings) {
	if settings.MaxItemsStore > 0 {
		fmt.Printf("%d key-value pairs will be stored in %s cluster\n", settings.MaxItemsStore, settings.ClusterType)
		return
	}

	numItemsCmd := fmt.Sprintf("find %s | wc -l", settings.ExploreDirectory)

	out, err := exec.Command("bash", "-c", numItemsCmd).Output()

	if err != nil {
		log.Panicf("exec.Command(find %s | wc -l): %v\n", settings.ExploreDirectory, err)
	}

	fmt.Printf("%v key-value pairs will be stored in %s cluster\n", string(out[:len(out)-1]), settings.ClusterType)
}