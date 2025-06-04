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
  Datastore: key-value backend using bbolt (github.com/etcd-io/bbolt) for persistent block storage.

## Project Structure

```
.
├── cmd/
│   └── p2pfs/           # CLI entry point and commands
├── internal/
│   ├── blockstore/      # Blockstore interfaces and implementations
│   ├── datastore/       # Persistent key-value datastore abstraction
│   ├── dag/             # Merkle-DAG node creation and traversal
│   │   ├── importer/    # File import logic
│   │   └── exporter/    # File export logic
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

## Usage

Examples:

```bash
# Add a file to the P2P file system and print its CID
./p2pfs add <path-to-file>

# Retrieve and print the contents of a file by CID
./p2pfs cat <CID>

# List links in a DAG node by CID
./p2pfs ls <CID>

# Pin a block locally and announce it to the network
./p2pfs pin <CID>

# Demo P2P file sharing between two nodes
./p2pfs demo <path-to-file>
```

Web Frontend

To launch a simple web UI:

1. 构建并启动服务：
   ```bash
   go build -o p2pfs ./cmd/p2pfs
   ./p2pfs serve
   ```
2. 在浏览器中打开：  
   http://localhost:8080/

服务器会在 `web/` 目录下提供静态页面（`index.html`)，并通过 `/api/...` 路由调用 CLI 底层功能。

启动服务后，控制台会打印出本节点的 Peer ID 与多地址，例如：
```
Node ID: QmYourPeerID
Node address: /ip4/0.0.0.0/tcp/8080/p2p/QmYourPeerID
```
可将上述多地址复制到 “Connect Peer” 输入框，以连接到该节点。

Connect Peer

在 “Connect Peer” 输入框中输入目标节点的多地址，例如：  
/​ip4/127.0.0.1/tcp/8080/p2p/QmPeerID  
然后点击 “Connect” 按钮，以连接到该节点。

## TODO

- [x] Implement `datastore` with bbolt backend  
- [x] Develop `blockstore` for block encoding/decoding  
- [x] Create Merkle-DAG node generation and file import/export  
- [x] Integrate libp2p host, streams, and protocol handlers  
- [x] Implement Kademlia DHT for peer routing and provider records  
- [x] Build Bitswap engine for efficient block exchange  
- [x] Add CLI commands: `add`, `get`, `pin`, `cat`, `ls`  
- [ ] Write comprehensive unit and integration tests  
- [ ] Prepare user documentation and usage examples  
