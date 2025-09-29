package core

import (
	"fmt"
	"in-mem-store/config"
	"log"
	"os"
	"strings"
)

func dumpKey(file *os.File, key string, value *Obj) {
	cmd := fmt.Sprintf("SET %s %s", key, value.Value)
	tokens := strings.Split(cmd, " ")
	log.Println(tokens)
	bytes := Encode(tokens, false)
	file.Write(bytes)
}

func dumpAllAOF() {
	file, err := os.OpenFile(config.AOF_File, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		fmt.Print("error: ", err)
		return
	}

	log.Println("rewriting AOF file at: ", config.AOF_File)
	for k, obj := range store {
		dumpKey(file, k, obj)
	}
	log.Println("AOF rewrite complete!")
	file.Close()
}
