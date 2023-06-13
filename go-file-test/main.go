package main

import (
	"fmt"
	"os"
)

func main() {
	filepath := "/home/alejandroc/test/ae789d32-a5d6-4638-9792-c53793e43135"

	// file, _ := os.OpenFile(filepath, os.O_RDWR, 0755)

	// file.Seek(10, 0)

	// writeBuffer := []byte("HOLA SOY ALEJANDRO REMOTO")
	// file.Write(writeBuffer)

	// file.Close()

	file, _ := os.Open(filepath)

	readBuffer := make([]byte, 100)
	file.Read(readBuffer)
	fmt.Printf("Leído: %s\n", string(readBuffer))

	file.Close()
	fmt.Printf("Closed: %v\n", file)

	readBuffer = make([]byte, 50)
	bytesRead, err := file.Read(readBuffer)
	fmt.Printf("Read: %d / Err: %v\n", bytesRead, err)
}
