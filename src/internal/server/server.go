package server

import (
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

	path := filepath.Join(s.dataDir, name+".syncdiff")
	f, err := os.Open(path)
	if os.IsNotExist(err) {
		http.Error(w, "no diff for world", http.StatusNotFound)
		return
	}
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

	if err := os.MkdirAll(s.dataDir, 0755); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	path := filepath.Join(s.dataDir, name+".syncdiff")
	f, err := os.Create(path)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer f.Close()

	if _, err := io.Copy(f, r.Body); err != nil {
		http.Error(w, "write failed", http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "ok\n")
}

// validName rejects names with path separators or dots to prevent directory traversal.
func validName(name string) bool {
	return name != "" &&
		!strings.Contains(name, "/") &&
		!strings.Contains(name, "\\") &&
		!strings.Contains(name, ".")
}
