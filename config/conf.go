package config

var (
	Host             string
	Port             int
	EvictionLimit    int    = 5
	EvictionStrategy string = "simple-string"
)
