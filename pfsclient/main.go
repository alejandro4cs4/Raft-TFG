package main

import (
	"os"

	"raft-tfg.com/alejandroc/pfslib"
)

func main() {
	pfslib.PfsInit()

	pfsFile, _ := pfslib.PfsOpen("/home/alejandroc/test/test.txt")

	pfsFile.Lseek(10, os.SEEK_CUR)

	writeBuffer := []byte("SOY ALEALEJANDRO")
	pfsFile.Write(writeBuffer)

}
