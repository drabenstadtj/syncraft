package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type Server struct {
	dataDir string
}

type manifest struct {
	Count int `json:"count"`
}

func New(dataDir string) *Server {
	return &Server{dataDir: dataDir}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /worlds/{name}", s.handleGet)
	mux.HandleFunc("POST /worlds/{name}", s.handlePost)
	return mux
}

func (s *Server) handleGet(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if !validName(name) {
		http.Error(w, "invalid world name", http.StatusBadRequest)
		return
	}

	m, err := s.loadManifest(name)
	if err != nil || m.Count == 0 {
		http.Error(w, "no diff for world", http.StatusNotFound)
		return
	}

	path := s.diffPath(name, m.Count)
	f, err := os.Open(path)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer f.Close()

	w.Header().Set("Content-Type", "application/octet-stream")
	io.Copy(w, f)
}

func (s *Server) handlePost(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if !validName(name) {
		http.Error(w, "invalid world name", http.StatusBadRequest)
		return
	}

	worldDir := filepath.Join(s.dataDir, name)
	if err := os.MkdirAll(worldDir, 0755); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	m, err := s.loadManifest(name)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	next := m.Count + 1
	f, err := os.Create(s.diffPath(name, next))
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer f.Close()

	if _, err := io.Copy(f, r.Body); err != nil {
		http.Error(w, "write failed", http.StatusInternalServerError)
		return
	}
	f.Close()

	m.Count = next
	if err := s.saveManifest(name, m); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "ok — diff #%d stored\n", next)
}

func (s *Server) diffPath(name string, n int) string {
	return filepath.Join(s.dataDir, name, fmt.Sprintf("%04d.syncdiff", n))
}

func (s *Server) manifestPath(name string) string {
	return filepath.Join(s.dataDir, name, "manifest.json")
}

func (s *Server) loadManifest(name string) (manifest, error) {
	data, err := os.ReadFile(s.manifestPath(name))
	if os.IsNotExist(err) {
		return manifest{}, nil
	}
	if err != nil {
		return manifest{}, err
	}
	var m manifest
	return m, json.Unmarshal(data, &m)
}

func (s *Server) saveManifest(name string, m manifest) error {
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.manifestPath(name), data, 0644)
}

func validName(name string) bool {
	return name != "" &&
		!strings.Contains(name, "/") &&
		!strings.Contains(name, "\\") &&
		!strings.Contains(name, ".")
}
