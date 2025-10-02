package core

import (
	"in-mem-store/config"
	"log"
	"time"
)

var store map[string]*Obj

func init() {
	store = make(map[string]*Obj)
}

func NewObj(value interface{}, durationInMs int64, typeB uint8, encodingB uint8) *Obj {
	var expiresAt int64 = -1
	if durationInMs > 0 {
		expiresAt = time.Now().UnixMilli() + durationInMs
	}

	return &Obj{
		Value:        value,
		TypeEncoding: typeB | encodingB,
		ExpiresAt:    expiresAt,
	}
}

func Put(key string, value *Obj) {
	if len(store) >= config.EvictionLimit {
		evict()
	}
	store[key] = value
}

func Get(key string) *Obj {
	v := store[key]
	if v != nil {
		if v.ExpiresAt != -1 && v.ExpiresAt < time.Now().UnixMilli() {
			// Passive deletion
			log.Println("PASSIVE deletion")
			delete(store, key)
			return nil
		}
	}
	return store[key]
}

func Del(key string) bool {
	if _, ok := store[key]; ok {
		delete(store, key)
		return true
	}
	return false
}
