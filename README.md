# p2pfs

A Go implementation of a P2P file system inspired by IPFS. This project demonstrates core IPFS design principles: content addressing, Merkle-DAG data structures, libp2p networking, DHT-based peer discovery, and Bitswap block exchange.

## Architecture Overview

- **Content Addressing**  
  Each piece of data is identified by a Content ID (CID), a multihash of the data block.

- **Merkle-DAG**  
  Files and directories are represented as directed acyclic graphs of blocks linked by CIDs.

- **libp2p Networking**  
  Modular P2P networking stack for peer connections, stream multiplexing, and protocol negotiation.

- **Distributed Hash Table (DHT)**  
  Kademlia-based DHT for peer routing, provider records, and content discovery.

- **Bitswap**  
  Block exchange protocol enabling peers to request and serve blocks.

- **Blockstore & Datastore**  
  Blockstore: in-memory or on-disk block storage interface.  
  Datastore: key-value backend (e.g., BoltDB) for persistent block storage.

## Project Structure

```
.
├── cmd/
│   └── p2pfs/           # CLI entry point and commands
├── internal/
│   ├── blockstore/      # Blockstore interfaces and implementations
│   ├── datastore/       # Persistent key-value datastore abstraction
│   ├── dag/             # Merkle-DAG node creation and traversal
│   ├── p2p/             # libp2p host setup and protocol handlers
│   ├── routing/         # DHT peer routing and discovery
│   ├── bitswap/         # Bitswap session management and message handling
│   └── cli/             # CLI command definitions and helpers
├── pkg/                 # Publicly consumable packages
└── docs/                # Design documents and specifications
```

## Getting Started

1. Install Go (>=1.20).  
2. Clone this repository and run:  
   ```bash
   go build ./cmd/p2pfs
   ```  
3. Initialize a node and view help:  
   ```bash
   ./p2pfs --help
   ```

## TODO

- [ ] Implement `datastore` with BoltDB backend  
- [ ] Develop `blockstore` for block encoding/decoding  
- [ ] Create Merkle-DAG node generation and file import/export  
- [ ] Integrate libp2p host, streams, and protocol handlers  
- [ ] Implement Kademlia DHT for peer routing and provider records  
- [ ] Build Bitswap engine for efficient block exchange  
- [ ] Add CLI commands: `add`, `get`, `pin`, `cat`, `ls`  
- [ ] Write comprehensive unit and integration tests  
- [ ] Prepare user documentation and usage examples  
