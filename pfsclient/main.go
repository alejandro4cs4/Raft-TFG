package main

import (
	"fmt"

	"raft-tfg.com/alejandroc/pfslib"
)

func main() {
	pfslib.PfsInit()

	fd, _ := pfslib.PfsOpen("/home/alejandroc/test/test.txt")

	buffer, _ := pfslib.PfsRead(fd, 100)

	fmt.Printf("Se ha leido:\n%s\n", string(buffer))

}
