package core

import (
	"errors"
	"fmt"
	"io"
	"log"
	"strconv"
	"time"
)

var RESP_NIL []byte = []byte("$-1\r\n")

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

func evalSET(args []string, sock io.ReadWriter) error {
	if len(args) < 2 {
		return errors.New("not a valid args to set")
	}

	var key, value string
	var expiryDuration int64 = -1
	key, value = args[0], args[1]

	for i := 2; i < len(args); i++ {
		switch args[i] {
		case "EX", "ex":
			i++
			if i == len(args) {
				return errors.New("syntax error")
			}

			exDurationSec, err := strconv.ParseInt(args[3], 10, 64)
			if err != nil {
				return errors.New("expiry value is not an integer or out of range")
			}
			expiryDuration = exDurationSec * 1000
		}
	}

	obj := NewObj(value, expiryDuration)
	Put(key, obj)
	sock.Write([]byte("+OK\r\n"))
	return nil
}

func evalGET(args []string, sock io.ReadWriter) error {
	if len(args) < 1 {
		return errors.New("not a valid args to get")
	}

	var key string = args[0]
	val := Get(key)

	if val == nil {
		sock.Write(RESP_NIL)
		return nil
	}

	if val.ExpiresAt != -1 && val.ExpiresAt <= time.Now().UnixMilli() {
		sock.Write((RESP_NIL))
		return nil
	}

	encoded := Encode(val.Value, false)
	sock.Write(encoded)
	return nil
}

func evalTTL(args []string, sock io.ReadWriter) error {
	if len(args) < 1 {
		return errors.New("not a valid args to ttl")
	}

	var key = args[0]
	var obj = Get(key)
	if obj == nil {
		sock.Write([]byte(":-2\r\n"))
		return nil
	}

	if obj.ExpiresAt == -1 {
		sock.Write([]byte(":-1\r\n"))
		return nil
	}

	if time.Now().UnixMilli() > obj.ExpiresAt {
		sock.Write([]byte(":-2\r\n"))
		return nil
	}

	msRemaining := obj.ExpiresAt - time.Now().UnixMilli()
	secs := msRemaining / 1000
	sock.Write(Encode(secs, false))
	return nil
}

func evalEXPIRY(args []string, sock io.ReadWriter) error {
	if len(args) < 2 {
		return errors.New("not a valid args to set")
	}

	var key string
	var expiryDuration int64 = -1
	key = args[0]
	value := Get(key)
	if value == nil || (value.ExpiresAt != -1 && value.ExpiresAt < time.Now().UnixMilli()) {
		sock.Write(RESP_NIL)
		return nil
	}

	exDurationSec, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		return errors.New("expiry value is not an integer or out of range")
	}
	expiryDuration = exDurationSec * 1000

	value.ExpiresAt = time.Now().UnixMilli() + expiryDuration
	Put(key, value)
	sock.Write([]byte("+OK\r\n"))
	return nil
}

func evalDELETE(args []string, sock io.ReadWriter) error {
	deletedCount := 0
	for i := 0; i < len(args); i++ {
		if ok := Del(args[i]); ok {
			deletedCount++
		}
	}
	sock.Write(Encode(deletedCount, false))
	return nil
}

func EvalAndRespond(cmd *RedisCmd, sock io.ReadWriter) error {
	log.Println("evalresp", cmd, sock)
	switch cmd.Cmd {
	case "PING":
		return evalPING(cmd.Args, sock)
	case "SET":
		return evalSET(cmd.Args, sock)
	case "GET":
		return evalGET(cmd.Args, sock)
	case "TTL":
		return evalTTL(cmd.Args, sock)
	case "DELETE":
		return evalDELETE(cmd.Args, sock)
	case "EXPIRY":
		return evalEXPIRY(cmd.Args, sock)
	default:
		errorMsg := fmt.Sprintf("ERR unknown command '%s'", cmd.Cmd)
		_, err := sock.Write([]byte(fmt.Sprintf("-ERR %s\r\n", errorMsg)))
		return err
	}
}
