package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type project struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type projectStore struct {
	mu       sync.Mutex
	path     string
	projects map[string][]project
	nextID   int
}

type diskState struct {
	Projects map[string][]project `json:"projects"`
	NextID   int                  `json:"next_id"`
}

func openProjectStore(path string) (*projectStore, error) {
	store := &projectStore{path: path, projects: make(map[string][]project)}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return store, nil
		}
		return nil, err
	}
	var state diskState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("decode project store: %w", err)
	}
	if state.Projects != nil {
		store.projects = state.Projects
	}
	store.nextID = state.NextID
	return store, nil
}

func (s *projectStore) persist() error {
	data, err := json.Marshal(diskState{Projects: s.projects, NextID: s.nextID})
	if err != nil {
		return err
	}
	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	f, err := os.CreateTemp(dir, ".projects-*.tmp")
	if err != nil {
		return err
	}
	tmp := f.Name()
	defer os.Remove(tmp)
	if err := f.Chmod(0o600); err != nil {
		f.Close()
		return err
	}
	if _, err := f.Write(data); err != nil {
		f.Close()
		return err
	}
	if err := f.Sync(); err != nil {
		f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmp, s.path); err != nil {
		return err
	}
	d, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer d.Close()
	if err := d.Sync(); err != nil {
		return err
	}
	return nil
}

func (s *projectStore) create(tenant, name string) (project, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.nextID++
	p := project{ID: "project-" + fmt.Sprint(s.nextID), Name: name}
	s.projects[tenant] = append(s.projects[tenant], p)
	if err := s.persist(); err != nil {
		s.projects[tenant] = s.projects[tenant][:len(s.projects[tenant])-1]
		s.nextID--
		return project{}, err
	}
	return p, nil
}

func (s *projectStore) list(tenant string) []project {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]project{}, s.projects[tenant]...)
}

func tenantID(value string) (string, bool) {
	value = strings.TrimSpace(value)
	return value, value != ""
}
