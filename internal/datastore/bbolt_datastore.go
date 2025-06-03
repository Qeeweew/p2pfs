package datastore

import (
	"context"
	"errors"
	"os"

	bbolt "go.etcd.io/bbolt"
)

// bboltDatastore implements Datastore using a bbolt backend.
type bboltDatastore struct {
	db *bbolt.DB
}

// NewBboltDatastore opens or creates a bbolt database at path with given file mode and options.
func NewBboltDatastore(path string, mode os.FileMode, options *bbolt.Options) (Datastore, error) {
	db, err := bbolt.Open(path, mode, options)
	if err != nil {
		return nil, err
	}
	return &bboltDatastore{db: db}, nil
}

func (b *bboltDatastore) Put(ctx context.Context, bucket string, key []byte, value []byte) error {
	return b.db.Update(func(tx *bbolt.Tx) error {
		bkt, err := tx.CreateBucketIfNotExists([]byte(bucket))
		if err != nil {
			return err
		}
		return bkt.Put(key, value)
	})
}

func (b *bboltDatastore) Get(ctx context.Context, bucket string, key []byte) ([]byte, error) {
	var val []byte
	err := b.db.View(func(tx *bbolt.Tx) error {
		bkt := tx.Bucket([]byte(bucket))
		if bkt == nil {
			return errors.New("bucket not found")
		}
		v := bkt.Get(key)
		if v == nil {
			return errors.New("key not found")
		}
		val = append([]byte{}, v...)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return val, nil
}

func (b *bboltDatastore) Delete(ctx context.Context, bucket string, key []byte) error {
	return b.db.Update(func(tx *bbolt.Tx) error {
		bkt := tx.Bucket([]byte(bucket))
		if bkt == nil {
			return errors.New("bucket not found")
		}
		return bkt.Delete(key)
	})
}

func (b *bboltDatastore) Close() error {
	return b.db.Close()
}
