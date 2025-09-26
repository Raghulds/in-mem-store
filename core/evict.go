package core

import "in-mem-store/config"

func evictFirst() {
	for key := range store {
		Del(key)
		return
	}
}

func evict() {
	switch config.EvictionStrategy {
	case "simple-first":
		evictFirst()
	default:
		evictFirst()
	}
}
