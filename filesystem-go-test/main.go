package main

import (
	"fmt"
	"log"
	"os/exec"
	"strings"
	"time"
)

// find /usr/ -printf "%p:%y:%m:%U:%G:%A@:%s:%i\n"
// find /usr/ -printf "file_name : file_type : permission_bits : user_id : group_id : last_access_time : file_size : inode_number"

type KeyValue struct {
	Key string
	Value string
}

var systemData []KeyValue

func main() {
	out, err := exec.Command("find", "/usr/", "-printf", "%p:%y:%m:%U:%G:%A@:%s:%i\n").Output()

	if err != nil {
		log.Fatalln("Command():", err)
	}

	output := string(out)

	lines := strings.Split(output, "\n")

	startTime := time.Now()

	for _, line := range lines {
		key, value, _ := strings.Cut(line, ":")

		newData := KeyValue{
			Key: key,
			Value: value,
		}

		systemData = append(systemData, newData)
	}

	elapsed := time.Since(startTime)

	fmt.Printf("Data parsing took %vms\n", elapsed.Milliseconds())
}