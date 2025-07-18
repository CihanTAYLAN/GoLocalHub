package clipboard

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"
)

type Item struct {
	ID      string    `json:"id"`
	Content string    `json:"content"`
	Exp     time.Time `json:"expiresAt"`
}

// in-memory store
var (
	store = make(map[string]Item)
	mu    sync.RWMutex
)

func SetHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Content string `json:"content"`
		TTL     int    `json:"ttl"` // saniye
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	id := time.Now().UTC().Format("150405.000") // HHMMSS.mmm
	mu.Lock()
	store[id] = Item{
		ID:      id,
		Content: req.Content,
		Exp:     time.Now().Add(time.Duration(req.TTL) * time.Second),
	}
	mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"id": id})
}

func GetHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "id required", http.StatusBadRequest)
		return
	}

	mu.RLock()
	item, ok := store[id]
	mu.RUnlock()
	if !ok || time.Now().After(item.Exp) {
		http.Error(w, "not found or expired", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(item)
}

// arka planda temizlik
func init() {
	go func() {
		for range time.Tick(30 * time.Second) {
			now := time.Now()
			mu.Lock()
			for k, v := range store {
				if now.After(v.Exp) {
					delete(store, k)
				}
			}
			mu.Unlock()
		}
	}()
}
