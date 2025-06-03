package exporter

import (
	"context"
	"os"

	"github.com/ipfs/go-cid"
	blockformat "github.com/ipfs/go-block-format"
	merkledag "github.com/ipfs/go-merkledag"

	"p2pfs/internal/blockstore"
)

// ExportFile retrieves the block for root CID and writes its raw data to path.
func ExportFile(ctx context.Context, root cid.Cid, bs blockstore.Blockstore, path string) error {
	blk, err := bs.Get(ctx, root)
	if err != nil {
		return err
	}

	node, err := merkledag.DecodeProtobuf(blk.RawData())
	if err != nil {
		return err
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write(node.RawData())
	return err
}
