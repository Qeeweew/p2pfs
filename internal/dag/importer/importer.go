package importer

import (
	"context"
	"io"
	"os"

	"github.com/ipfs/go-cid"
	blockformat "github.com/ipfs/go-block-format"

	"p2pfs/internal/blockstore"
)

// ImportFile reads a file at path, stores it as a single DAG block, and returns its CID.
func ImportFile(ctx context.Context, path string, bs blockstore.Blockstore) (cid.Cid, error) {
	f, err := os.Open(path)
	if err != nil {
		return cid.Undef, err
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		return cid.Undef, err
	}

	node := merkledag.NewRawNode(data)
	raw := node.RawData()
	blk := blockformat.NewBlock(raw)

	if err := bs.Put(ctx, blk); err != nil {
		return cid.Undef, err
	}
	return node.Cid(), nil
}
