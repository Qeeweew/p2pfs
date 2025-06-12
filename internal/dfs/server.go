package dfs

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

type Server struct {
	peers   []string
	dataDir string
}

func NewServer(peers []string) *Server {
	if err := os.MkdirAll("data", os.ModePerm); err != nil {
		log.Fatalf("failed to create data dir: %v", err)
	}
	return &Server{peers: peers, dataDir: "data"}
}

func (s *Server) ListHandler(w http.ResponseWriter, r *http.Request) {
	files, err := os.ReadDir(s.dataDir)
	if err != nil {
		http.Error(w, "failed to list files", http.StatusInternalServerError)
		return
	}
	names := []string{}
	for _, f := range files {
		if !f.IsDir() {
			names = append(names, f.Name())
		}
	}
	json.NewEncoder(w).Encode(names)
}

func (s *Server) UploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "invalid method", http.StatusMethodNotAllowed)
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "failed to read file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	dstPath := filepath.Join(s.dataDir, header.Filename)
	dst, err := os.Create(dstPath)
	if err != nil {
		http.Error(w, "failed to save file", http.StatusInternalServerError)
		return
	}
	defer dst.Close()
	if _, err := io.Copy(dst, file); err != nil {
		http.Error(w, "failed to write file", http.StatusInternalServerError)
		return
	}

	// replicate to peers
	for _, peer := range s.peers {
		go replicateToPeer(peer, header.Filename, dstPath)
	}
	w.WriteHeader(http.StatusCreated)
}

func (s *Server) DownloadHandler(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("file")
	http.ServeFile(w, r, filepath.Join(s.dataDir, name))
}

func (s *Server) ReplicateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "invalid method", http.StatusMethodNotAllowed)
		return
	}
	name := r.URL.Query().Get("file")
	dstPath := filepath.Join(s.dataDir, name)
	dst, err := os.Create(dstPath)
	if err != nil {
		http.Error(w, "failed to create file", http.StatusInternalServerError)
		return
	}
	defer dst.Close()
	if _, err := io.Copy(dst, r.Body); err != nil {
		http.Error(w, "failed to write file", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func replicateToPeer(peer, name, path string) {
	client := NewClient(peer)
	if err := client.Replicate(name, path); err != nil {
		log.Printf("replicate to %s failed: %v", peer, err)
	}
}
