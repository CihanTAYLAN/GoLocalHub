package metrics

import (
	"encoding/json"
	"net/http"
	"runtime"
)

type Stats struct {
	Alloc        uint64 `json:"alloc"`
	TotalAlloc   uint64 `json:"totalAlloc"`
	Sys          uint64 `json:"sys"`
	NumGoroutine int    `json:"numGoroutine"`
}

func Handler(w http.ResponseWriter, r *http.Request) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	json.NewEncoder(w).Encode(Stats{
		Alloc:        m.Alloc,
		TotalAlloc:   m.TotalAlloc,
		Sys:          m.Sys,
		NumGoroutine: runtime.NumGoroutine(),
	})
}
