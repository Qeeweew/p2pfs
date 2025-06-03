package dag

import (
	"github.com/ipfs/go-cid"
	merkledag "github.com/ipfs/go-merkledag"
)

// CreateNode wraps data in a Merkle-DAG raw node and returns the node and its CID.
func CreateNode(data []byte) (*merkledag.RawNode, cid.Cid) {
	node := merkledag.NewRawNode(data)
	return node, node.Cid()
}

// DecodeNode parses a serialized DAG node and returns a RawNode.
func DecodeNode(raw []byte) (*merkledag.ProtoNode, error) {
	node, err := merkledag.DecodeProtobuf(raw)
	if err != nil {
		return nil, err
	}
	return node, nil
}
