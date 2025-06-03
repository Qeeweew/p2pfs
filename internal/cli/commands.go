package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/ipfs/go-cid"
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
	RootCmd.AddCommand(addCmd, getCmd, pinCmd, catCmd, lsCmd)
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
