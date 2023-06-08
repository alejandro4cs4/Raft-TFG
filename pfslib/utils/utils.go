package utils

import (
	"os"

	"gopkg.in/yaml.v3"

	"raft-tfg.com/alejandroc/pfslib/globals"
)

func ReadSettings() {
	settingsFileContent, err := os.ReadFile("./settings.yaml")
	CheckError(err)

	yaml.Unmarshal(settingsFileContent, &globals.PfsSettings)
}

func CheckError(e error) {
	if e != nil {
		panic(e)
	}
}
