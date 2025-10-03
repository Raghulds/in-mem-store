package config

var (
	Host             string
	Port             int
	EvictionLimit    int     = 100
	EvictionStrategy string  = "allkeys-random"
	EvictionRatio    float32 = 0.40
	AOF_File         string  = "./aof-log.aof"
)
