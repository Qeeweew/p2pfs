# p2pfs

p2pfs 是一个基于 Go 语言实现的 P2P 文件系统，设计灵感来源于 IPFS。项目通过以下核心模块展示了分布式内容寻址与数据交换的原理：

## 核心原理

- 内容寻址（Content Addressing）  
  每个数据块通过多哈希（multihash）计算生成唯一的内容 ID（CID），保证数据不可篡改且可去重。

- Merkle-DAG  
  文件和目录表示为以 CID 连接的有向无环图（DAG），支持高效的增量更新和并行传输。

- libp2p 网络  
  模块化 P2P 网络栈负责节点连接、流复用和协议协商，提供点对点通信能力。

- 分布式哈希表（DHT）  
  基于 Kademlia 算法的 DHT 用于节点发现和提供者记录，实现内容的高效定位与路由。

- Bitswap 协议  
  一种块交换协议，节点可请求其他对等节点提供指定 CID 的数据块，并支持请求合并与并行下载。

- 存储层  
  使用 bbolt（github.com/etcd-io/bbolt）作为持久化键值存储，实现持久化 Blockstore 与内存或磁盘存储接口。

## 项目结构

```
.
├── cmd/p2pfs         CLI 入口及命令定义
├── internal/
│   ├── blockstore    块存储接口与实现
│   ├── datastore     bbolt 持久化存储抽象
│   ├── dag           Merkle-DAG 节点创建与遍历
│   ├── p2p           libp2p 主机与协议处理
│   ├── routing       DHT 路由与内容发现
│   ├── bitswap       Bitswap 块交换协议引擎
│   └── cli           命令行工具实现
├── pkg               公共可复用包
└── web               静态 Web 界面（index.html）
```

## 快速开始

1. 安装 Go (>=1.20)。  
2. 克隆仓库并编译：  
   ```bash
   go build ./cmd/p2pfs
   ```  
3. 查看帮助：  
   ```bash
   ./p2pfs --help
   ```

## 使用示例

```bash
# 添加文件并打印 CID
./p2pfs add <文件路径>

# 根据 CID 导出文件内容
./p2pfs cat <CID>

# 列出 DAG 节点中的链接
./p2pfs ls <CID>

# 本地固定并广播块
./p2pfs pin <CID>

# 节点间 P2P 文件共享演示
./p2pfs demo <文件路径>
```

## Web 前端

1. 构建并启动服务：  
   ```bash
   go build -o p2pfs ./cmd/p2pfs
   ./p2pfs serve
   ```  
2. 打开浏览器访问：  
   http://localhost:8080/  

前端界面可发起 /api 路由请求，与底层 CLI 功能交互，实现文件上传、下载及节点管理。
