package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

type project struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type projectStore struct {
	mu       sync.RWMutex
	projects map[string][]project
	nextID   int
}

func (s *projectStore) create(tenant, name string) project {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.nextID++
	p := project{ID: "project-" + strconv.Itoa(s.nextID), Name: name}
	s.projects[tenant] = append(s.projects[tenant], p)
	return p
}

func (s *projectStore) list(tenant string) []project {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]project{}, s.projects[tenant]...)
}

func newHandler() http.Handler {
	store := &projectStore{projects: make(map[string][]project)}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})
	mux.HandleFunc("/projects", func(w http.ResponseWriter, r *http.Request) {
		tenant := strings.TrimSpace(r.Header.Get("X-Tenant-ID"))
		if tenant == "" {
			http.Error(w, "tenant identifier is required", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodGet:
			_ = json.NewEncoder(w).Encode(store.list(tenant))
		case http.MethodPost:
			var input struct {
				Name string `json:"name"`
			}
			if err := json.NewDecoder(r.Body).Decode(&input); err != nil || strings.TrimSpace(input.Name) == "" {
				http.Error(w, "valid project name is required", http.StatusBadRequest)
				return
			}
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(store.create(tenant, strings.TrimSpace(input.Name)))
		default:
			w.Header().Set("Allow", "GET, POST")
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
	return mux
}

func main() {
	_ = http.ListenAndServe(":8080", newHandler())
}
