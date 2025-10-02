package core

var KeySpaceStats []map[string]int = make([]map[string]int, 4)

func init() {
	for i := range KeySpaceStats {
		KeySpaceStats[i] = make(map[string]int)
	}
}

func updateKeySpaceStats(db int, metric string) {
	KeySpaceStats[db][metric]++
}
