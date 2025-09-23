package core

import "time"

var store map[string]*Obj

type Obj struct {
	Value     interface{}
	ExpiresAt int64
}

func init() {
	store = make(map[string]*Obj)
}

func NewObj(value interface{}, durationInMs int64) *Obj {
	var expiresAt int64 = -1
	if durationInMs > 0 {
		expiresAt = time.Now().UnixMilli() + durationInMs
	}

	return &Obj{
		Value:     value,
		ExpiresAt: expiresAt,
	}
}

func Put(key string, value *Obj) {
	store[key] = value
}

func Get(key string) *Obj {
	return store[key]
}
