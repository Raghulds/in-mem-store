package core

import (
	"errors"
	"fmt"
	"io"
	"log"
)

func evalPING(args []string, sock io.ReadWriter) error {
	var b []byte
	if len(args) >= 2 {
		return errors.New("ERR wrong number of arguments from 'ping' command")
	}

	if len(args) == 0 {
		b = Encode("PONG", true)
	} else {
		b = Encode(args[0], false)
	}

	_, err := sock.Write(b)
	return err
}

func EvalAndRespond(cmd *RedisCmd, sock io.ReadWriter) error {
	log.Println("evalresp", cmd, sock)
	switch cmd.Cmd {
	case "PING":
		return evalPING(cmd.Args, sock)
	default:
		// Return error for unknown commands
		errorMsg := fmt.Sprintf("ERR unknown command '%s'", cmd.Cmd)
		_, err := sock.Write([]byte(fmt.Sprintf("-ERR %s\r\n", errorMsg)))
		return err
	}
}
