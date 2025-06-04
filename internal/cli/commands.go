package cli

import (
	"context"
	"log"
	"fmt"
	"net/http"
	"encoding/json"
	"io"
	"os"
	"path/filepath"

	"github.com/ipfs/go-cid"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/spf13/cobra"

	"p2pfs/internal/bitswap"
	"p2pfs/internal/blockstore"
	"p2pfs/internal/dag"
	"github.com/ipfs/go-merkledag"
	"p2pfs/internal/dag/exporter"
	"p2pfs/internal/dag/importer"
	"p2pfs/internal/datastore"
	"p2pfs/internal/p2p"
	"p2pfs/internal/routing"
	"time"
	"github.com/multiformats/go-multiaddr"
)

// RootCmd is the base command for the p2pfs CLI.
var RootCmd = &cobra.Command{
	Use:   "p2pfs",
	Short: "p2pfs CLI",
}

func init() {
	RootCmd.AddCommand(addCmd, getCmd, pinCmd, catCmd, lsCmd, demoCmd, serveCmd)
	serveCmd.Flags().IntVarP(&servePort, "port", "p", 8080, "port to serve on")
}

var addCmd = &cobra.Command{
	Use:   "add [file]",
	Short: "Add a file to the P2P file system",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		dbPath := "p2pfs.db"
		ds, err := datastore.NewBboltDatastore(dbPath, 0600, nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to open datastore: %v\n", err)
			os.Exit(1)
		}
		defer ds.Close()
		bs := blockstore.NewBboltBlockstore(ds)
		defer bs.Close()

		cidKey, err := importer.ImportFile(context.Background(), args[0], bs)
		if err != nil {
			fmt.Fprintf(os.Stderr, "add failed: %v\n", err)
			os.Exit(1)
		}
		cmd.Println(cidKey.String())
	},
}

var servePort int

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start HTTP web interface",
	Run: func(cmd *cobra.Command, args []string) {
		log.Printf("Starting web server on :%d", servePort)
		// initialize datastore and blockstore
		dbPath := "p2pfs.db"
		ds, err := datastore.NewBboltDatastore(dbPath, 0600, nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "datastore error: %v\n", err)
			os.Exit(1)
		}
		defer ds.Close()
		bs := blockstore.NewBboltBlockstore(ds)
		defer bs.Close()

		// initialize P2P host, DHT, and Bitswap engine
		host, err := p2p.NewHost(context.Background(), 0)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to create host: %v\n", err)
			os.Exit(1)
		}
		dhtEngine, err := routing.NewKademliaDHT(context.Background(), host)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to create dht: %v\n", err)
			os.Exit(1)
		}
		if err := dhtEngine.Bootstrap(context.Background()); err != nil {
			fmt.Fprintf(os.Stderr, "dht bootstrap warning: %v\n", err)
		}
		bsEngine := bitswap.NewBitswap(host, dhtEngine, bs)
		// print this node's Peer ID and multiaddrs for P2P connections
		log.Printf("Node ID: %s", host.ID().String())
		for _, addr := range host.Addrs() {
			log.Printf("Node address: %s/p2p/%s", addr.String(), host.ID().String())
		}

		mux := http.NewServeMux()

		metadataBucket := "metadata"
		sharedFiles := make(map[string]string)
		if data, err := ds.Get(context.Background(), metadataBucket, []byte("shared_meta")); err == nil {
		    json.Unmarshal(data, &sharedFiles)
		}
		mux.Handle("/", http.FileServer(http.Dir("web")))

		mux.HandleFunc("/api/add", func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
				return
			}
			file, fh, err := r.FormFile("file")
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			defer file.Close()
			tmp, err := os.CreateTemp("", "upload-*")
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			defer os.Remove(tmp.Name())
			defer tmp.Close()
			if _, err := io.Copy(tmp, file); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			cidKey, err := importer.ImportFile(context.Background(), tmp.Name(), bs)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			json.NewEncoder(w).Encode(map[string]string{"cid": cidKey.String()})
			// record uploaded file metadata and persist
			sharedFiles[fh.Filename] = cidKey.String()
			if metaBytes, err := json.Marshal(sharedFiles); err != nil {
			    log.Printf("failed to marshal shared metadata: %v", err)
			} else if err := ds.Put(context.Background(), metadataBucket, []byte("shared_meta"), metaBytes); err != nil {
			    log.Printf("failed to persist shared metadata: %v", err)
			}
		})

		mux.HandleFunc("/api/cat", func(w http.ResponseWriter, r *http.Request) {
			cidStr := r.URL.Query().Get("cid")
			cidKey, err := cid.Parse(cidStr)
			if err != nil {
				http.Error(w, "invalid cid", http.StatusBadRequest)
				return
			}
			blk, err := bs.Get(context.Background(), cidKey)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/octet-stream")
			w.Write(blk.RawData())

			// record fetched block into shared metadata
			if _, exists := sharedFiles[cidStr]; !exists {
				sharedFiles[cidStr] = cidStr
				if metaBytes, err := json.Marshal(sharedFiles); err != nil {
					log.Printf("failed to marshal shared metadata after fetch: %v", err)
				} else if err := ds.Put(context.Background(), metadataBucket, []byte("shared_meta"), metaBytes); err != nil {
					log.Printf("failed to persist shared metadata after fetch: %v", err)
				}
			}
		})

		mux.HandleFunc("/api/ls", func(w http.ResponseWriter, r *http.Request) {
			cidStr := r.URL.Query().Get("cid")
			cidKey, err := cid.Parse(cidStr)
			if err != nil {
				http.Error(w, "invalid cid", http.StatusBadRequest)
				return
			}
			blk, err := bs.Get(context.Background(), cidKey)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			nodeProto, err := merkledag.DecodeProtobuf(blk.RawData())
			var links []map[string]string
			if err != nil {
				links = []map[string]string{}
			} else {
				links = make([]map[string]string, len(nodeProto.Links()))
				for i, link := range nodeProto.Links() {
					links[i] = map[string]string{"name": link.Name, "cid": link.Cid.String()}
				}
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(links)
		})

		// P2P connect endpoint
		mux.HandleFunc("/api/connect", func(w http.ResponseWriter, r *http.Request) {
			addrStr := r.URL.Query().Get("addr")
			if addrStr == "" {
				http.Error(w, "addr query param required", http.StatusBadRequest)
				return
			}
			maddr, err := multiaddr.NewMultiaddr(addrStr)
			if err != nil {
				http.Error(w, "invalid multiaddr", http.StatusBadRequest)
				return
			}
			info, err := peer.AddrInfoFromP2pAddr(maddr)
			if err != nil {
				http.Error(w, "invalid peer addr", http.StatusBadRequest)
				return
			}
			ctxDial, cancel := context.WithTimeout(r.Context(), 10*time.Second)
			defer cancel()
			if err := host.Connect(ctxDial, *info); err != nil {
				http.Error(w, "connect failed: "+err.Error(), http.StatusInternalServerError)
				return
			}
			json.NewEncoder(w).Encode(map[string]string{"status": "connected"})
		})

		// P2P fetch endpoint via Bitswap
		mux.HandleFunc("/api/fetch", func(w http.ResponseWriter, r *http.Request) {
			cidStr := r.URL.Query().Get("cid")
			cidKey, err := cid.Parse(cidStr)
			if err != nil {
				http.Error(w, "invalid cid", http.StatusBadRequest)
				return
			}
			blk, err := bsEngine.GetBlock(context.Background(), cidKey)
			if err != nil {
				http.Error(w, "fetch error: "+err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/octet-stream")
			w.Write(blk.RawData())
		})

		// Shared files listing
		mux.HandleFunc("/api/shared", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(sharedFiles)
		})

		if err := http.ListenAndServe(fmt.Sprintf(":%d", servePort), mux); err != nil {
			fmt.Fprintf(os.Stderr, "server error: %v\n", err)
			os.Exit(1)
		}
	},
}

var getCmd = &cobra.Command{
	Use:   "get [cid] [output]",
	Short: "Retrieve a file by CID",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		dbPath := "p2pfs.db"
		ds, err := datastore.NewBboltDatastore(dbPath, 0600, nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to open datastore: %v\n", err)
			os.Exit(1)
		}
		defer ds.Close()
		bs := blockstore.NewBboltBlockstore(ds)
		defer bs.Close()

		cidKey, err := cid.Parse(args[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "invalid cid: %v\n", err)
			os.Exit(1)
		}
		if err := exporter.ExportFile(context.Background(), cidKey, bs, args[1]); err != nil {
			fmt.Fprintf(os.Stderr, "get failed: %v\n", err)
			os.Exit(1)
		}
	},
}

var pinCmd = &cobra.Command{
	Use:   "pin [cid]",
	Short: "Pin a block locally",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		dbPath := "p2pfs.db"
		ds, err := datastore.NewBboltDatastore(dbPath, 0600, nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to open datastore: %v\n", err)
			os.Exit(1)
		}
		defer ds.Close()
		bs := blockstore.NewBboltBlockstore(ds)
		defer bs.Close()

		cidKey, err := cid.Parse(args[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "invalid cid: %v\n", err)
			os.Exit(1)
		}

		host, err := p2p.NewHost(context.Background(), 0)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to create host: %v\n", err)
			os.Exit(1)
		}
		dht, err := routing.NewKademliaDHT(context.Background(), host)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to create dht: %v\n", err)
			os.Exit(1)
		}
		bsEngine := bitswap.NewBitswap(host, dht, bs)
		if err := bsEngine.ProvideBlock(context.Background(), cidKey); err != nil {
			fmt.Fprintf(os.Stderr, "pin failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("pinned", cidKey.String())
	},
}

var catCmd = &cobra.Command{
	Use:   "cat [cid]",
	Short: "Print block raw data",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		dbPath := "p2pfs.db"
		ds, err := datastore.NewBboltDatastore(dbPath, 0600, nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to open datastore: %v\n", err)
			os.Exit(1)
		}
		defer ds.Close()
		bs := blockstore.NewBboltBlockstore(ds)
		defer bs.Close()

		cidKey, err := cid.Parse(args[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "invalid cid: %v\n", err)
			os.Exit(1)
		}
		blk, err := bs.Get(context.Background(), cidKey)
		if err != nil {
			fmt.Fprintf(os.Stderr, "cat failed: %v\n", err)
			os.Exit(1)
		}
		cmd.Print(string(blk.RawData()))
	},
}

var lsCmd = &cobra.Command{
	Use:   "ls [cid]",
	Short: "List links in a DAG node",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		dbPath := "p2pfs.db"
		ds, err := datastore.NewBboltDatastore(dbPath, 0600, nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to open datastore: %v\n", err)
			os.Exit(1)
		}
		defer ds.Close()
		bs := blockstore.NewBboltBlockstore(ds)
		defer bs.Close()

		cidKey, err := cid.Parse(args[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "invalid cid: %v\n", err)
			os.Exit(1)
		}
		blk, err := bs.Get(context.Background(), cidKey)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ls failed: %v\n", err)
			os.Exit(1)
		}
		node, err := dag.DecodeNode(blk.RawData())
		if err != nil {
			return
		}
		for _, link := range node.Links() {
			cmd.Printf("%s\t%s\n", link.Name, link.Cid)
		}
	},
}

var demoCmd = &cobra.Command{
	Use:   "demo [file]",
	Short: "Demo P2P file sharing between two nodes",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		// setup node A
		dirA, _ := os.MkdirTemp("", "nodeA")
		dsA, err := datastore.NewBboltDatastore(filepath.Join(dirA, "dbA.db"), 0600, nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "nodeA datastore error: %v\n", err)
			os.Exit(1)
		}
		defer dsA.Close()
		bsA := blockstore.NewBboltBlockstore(dsA)
		defer bsA.Close()
		hostA, err := p2p.NewHost(ctx, 0)
		if err != nil {
			fmt.Fprintf(os.Stderr, "nodeA host error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Node A ID:", hostA.ID().String())
		for _, addr := range hostA.Addrs() {
			fmt.Printf("Node A address: %s/p2p/%s\n", addr.String(), hostA.ID().String())
		}
		dhtA, err := routing.NewKademliaDHT(ctx, hostA)
		if err != nil {
			fmt.Fprintf(os.Stderr, "nodeA dht error: %v\n", err)
			os.Exit(1)
		}
		if err := dhtA.Bootstrap(ctx); err != nil {
			fmt.Fprintf(os.Stderr, "nodeA bootstrap error: %v\n", err)
			os.Exit(1)
		}
		bsEngA := bitswap.NewBitswap(hostA, dhtA, bsA)

		// setup node B
		dirB, _ := os.MkdirTemp("", "nodeB")
		dsB, err := datastore.NewBboltDatastore(filepath.Join(dirB, "dbB.db"), 0600, nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "nodeB datastore error: %v\n", err)
			os.Exit(1)
		}
		defer dsB.Close()
		bsB := blockstore.NewBboltBlockstore(dsB)
		defer bsB.Close()
		hostB, err := p2p.NewHost(ctx, 0)
		if err != nil {
			fmt.Fprintf(os.Stderr, "nodeB host error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Node B ID:", hostB.ID().String())
		for _, addr := range hostB.Addrs() {
			fmt.Printf("Node B address: %s/p2p/%s\n", addr.String(), hostB.ID().String())
		}
		dhtB, err := routing.NewKademliaDHT(ctx, hostB)
		if err != nil {
			fmt.Fprintf(os.Stderr, "nodeB dht error: %v\n", err)
			os.Exit(1)
		}
		if err := dhtB.Bootstrap(ctx); err != nil {
			fmt.Fprintf(os.Stderr, "nodeB bootstrap error: %v\n", err)
			os.Exit(1)
		}
		bsEngB := bitswap.NewBitswap(hostB, dhtB, bsB)

		// connect B → A
		if err := hostB.Connect(ctx, peer.AddrInfo{ID: hostA.ID(), Addrs: hostA.Addrs()}); err != nil {
			fmt.Fprintf(os.Stderr, "connect error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Node B connected to Node A")

		// import & provide on A
		cidKey, err := importer.ImportFile(ctx, args[0], bsA)
		if err != nil {
			fmt.Fprintf(os.Stderr, "import error: %v\n", err)
			os.Exit(1)
		}
		if err := bsEngA.ProvideBlock(ctx, cidKey); err != nil {
			fmt.Fprintf(os.Stderr, "provide warning: %v\n", err)
		}

		// fetch on B
		blk, err := bsEngB.GetBlock(ctx, cidKey)
		if err != nil {
			fmt.Fprintf(os.Stderr, "get block error: %v\n", err)
			os.Exit(1)
		}
		outPath := filepath.Join(dirB, filepath.Base(args[0]))
		if err := os.WriteFile(outPath, blk.RawData(), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "write file error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Demo completed. Node B stored file at", outPath)
	},
}
