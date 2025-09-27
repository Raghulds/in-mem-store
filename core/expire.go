package core

import (
	"log"
	"time"
)

// Central Limit Theorem
func expireSample() float32 {
	limit := 20
	var expiredKeysCount float32 = 0

	for key, val := range store {
		if val.ExpiresAt != -1 {
			limit--
			if val.ExpiresAt < time.Now().UnixMilli() {
				Del(key)
				expiredKeysCount++
			}
		}

		if limit == 0 {
			break
		}
	}

	return expiredKeysCount / 20
}

func DeleteExpiredKeys() {
	for {
		samplingFactor := expireSample()

		if samplingFactor < 0.25 {
			break
		}
	}
	log.Printf("Sampled for deleting expired keys. Total keys in store: %v", len(store))

}
