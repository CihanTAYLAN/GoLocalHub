package notes

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var notesDir = filepath.Join(os.Getenv("HOME"), "go-local-hub-data", "notes")

type Note struct {
	Slug      string `json:"slug"`
	Title     string `json:"title"`
	Body      string `json:"body"`
	UpdatedAt string `json:"updatedAt"`
}

// ensure dir exists
func init() { _ = os.MkdirAll(notesDir, 0o755) }

func slugify(title string) string {
	s := strings.ReplaceAll(strings.ToLower(title), " ", "-")
	return time.Now().Format("2006-01-02") + "-" + s + ".md"
}

// POST /notes → yeni veya güncelle
func SaveHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("notesDir: %s\n", notesDir)
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Slug  string `json:"slug,omitempty"` // boşsa yeni
		Title string `json:"title"`
		Body  string `json:"body"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.Slug == "" {
		req.Slug = slugify(req.Title)
	}
	path := filepath.Join(notesDir, req.Slug)
	note := Note{
		Slug:      req.Slug,
		Title:     req.Title,
		Body:      req.Body,
		UpdatedAt: time.Now().UTC().Format(time.RFC3339),
	}

	if err := os.WriteFile(path, []byte("# "+note.Title+"\n\n"+note.Body), 0o644); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(note)
}

// GET /notes?slug=… → tek not
// GET /notes         → tüm list
func ListHandler(w http.ResponseWriter, r *http.Request) {
	if slug := r.URL.Query().Get("slug"); slug != "" {
		b, err := os.ReadFile(filepath.Join(notesDir, slug))
		if err != nil {
			http.NotFound(w, r)
			return
		}
		lines := strings.SplitN(string(b), "\n", 2)
		title := strings.TrimPrefix(lines[0], "# ")
		body := ""
		if len(lines) > 1 {
			body = strings.TrimSpace(lines[1])
		}
		json.NewEncoder(w).Encode(Note{
			Slug:      slug,
			Title:     title,
			Body:      body,
			UpdatedAt: "n/a", // basit tutuyoruz
		})
		return
	}

	var notes []Note
	_ = filepath.WalkDir(notesDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}
		slug := filepath.Base(path)
		b, _ := os.ReadFile(path)
		lines := strings.SplitN(string(b), "\n", 2)
		title := strings.TrimPrefix(lines[0], "# ")
		notes = append(notes, Note{Slug: slug, Title: title})
		return nil
	})
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(notes)
}
