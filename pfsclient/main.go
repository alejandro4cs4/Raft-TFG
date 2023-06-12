package main

import (
	"raft-tfg.com/alejandroc/pfslib"
)

func main() {
	pfslib.PfsInit()

	fd, _ := pfslib.PfsOpen("/home/alejandroc/test/test.txt")

	pfslib.PfsLseek(fd, 10, 1)

	writeBuffer := []byte("HOLA SOY ALEJANDRO")
	pfslib.PfsWrite(fd, writeBuffer)

}
