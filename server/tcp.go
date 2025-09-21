package server

import (
	"fmt"
	"in-mem-store/config"
	"in-mem-store/core"
	"io"
	"log"
	"net"
	"strconv"
	"strings"
)

/*
TCP is always open stream - sends bytes to the connected server & always open until closed from either ends
HTTP does various mechanisms to understnad the end of the request
Closes conn after,
HTTP 1.0 - Content-length header or inferred by TCP conn close
HTTP 1.1 - Content-length header or Transfer-Encoding: chunked header. Persistent conn are default. Kept alive & reused
HTTP 2 - Open always. No reliance on the above headers. Relies on END_STREAM flag in stream.
*/
func readCommand(sock io.ReadWriter) (*core.RedisCmd, error) {
	var buf []byte = make([]byte, 0, 1024)
	temp := make([]byte, 512)

	for {
		n, err := sock.Read(temp)
		if err != nil {
			return nil, err
		}

		buf = append(buf, temp[:n]...)

		// Try to decode - if it works, we have a complete command
		tokens, err := core.DecodeArrayString(buf)
		if err == nil {
			cmd := core.RedisCmd{
				Cmd:  strings.ToUpper(tokens[0]),
				Args: tokens[1:],
			}
			return &cmd, nil
		}

		// If buffer gets too large, something's wrong
		if len(buf) > 4096 {
			return nil, fmt.Errorf("command too large")
		}
	}
}

func respondError(err error, sock io.ReadWriter) {
	sock.Write([]byte(fmt.Sprintf("-ERR %v\r\n", err)))
}

func respond(cmd *core.RedisCmd, sock io.ReadWriter) {
	err := core.EvalAndRespond(cmd, sock)
	if err != nil {
		respondError(err, sock)
	}
}

func RunSyncTcpServer() {
	log.Println("Starting TCP server..", config.Host, config.Port)
	var con_clients = 0

	// Listening to the configured host:port
	listener, err := net.Listen("tcp", config.Host+":"+strconv.Itoa(config.Port))
	if err != nil {
		panic(err)
	}

	log.Println("TCP server listening on", config.Host+":"+strconv.Itoa(config.Port))
	socket, err := listener.Accept()
	if err != nil {
		panic(err)
	}
	con_clients += 1
	log.Println("Client connected! ", socket.RemoteAddr(), " concurrent clients: ", con_clients)

	for {
		cmd, err := readCommand(socket)
		if err != nil {
			socket.Close()
			con_clients -= 1
			log.Println("Client disconnected", socket.RemoteAddr(), "Concurrent clients ", con_clients)
			if err != nil {
				break
			}
			log.Println("Disconnected err:", err)
		}
		log.Println("command", cmd)

		// string(value) is type conversion. Can be used if we know the type of the value.
		// value.(string) is type assertion. interface{} to string can be done by type assertion. Panics if assertion fails
		respond(cmd, socket)
	}
}
