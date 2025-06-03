package blockstore

import (
	"context"
	"os"
	"testing"

	blockformat "github.com/ipfs/go-block-format"
	bolt "go.etcd.io/bbolt"
	"p2pfs/internal/datastore"
)

func TestBboltBlockstore_PutGetDeleteHas(t *testing.T) {
	// create a temp bbolt file
	f, err := os.CreateTemp("", "testdb-*.db")
	if err != nil {
		t.Fatal(err)
	}
	path := f.Name()
	f.Close()
	defer os.Remove(path)

	// open datastore
	ds, err := datastore.NewBboltDatastore(path, 0600, (*bolt.Options)(nil))
	if err != nil {
		t.Fatal(err)
	}
	defer ds.Close()

	// wrap blockstore
	bs := NewBboltBlockstore(ds)
	defer bs.Close()

	ctx := context.Background()

	// create a block
	data := []byte("hello blockstore")
	blk, err := blockformat.NewBlock(data)
	if err != nil {
		t.Fatal(err)
	}
	cid := blk.Cid()

	// ensure not present
	has, err := bs.Has(ctx, cid)
	if err != nil {
		t.Fatal(err)
	}
	if has {
		t.Fatal("expected block to be absent")
	}

	// put block
	if err := bs.Put(ctx, blk); err != nil {
		t.Fatal(err)
	}

	// ensure present
	has, err = bs.Has(ctx, cid)
	if err != nil {
		t.Fatal(err)
	}
	if !has {
		t.Fatal("expected block to be present")
	}

	// get block and verify data
	got, err := bs.Get(ctx, cid)
	if err != nil {
		t.Fatal(err)
	}
	if string(got.RawData()) != string(data) {
		t.Fatalf("block data mismatch: got %q, want %q", got.RawData(), data)
	}

	// delete block
	if err := bs.Delete(ctx, cid); err != nil {
		t.Fatal(err)
	}

	// ensure removed
	has, err = bs.Has(ctx, cid)
	if err != nil {
		t.Fatal(err)
	}
	if has {
		t.Fatal("expected block to be deleted")
	}
}
