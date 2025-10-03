package core

import (
	"in-mem-store/config"
	"log"
)

func evictFirst() {
	for key := range store {
		Del(key)
		return
	}
}

func evictAllKeysRandom() {
	evictCount := int(config.EvictionRatio * float32(config.EvictionLimit))

	log.Println("Evicting all keys random")
	// Golang dictionary iteration is considered to be random as it depends on the hash of the inserted key
	for k := range store {
		Del(k)
		evictCount--
		if evictCount <= 0 {
			break
		}
	}
}

func evict() {
	switch config.EvictionStrategy {
	case "simple-first":
		evictFirst()
	case "allkeys-random":
		evictAllKeysRandom()
	default:
		evictFirst()
	}
}
