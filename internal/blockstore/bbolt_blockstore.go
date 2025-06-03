package blockstore

import (
    "context"

    blockformat "github.com/ipfs/go-block-format"
    "github.com/ipfs/go-cid"

    "p2pfs/internal/datastore"
)

const bucketName = "blocks"

// BboltBlockstore persists blocks in a bbolt-backed Datastore.
type BboltBlockstore struct {
    ds datastore.Datastore
}

// NewBboltBlockstore wraps a Datastore in a Blockstore.
func NewBboltBlockstore(ds datastore.Datastore) *BboltBlockstore {
    return &BboltBlockstore{ds: ds}
}

func (b *BboltBlockstore) Put(ctx context.Context, block blockformat.Block) error {
    key := block.Cid().Bytes()
    data := block.RawData()
    return b.ds.Put(ctx, bucketName, key, data)
}

func (b *BboltBlockstore) Get(ctx context.Context, id cid.Cid) (blockformat.Block, error) {
    data, err := b.ds.Get(ctx, bucketName, id.Bytes())
    if err != nil {
        return nil, err
    }
    blk := blockformat.NewBlock(data)
    return blk, nil
}

func (b *BboltBlockstore) Delete(ctx context.Context, id cid.Cid) error {
    return b.ds.Delete(ctx, bucketName, id.Bytes())
}

func (b *BboltBlockstore) Has(ctx context.Context, id cid.Cid) (bool, error) {
    data, err := b.ds.Get(ctx, bucketName, id.Bytes())
    if err != nil {
        if err.Error() == "key not found" {
            return false, nil
        }
        return false, err
    }
    return len(data) > 0, nil
}

func (b *BboltBlockstore) Close() error {
    return b.ds.Close()
}
