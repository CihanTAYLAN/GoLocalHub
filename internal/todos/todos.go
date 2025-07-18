package todos

import (
	"bufio"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var todoFile = filepath.Join(os.Getenv("HOME"), "go-local-hub-data", "todos", "todo.txt")

func ensureFile() {
	dir := filepath.Dir(todoFile)
	_ = os.MkdirAll(dir, 0o755)
	if _, err := os.Stat(todoFile); os.IsNotExist(err) {
		_, _ = os.Create(todoFile)
	}
}

type Todo struct {
	Raw            string   `json:"raw"`
	Priority       *string  `json:"priority,omitempty"`
	Completed      bool     `json:"completed"`
	CompletionDate *string  `json:"completionDate,omitempty"`
	CreationDate   *string  `json:"creationDate,omitempty"`
	Contexts       []string `json:"contexts"`
	Projects       []string `json:"projects"`
	Text           string   `json:"text"`
}

var (
	rePriority = regexp.MustCompile(`^\(([A-Z])\) `)
	reDate     = regexp.MustCompile(`\b(\d{4}-\d{2}-\d{2})\b`)
	reContext  = regexp.MustCompile(`\B@\w+`)
	reProject  = regexp.MustCompile(`\B\+\w+`)
)

func parseLine(line string) Todo {
	t := Todo{Raw: line}
	rest := strings.TrimSpace(line)

	// completed?
	if strings.HasPrefix(rest, "x ") {
		t.Completed = true
		rest = strings.TrimPrefix(rest, "x ")
	}

	// priority
	if m := rePriority.FindStringSubmatch(rest); m != nil {
		p := m[1]
		t.Priority = &p
		rest = strings.TrimPrefix(rest, m[0])
	}

	// dates
	dates := reDate.FindAllString(rest, -1)
	if len(dates) > 0 && t.Completed {
		t.CompletionDate = &dates[0]
		if len(dates) > 1 {
			t.CreationDate = &dates[1]
		}
	} else if len(dates) > 0 {
		t.CreationDate = &dates[0]
	}

	// contexts & projects
	t.Contexts = reContext.FindAllString(rest, -1)
	t.Projects = reProject.FindAllString(rest, -1)

	// clean text
	text := rePriority.ReplaceAllString(rest, "")
	text = strings.Join(reDate.Split(text, -1), " ")
	text = reContext.ReplaceAllString(text, "")
	text = reProject.ReplaceAllString(text, "")
	t.Text = strings.TrimSpace(text)

	return t
}

// GET /todos
func ListHandler(w http.ResponseWriter, r *http.Request) {
	ensureFile()
	f, err := os.Open(todoFile)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer f.Close()

	var todos []Todo
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		l := strings.TrimSpace(sc.Text())
		if l != "" {
			todos = append(todos, parseLine(l))
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(todos)
}

// POST /todos/add  {text:"..."}
func AddHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Text string `json:"text"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ensureFile()
	f, err := os.OpenFile(todoFile, os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer f.Close()

	_, _ = f.WriteString(req.Text + "\n")

	w.WriteHeader(http.StatusCreated)
}
