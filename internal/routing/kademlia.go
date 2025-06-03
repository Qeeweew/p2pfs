package routing

import (
    "context"

    kaddht "github.com/libp2p/go-libp2p-kad-dht"
    "github.com/ipfs/go-cid"
    "github.com/libp2p/go-libp2p-core/host"
    "github.com/libp2p/go-libp2p-core/peer"
)

// KademliaDHT wraps a libp2p Kademlia DHT instance.
type KademliaDHT struct {
    dht *kaddht.IpfsDHT
}

// NewKademliaDHT constructs and bootstraps a Kademlia DHT.
func NewKademliaDHT(ctx context.Context, h host.Host) (*KademliaDHT, error) {
    d, err := kaddht.New(ctx, h)
    if err != nil {
        return nil, err
    }
    // bootstrapping must be done by the caller (e.g. Connect to peers)
    return &KademliaDHT{dht: d}, nil
}

// Provide announces to the DHT that we can serve the given CID.
func (k *KademliaDHT) Provide(ctx context.Context, c cid.Cid, announce bool) error {
    return k.dht.Provide(ctx, c, announce)
}

// FindProviders finds up to `max` providers for the CID.
func (k *KademliaDHT) FindProviders(ctx context.Context, c cid.Cid, max int) ([]peer.AddrInfo, error) {
    ch := k.dht.FindProvidersAsync(ctx, c, max)
    var infos []peer.AddrInfo
    for pi := range ch {
        infos = append(infos, pi)
    }
    return infos, nil
}

// Bootstrap triggers the DHT bootstrap process.
func (k *KademliaDHT) Bootstrap(ctx context.Context) error {
    return k.dht.Bootstrap(ctx)
}
