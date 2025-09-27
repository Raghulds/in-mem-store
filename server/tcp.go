package server

import (
	"fmt"
	"in-mem-store/config"
	"in-mem-store/core"
	"io"
	"log"
	"net"
	"strconv"
)

func toArrayString(ai []interface{}) ([]string, error) {
	arr := make([]string, len(ai))
	for i := range ai {
		arr[i] = ai[i].(string)
	}
	return arr, nil
}

/*
TCP is always open stream - sends bytes to the connected server & always open until closed from either ends
HTTP does various mechanisms to understnad the end of the request
Closes conn after,
HTTP 1.0 - Content-length header or inferred by TCP conn close
HTTP 1.1 - Content-length header or Transfer-Encoding: chunked header. Persistent conn are default. Kept alive & reused
HTTP 2 - Open always. No reliance on the above headers. Relies on END_STREAM flag in stream.
*/
func readCommands(sock io.ReadWriter) (core.RedisCmds, error) {
	var buf []byte = make([]byte, 0, 1024)
	temp := make([]byte, 512)

	var cmds []*core.RedisCmd = make([]*core.RedisCmd, 0)
	for {
		n, err := sock.Read(temp)
		if err != nil {
			return nil, err
		}
		log.Println("n", n)

		buf = append(buf, temp[:n]...)

		// If buffer gets too large, something's wrong
		if len(buf) > 4096 {
			return nil, fmt.Errorf("command too large")
		}

		// Try to decode - if it works, we have a complete command
		values, err := core.Decode(buf)
		if err != nil {
			return nil, err
		}

		for _, value := range values {
			log.Println(value)
			tokens, err := toArrayString(value.([]interface{}))
			if err != nil {
				return nil, err
			}

			cmds = append(cmds, &core.RedisCmd{
				Cmd:  tokens[0],
				Args: tokens[1:],
			})
		}

		break
	}

	return cmds, nil
}

func respond(cmds core.RedisCmds, sock io.ReadWriter) {
	core.EvalAndRespond(cmds, sock)
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
		cmd, err := readCommands(socket)
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
