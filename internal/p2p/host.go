package p2p

import (
	"context"
	"fmt"

	libp2p "github.com/libp2p/go-libp2p"
	corehost "github.com/libp2p/go-libp2p-core/host"
)

// NewHost initializes a libp2p host listening on the given TCP port.
func NewHost(ctx context.Context, listenPort int) (corehost.Host, error) {
	addr := fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", listenPort)
	h, err := libp2p.New(
		libp2p.ListenAddrStrings(addr),
	)
	if err != nil {
		return nil, err
	}
	return h, nil
}
