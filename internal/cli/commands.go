package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ipfs/go-cid"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/spf13/cobra"

	"p2pfs/internal/bitswap"
	"p2pfs/internal/blockstore"
	"p2pfs/internal/dag"
	"p2pfs/internal/dag/exporter"
	"p2pfs/internal/dag/importer"
	"p2pfs/internal/datastore"
	"p2pfs/internal/p2p"
	"p2pfs/internal/routing"
)

// RootCmd is the base command for the p2pfs CLI.
var RootCmd = &cobra.Command{
	Use:   "p2pfs",
	Short: "p2pfs CLI",
}

func init() {
	RootCmd.AddCommand(addCmd, getCmd, pinCmd, catCmd, lsCmd, demoCmd)
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
		fmt.Println("Node A ID:", hostA.ID().Pretty())
		for _, addr := range hostA.Addrs() {
			fmt.Printf("Node A address: %s/p2p/%s\n", addr.String(), hostA.ID().Pretty())
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
		fmt.Println("Node B ID:", hostB.ID().Pretty())
		for _, addr := range hostB.Addrs() {
			fmt.Printf("Node B address: %s/p2p/%s\n", addr.String(), hostB.ID().Pretty())
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
