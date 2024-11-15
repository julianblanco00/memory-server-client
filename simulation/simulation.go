package main

import (
	"fmt"
	"math/rand"
	"time"

	memoryserver "github.com/julianblanco00/memory-server-client"
)

const (
	clients       = 100
	maxIterations = 20
)

func disconnect(idx int, done chan<- int, ms *memoryserver.MemoryServer) {
	err := ms.Disconnect()
	if err != nil {
		fmt.Println("error desconnecting client...", err)
	}
	done <- idx
}

func writeToServer(idx int, done chan<- int, ms *memoryserver.MemoryServer) {
	iterations := 0

	for {
		if iterations == maxIterations {
			disconnect(idx, done, ms)
			return
		}

		k := fmt.Sprintf("mykey-%d", iterations)

		_, err := ms.Set(k, "myvalue")
		_, err = ms.Get(k)

		iterations++

		if err != nil {
			disconnect(idx, done, ms)
			return
		}
	}
}

func simulateClient(idx int, done chan<- int) {
	ms := memoryserver.New("localhost", "4444")
	err := ms.Connect()
	if err != nil {
		fmt.Printf("Error connecting client %d: %v\n", idx, err)
		done <- idx
		return
	}

	writeToServer(idx, done, ms)
}

func main() {
	done := make(chan int)

	for v := range clients {
		go simulateClient(v, done)
	}

	for {
		id := <-done
		random := rand.Intn(3)
		fmt.Printf("Client %d done, restarting in %d seconds...\n", id, random)
		time.Sleep(time.Duration(random) * time.Second)
		go simulateClient(id, done)
	}
}
