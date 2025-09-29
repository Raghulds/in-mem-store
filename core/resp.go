package core

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"strconv"
)

func readLength(command []byte) (int, int) {
	length, pos := 0, 0
	for pos = range command {
		b := command[pos]
		if !(b >= '0' && b <= '9') {
			return length, pos + 2
		}

		length = (length * 10) + int(b-'0')
	}

	return length, pos + 2
}

func readString(command []byte) (string, int, error) {
	pos := 1
	for ; command[pos] != '\r'; pos++ {
	}
	return string(command[1:pos]), pos + 2, nil
}

func readBulkString(command []byte) (string, int, error) {
	log.Printf("Reading bulk string: %q", string(command))

	pos := 1
	strLength, delta := readLength(command[pos:])
	pos += delta

	log.Printf("String length: %d, pos after length: %d", strLength, pos)

	// pos is already at the start of the string (readLength handles the \r\n)
	log.Printf("Pos at start of string: %d", pos)

	// Return the string content and the total bytes consumed
	totalConsumed := pos + strLength + 2 // +2 for final \r\n
	stringContent := string(command[pos : pos+strLength])
	log.Printf("String content: %q, total consumed: %d", stringContent, totalConsumed)
	return stringContent, totalConsumed, nil
}

func readInt64(command []byte) (int64, int, error) {

	pos := 1
	_, delta := readLength(command[pos:])
	pos += delta
	var value int64 = 0
	for ; command[pos] != '\r'; pos++ {
		// Convert byte to int64
		x, err := strconv.ParseInt(string(command[pos]), 10, 64)
		if err != nil {
			return 0, 0, err
		}
		value = (value * 10) + x
	}
	return value, pos + 2, nil
}

func readArray(command []byte) ([]interface{}, int, error) {
	pos := 1
	length, delta := readLength(command[pos:])
	pos += delta

	log.Printf("Array length: %d, pos after length: %d", length, pos)

	var elements []interface{} = make([]interface{}, length)
	for i := range elements {
		log.Printf("Parsing element %d at pos %d", i, pos)
		elem, newDelta, err := DecodeOne(command[pos:])
		if err != nil {
			log.Printf("Error parsing element %d: %v", i, err)
			return nil, 0, err
		}
		elements[i] = elem
		pos += newDelta
		log.Printf("Element %d parsed, new pos: %d", i, pos)
	}
	return elements, pos, nil
}

func readError(command []byte) (string, int, error) {
	pos := 1
	for ; command[pos] != '\r'; pos++ {
	}
	return string(command[1:pos]), pos + 2, nil
}

func DecodeOne(command []byte) (interface{}, int, error) {
	cmdType := command[0]
	log.Println("Command type", cmdType)
	log.Println("Command", string(command))
	switch cmdType {
	case '*': // array
		return readArray(command)
	case '$': // bulk string
		return readBulkString(command)
	case '+': // string
		return readString(command)
	case ':': // int
		return readInt64(command)
	case '-': // error
		return readError(command)
	default:
		return "", 0, fmt.Errorf("unknown command type: %c", cmdType)
	}
}

func Decode(data []byte) ([]interface{}, error) {
	if len(data) == 0 {
		return nil, errors.New("no data")
	}
	var index int = 0
	var values []interface{} = make([]interface{}, 0)
	for index < len(data) {
		value, delta, err := DecodeOne(data[index:])
		if err != nil {
			return nil, err
		}

		index = index + delta
		values = append(values, value)
	}

	return values, nil
}

func EncodeString(str string) []byte {
	return []byte(fmt.Sprintf("$%d\r\n%s\r\n", len(str), str))
}

func Encode(str interface{}, isSimpleString bool) []byte {
	switch v := str.(type) {
	case string:
		if isSimpleString {
			return []byte(fmt.Sprintf("+%s\r\n", v))
		}
		return []byte(fmt.Sprintf("$%d\r\n%s\r\n", len(v), v))

	case int, int8, int16, int32, int64:
		return []byte(fmt.Sprintf(":%d\r\n", v))

	case error:
		return []byte(fmt.Sprintf("-%s\r\n", v))

	case []string:
		var b []byte
		buf := bytes.NewBuffer(b)
		for _, s := range str.([]string) {
			buf.Write(EncodeString(s))
		}
		return []byte(fmt.Sprintf("*%d\r\n%s", len(v), buf.Bytes()))
	default:
		return RESP_NIL
	}
}
