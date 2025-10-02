package core

import (
	"bytes"
	"errors"
	"io"
	"log"
	"strconv"
	"time"
)

var RESP_NIL []byte = []byte("$-1\r\n")
var RESP_MINUS_1 []byte = []byte(":-1\r\n")
var RESP_MINUS_2 []byte = []byte(":-2\r\n")
var RESP_ONE []byte = []byte(":1\r\n")
var RESP_ZERO []byte = []byte(":0\r\n")
var RESP_OK []byte = []byte("+OK\r\n")

func evalPING(args []string) []byte {
	var b []byte
	if len(args) >= 2 {
		return Encode(errors.New("ERR wrong number of arguments from 'ping' command"), false)
	}

	if len(args) == 0 {
		b = Encode("PONG", true)
	} else {
		b = Encode(args[0], false)
	}

	return b
}

func evalSET(args []string) []byte {
	if len(args) < 2 {
		return Encode(errors.New("not a valid args to set"), false)
	}

	var key, value string
	var expiryDuration int64 = -1
	key, value = args[0], args[1]

	typeB, encodingB := deduceTypeEncoding(value)
	for i := 2; i < len(args); i++ {
		switch args[i] {
		case "EX", "ex":
			i++
			if i == len(args) {
				return Encode(errors.New("syntax error"), false)
			}

			exDurationSec, err := strconv.ParseInt(args[3], 10, 64)
			if err != nil {
				return Encode(errors.New("expiry value is not an integer or out of range"), false)
			}
			expiryDuration = exDurationSec * 1000
		}
	}

	obj := NewObj(value, expiryDuration, typeB, encodingB)
	Put(key, obj)
	return RESP_OK
}

func evalGET(args []string) []byte {
	if len(args) < 1 {
		return Encode(errors.New("not a valid args to get"), false)
	}

	var key string = args[0]
	val := Get(key)

	if val == nil {
		return RESP_NIL
	}

	if val.ExpiresAt != -1 && val.ExpiresAt <= time.Now().UnixMilli() {
		return RESP_NIL
	}

	encoded := Encode(val.Value, false)
	return encoded
}

func evalTTL(args []string) []byte {
	if len(args) < 1 {
		return Encode(errors.New("not a valid args to ttl"), false)
	}

	var key = args[0]
	var obj = Get(key)
	if obj == nil {
		return RESP_MINUS_2
	}

	if obj.ExpiresAt == -1 {
		return RESP_MINUS_1
	}

	if time.Now().UnixMilli() > obj.ExpiresAt {
		return RESP_MINUS_2
	}

	msRemaining := obj.ExpiresAt - time.Now().UnixMilli()
	secs := msRemaining / 1000
	return Encode(secs, false)
}

func evalEXPIRY(args []string) []byte {
	if len(args) < 2 {
		return Encode(errors.New("not a valid args to set"), false)
	}

	var key string
	var expiryDuration int64 = -1
	key = args[0]
	value := Get(key)
	if value == nil || (value.ExpiresAt != -1 && value.ExpiresAt < time.Now().UnixMilli()) {
		return RESP_NIL
	}

	exDurationSec, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		return Encode(errors.New("expiry value is not an integer or out of range"), false)
	}
	expiryDuration = exDurationSec * 1000

	value.ExpiresAt = time.Now().UnixMilli() + expiryDuration
	Put(key, value)
	return RESP_OK
}

func evalDELETE(args []string) []byte {
	deletedCount := 0
	for i := 0; i < len(args); i++ {
		if ok := Del(args[i]); ok {
			deletedCount++
		}
	}
	return Encode(deletedCount, false)
}

func evalBGREWRITEAOF() []byte {
	dumpAllAOF()
	return RESP_OK
}

func evalINCR(args []string) []byte {
	if len(args) != 1 {
		return Encode(errors.New("ERR wrong number of arguments for 'incr' command"), false)
	}

	var key string = args[0]
	var obj = Get(key)
	if obj == nil {
		obj = NewObj("0", -1, OBJ_TYPE_STRING, OBJ_ENCODING_INT)
		Put(key, obj)
	}

	if err := assertType(obj.TypeEncoding, OBJ_TYPE_STRING); err != nil {
		return Encode(err, false)
	}

	if err := assertEncoding(obj.TypeEncoding, OBJ_ENCODING_INT); err != nil {
		return Encode(err, false)
	}

	i, _ := strconv.ParseInt(obj.Value.(string), 10, 64)
	i++
	obj.Value = strconv.FormatInt(i, 10)
	return Encode(i, false)
}

func EvalAndRespond(cmds RedisCmds, sock io.ReadWriter) {
	log.Println("evalresp", cmds, sock)
	var buf bytes.Buffer

	for _, cmd := range cmds {
		switch cmd.Cmd {
		case "PING":
			buf.Write(evalPING(cmd.Args))
		case "SET":
			buf.Write(evalSET(cmd.Args))
		case "GET":
			buf.Write(evalGET(cmd.Args))
		case "TTL":
			buf.Write(evalTTL(cmd.Args))
		case "DEL":
			buf.Write(evalDELETE(cmd.Args))
		case "EXPIRE":
			buf.Write(evalEXPIRY(cmd.Args))
		case "INCR":
			buf.Write(evalINCR(cmd.Args))
		case "BGREWRITEAOF":
			buf.Write(evalBGREWRITEAOF())
		default:
			buf.Write(evalPING(cmd.Args))
		}
	}
	sock.Write(buf.Bytes())
}
