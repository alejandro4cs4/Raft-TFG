package utils

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"

	"raft-tfg.com/alejandroc/pfslib/globals"
)

func ReadSettings() {
	settingsFileContent, err := os.ReadFile("./settings.yaml")
	CheckError(err)

	yaml.Unmarshal(settingsFileContent, &globals.PfsSettings)
}

func GetAbsolutePath(pathname string) string {
	absolutePath, _ := filepath.Abs(pathname)

	return absolutePath
}

func MapRouteComponentName(routeComponentName string) string {
	if routeComponentName == "" {
		return "nil"
	}

	return routeComponentName
}

func CheckError(e error) {
	if e != nil {
		panicMsg := fmt.Sprintf("[pfslib]: %v\n", e)
		panic(errors.New(panicMsg))
	}
}
