package main

// import (
// 	"bufio"
// 	"context"
// 	"fmt"
// 	"log"
// 	"os/exec"
// 	"strings"
// 	"time"

// 	"github.com/tikv/client-go/v2/rawkv"
// )

// func printDataAmount() {
// 	numFilesCmd := "find /usr/ | wc -l"

// 	out, err := exec.Command("bash", "-c", numFilesCmd).Output()

// 	if err != nil {
// 		log.Panicf("exec.Command(find /usr/ | wc -l): %v\n", err)
// 	}

// 	fmt.Printf("%v files will be stored in TiKV cluster\n", string(out[:len(out)-1]))
// }

// func storeDataInCluster(cli *rawkv.Client, ctx context.Context) {
// 	cmd := exec.Command("find", "/usr/", "-printf", "%p:%y:%m:%U:%G:%A@:%s:%i\n")
// 	stdout, err := cmd.StdoutPipe()

// 	if err != nil {
// 		log.Panicf("cmd.StdoutPipe(): %v\n", err)
// 	}

// 	cmd.Start()

// 	scanner := bufio.NewScanner(stdout)

// 	// Start timer
// 	startTime := time.Now()

// 	for scanner.Scan() {
// 		key, value, _ := strings.Cut(scanner.Text(), ":")

// 		err := cli.Put(ctx, []byte(key), []byte(value))

// 		if err != nil {
// 			log.Panicf("cli.Put(%v, %v): %v\n", key, value, err)
// 		}
// 	}

// 	// Stop timer
// 	elapsedTime := time.Since(startTime)

// 	cmd.Wait()

// 	fmt.Printf("It took %d ms / %.2f sec / %.2f min to store the data in TiKV\n", elapsedTime.Milliseconds(), elapsedTime.Seconds(), elapsedTime.Minutes())
// }

// func main() {
// 	// Connect to TiKV cluster
// 	ctx := context.Background()

// 	cli, err := rawkv.NewClient(ctx, []string{"127.0.0.1:2379"}, config.DefaultConfig().Security)

// 	if err != nil {
// 		log.Panicf("rawkv.NewClient(): %v\n", err)
// 	}

// 	defer cli.Close()

// 	log.Default().Println("Connected to TiKV cluster successfully")

// 	printDataAmount()

// 	storeDataInCluster(cli, ctx)
// }
