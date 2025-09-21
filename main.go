package main

import (
	"flag"
	"in-mem-store/config"
	"in-mem-store/server"
	"log"
)

func setupFlags() {
	flag.StringVar(&config.Host, "host", "0.0.0.0", "host for the server")
	flag.IntVar(&config.Port, "port", 8389, "port for the server")
	flag.Parse()
}

func main() {
	setupFlags()
	log.Println("Starting In-memory store")
	server.RunAsyncTcpServer()
}
