package main

import (
	"fmt"

	"raft-tfg.com/alejandroc/pfslib"
)

func main() {
	pfslib.PfsInit()

	fd, _ := pfslib.PfsOpen("/home/alejandroc/test/test.txt")

	buffer := make([]byte, 100)
	bytesRead, _ := pfslib.PfsRead(fd, buffer)

	fmt.Printf("Se ha leido %d bytes\nEl contenido es:\n%s\n", bytesRead, string(buffer))

}
