package datastore

import "context"

// Datastore defines a simple key-value store interface.
type Datastore interface {
	Put(ctx context.Context, bucket string, key []byte, value []byte) error
	Get(ctx context.Context, bucket string, key []byte) ([]byte, error)
	Delete(ctx context.Context, bucket string, key []byte) error
	Close() error
}
