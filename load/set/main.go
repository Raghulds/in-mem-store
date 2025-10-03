package main

import (
	"fmt"
	"in-mem-store/core"
	"io"
	"log"
	"math/rand"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

func getRandomKeyValue() (string, int64) {
	token := int64(rand.Uint64()) % 5000000
	if token < 0 {
		token = -(token)
	}
	return "k" + strconv.FormatInt(token, 10), token
}

func stormSet(wg *sync.WaitGroup) {
	defer wg.Done()
	conn, err := net.Dial("tcp", "localhost:8389")
	if err != nil {
		panic(conn)
	}

	for {
		time.Sleep(500 * time.Millisecond)
		key, value := getRandomKeyValue()
		var buf [512]byte
		cmd := fmt.Sprintf("SET %s %d", key, value)
		fmt.Println(cmd)
		_, err := conn.Write(core.Encode(strings.Split(cmd, " "), false))
		if err != nil {
			log.Printf("err: %v", err)
			panic(err)
		}
		_, err = conn.Read(buf[:])
		if err != nil {
			if err == io.EOF {
				return
			}
			log.Printf("err: %v", err)
			panic(err)
		}
	}
	conn.Close()
}

func main() {
	var wg sync.WaitGroup
	for i := 0; i < 5; {
		wg.Add(1)
		go stormSet(&wg)
		i++
	}
	wg.Wait()
}
