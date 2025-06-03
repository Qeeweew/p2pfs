package blockstore

import (
    "context"

    "github.com/ipfs/go-block-format"
    "github.com/ipfs/go-cid"
)

// Blockstore defines storing and retrieving IPLD blocks.
type Blockstore interface {
    Put(ctx context.Context, block blockformat.Block) error
    Get(ctx context.Context, id cid.Cid) (blockformat.Block, error)
    Delete(ctx context.Context, id cid.Cid) error
    Has(ctx context.Context, id cid.Cid) (bool, error)
    Close() error
}
