package main

import (
	"raft-tfg.com/alejandroc/pfslib"
)

func main() {
	pfslib.PfsInit()

	pfslib.PfsOpen("/home/alejandroc/test/test.txt")

}
