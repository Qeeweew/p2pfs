package bitswap

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"

	"github.com/ipfs/go-cid"
	blockformat "github.com/ipfs/go-block-format"
	corehost "github.com/libp2p/go-libp2p/core/host"
	cnetwork "github.com/libp2p/go-libp2p/core/network"

	"p2pfs/internal/blockstore"
	"p2pfs/internal/routing"

	"github.com/libp2p/go-libp2p/core/peer"
)

const BitswapProtocol = "/p2pfs/bitswap/1.0.0"

// Bitswap implements a simple block exchange protocol.
type Bitswap struct {
	host corehost.Host
	dht  *routing.KademliaDHT
	bs   blockstore.Blockstore
}

// NewBitswap returns a new Bitswap instance and sets the stream handler.
func NewBitswap(host corehost.Host, dht *routing.KademliaDHT, bs blockstore.Blockstore) *Bitswap {
	b := &Bitswap{host: host, dht: dht, bs: bs}
	host.SetStreamHandler(BitswapProtocol, b.handleStream)
	return b
}

// GetBlock retrieves a block by CID, either from local store or peers.
func (b *Bitswap) GetBlock(ctx context.Context, cidKey cid.Cid) (blockformat.Block, error) {
	// Try local store
	has, err := b.bs.Has(ctx, cidKey)
	if err != nil {
		return nil, err
	}
	if has {
		return b.bs.Get(ctx, cidKey)
	}
	// Find providers
	providers, err := b.dht.FindProviders(ctx, cidKey, 10)
	if err != nil {
		return nil, err
	}
	// fallback to directly connected peers if no providers found via DHT
	if len(providers) == 0 {
		for _, pid := range b.host.Peerstore().Peers() {
			providers = append(providers, peer.AddrInfo{ID: pid, Addrs: b.host.Peerstore().Addrs(pid)})
		}
	}
	// Query each provider
	for _, pi := range providers {
		if err := b.host.Connect(ctx, pi); err != nil {
			continue
		}
		s, err := b.host.NewStream(ctx, pi.ID, BitswapProtocol)
		if err != nil {
			continue
		}
		// send request
		req := struct{ Cid string }{Cid: cidKey.String()}
		w := bufio.NewWriter(s)
		if err := json.NewEncoder(w).Encode(&req); err != nil {
			s.Close()
			continue
		}
		w.Flush()
		// read response
		r := bufio.NewReader(s)
		var resp struct {
			Data []byte
			Err  string
		}
		if err := json.NewDecoder(r).Decode(&resp); err != nil {
			s.Close()
			continue
		}
		s.Close()
		if resp.Err != "" {
			continue
		}
		blk := blockformat.NewBlock(resp.Data)
		_ = b.bs.Put(ctx, blk)
		return blk, nil
	}
	return nil, ErrNotFound
}

// ProvideBlock announces that we can provide this block.
func (b *Bitswap) ProvideBlock(ctx context.Context, cidKey cid.Cid) error {
	// attempt to announce block via DHT; ignore errors if no peers
	_ = b.dht.Provide(ctx, cidKey, true)
	return nil
}

// ErrNotFound is returned when a block cannot be retrieved.
var ErrNotFound = errors.New("block not found")

// handleStream services incoming Bitswap requests.
func (b *Bitswap) handleStream(s cnetwork.Stream) {
	defer s.Close()
	r := bufio.NewReader(s)
	var req struct{ Cid string }
	if err := json.NewDecoder(r).Decode(&req); err != nil {
		return
	}
	id, err := cid.Parse(req.Cid)
	if err != nil {
		return
	}
	data, err := func() ([]byte, error) {
		has, err := b.bs.Has(context.Background(), id)
		if err != nil || !has {
			return nil, err
		}
		blk, err := b.bs.Get(context.Background(), id)
		if err != nil {
			return nil, err
		}
		return blk.RawData(), nil
	}()
	var resp struct {
		Data []byte
		Err  string
	}
	if err != nil {
		resp.Err = err.Error()
	} else {
		resp.Data = data
	}
	w := bufio.NewWriter(s)
	json.NewEncoder(w).Encode(&resp)
	w.Flush()
}
