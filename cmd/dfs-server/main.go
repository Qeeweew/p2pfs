package main

import (
	"flag"
	"log"
	"net/http"
	"strings"

	"p2pfs/internal/dfs"
)

func main() {
	port := flag.String("port", "9090", "port to serve on")
	peers := flag.String("peers", "", "comma-separated list of peer addresses (http://host:port)")
	flag.Parse()

	peerList := []string{}
	if *peers != "" {
		peerList = strings.Split(*peers, ",")
	}

	srv := dfs.NewServer(peerList)
	http.Handle("/", http.FileServer(http.Dir("web/dfs")))
	http.HandleFunc("/api/files", srv.ListHandler)
	http.HandleFunc("/api/upload", srv.UploadHandler)
	http.HandleFunc("/api/download", srv.DownloadHandler)
	http.HandleFunc("/api/replicate", srv.ReplicateHandler)

	addr := ":" + *port
	log.Printf("DFS server listening on %s", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
