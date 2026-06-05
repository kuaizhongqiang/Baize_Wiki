# Baize Wiki — Phase 2 实施计划

> 版本: v1
> 状态: 待审计
> 焦点: MCP Server
> 预估工时: ~1 周（单人开发）

## 1. Phase 2 范围

**目标**：一个可运行的 MCP Server，让 AI Agent（Claude Code / Cline 等）能通过 stdio JSON-RPC 协议动态查询、浏览和编辑 Wiki 内容。

**包含**：
- ✅ MCP 协议层（JSON-RPC 消息定义、错误码）
- ✅ 传输层（stdio 模式，TCP 模式）
- ✅ 5 个 MCP Tools：wiki_build, wiki_read, wiki_list, wiki_add, wiki_stats
- ✅ CLI `mcp` 子命令
- ✅ E2E 测试（启动 Server → 发送请求 → 验证响应）
- ✅ Dockerfile

**不包含**（留给后续 Phase）：
- ❌ `wiki_search` 工具（→ Phase 3 全文索引）
- ❌ MCP Resources 实现（→ Phase 3+）
- ❌ MCP Prompts 模板（→ Phase 3+）
- ❌ 文件监听 watch 模式（→ Phase 2+）
- ❌ Docker 集成测试（当前无 Docker 环境，CI 中配置）

---

## 2. 架构

```
                          ┌──────────────────┐
                          │    AI Agent       │
                          │ (Claude Code...)  │
                          └────────┬─────────┘
                                   │ JSON-RPC 2.0 (stdin/stdout)
                                   ▼
┌──────────────────────────────────────────────────────┐
│                  MCP Server                           │
│  ┌────────────┐  ┌────────────┐  ┌────────────────┐  │
│  │ Transport   │  │ Protocol   │  │  Tool Router    │  │
│  │ (stdio/TCP) │──│ (JSON-RPC) │──│  (dispatch)     │  │
│  └────────────┘  └────────────┘  └───────┬────────┘  │
└──────────────────────────────────────────┼───────────┘
                                           │
              ┌────────────────────────────┼────────────┐
              │                            │            │
              ▼                            ▼            ▼
         app.RunBuild                app.RunInfo    storage.ReadMeta
              │                            │            │
              └────────── Phase 1 复用 ────┴────────────┘
```

### 分层设计

| 层 | 包 | 职责 |
|----|----|------|
| 传输层 | `internal/mcp/transport.go` | 从 stdin 读 JSON、写入 stdout、TCP listener |
| 协议层 | `internal/mcp/protocol.go` | JSON-RPC 2.0 请求/响应结构体、错误类型 |
| 核心 | `internal/mcp/server.go` | 主循环：接收请求 → 路由到工具 → 返回响应 |
| 工具 | `internal/mcp/tools.go` | 5 个工具的具体实现 |
| 应用层 | `internal/app/serve.go` | MCP Server 启动配置和生命周期管理 |

### 数据流

```
Agent 输入                    MCP Server

{"jsonrpc":"2.0",       ──▶  transport.Read()
 "method":"tools/list",       protocol.Parse(message)
 "id":1}                      server.dispatch("tools/list")
                         ◀──  tools.HandleList()
                         ──▶  protocol.Format(response)
                              transport.Write(response)
```

---

## 3. MCP 工具与 Phase 1 的依赖关系

| 工具 | 实现方式 | Phase 1 依赖 |
|------|---------|-------------|
| `wiki_build` | 调用 `app.RunBuild` 并格式化结果 | `app.RunBuild` ✅ |
| `wiki_read` | 读取 Wiki 目录下的 `.md` 文件，返回内容 | `os.ReadFile` |
| `wiki_list` | `os.ReadDir` 递归遍历目录，构造树结构 | `os.ReadDir` |
| `wiki_add` | 写入 `.md` 文件到 Wiki 目录（含原子写入） | `storage.Store` 或直接 `os.WriteFile` |
| `wiki_stats` | 读取 `.baize/meta.json` 并返回统计 | `storage.Store.ReadMeta` ✅ |

**5 个工具均不依赖新的核心域能力**，可直接基于 Phase 1 的存储层和文件系统实现。

---

## 4. 里程碑

| 里程碑 | 内容 | 预估 |
|--------|------|------|
| **M1: Protocol + Transport** | JSON-RPC 消息结构、stdio 读写、TCP listener | Day 1 |
| **M2: Server Core** | 主循环、工具注册机制、dispatch、错误处理 | Day 2-3 |
| **M3: Tools 实现** | 5 个工具 handler + `tools/list` + `tools/call` | Day 4-5 |
| **M4: CLI + Serve** | `mcp` 子命令、`app/serve.go` 生命周期 | Day 6 |
| **M5: E2E 测试** | 集成测试：启动 Server → JSON-RPC 交互 → 验证 | Day 6-7 |
| **M6: Dockerfile** | 多阶段构建、镜像发布 | Day 7 |

---

## 5. 详细任务分解

### M1: Protocol + Transport

```
文件清单:
├── internal/mcp/
│   ├── protocol.go       JSON-RPC 2.0 消息类型
│   └── transport.go      stdio / TCP 传输实现
```

**protocol.go**:

```go
// JSON-RPC 2.0 消息结构
type Request struct {
    JSONRPC string      `json:"jsonrpc"`
    ID      any         `json:"id"`
    Method  string      `json:"method"`
    Params  any         `json:"params,omitempty"`
}

type Response struct {
    JSONRPC string      `json:"jsonrpc"`
    ID      any         `json:"id"`
    Result  any         `json:"result,omitempty"`
    Error   *ErrorObj   `json:"error,omitempty"`
}

type ErrorObj struct {
    Code    int    `json:"code"`
    Message string `json:"message"`
    Data    any    `json:"data,omitempty"`
}
```

**transport.go**:

```go
// Transport 定义了 MCP 传输接口
type Transport interface {
    Read() ([]byte, error)
    Write([]byte) error
    Close() error
}

// StdioTransport 使用 stdin/stdout
// TCPTransport 使用 net.Listener (Phase 2 可选实现)
```

**验收条件：**
- [ ] JSON-RPC 消息可正确序列化/反序列化
- [ ] stdio 模式：从 stdin 读取 JSON、写入 stdout
- [ ] TCP 模式：`net.Listener` 监听端口

---

### M2: Server Core

```
文件清单:
├── internal/mcp/
│   └── server.go         MCP Server 主循环 + 工具注册
```

核心流程：

```go
type Server struct {
    transport Transport
    tools     map[string]ToolHandler
}

type ToolHandler func(ctx context.Context, params any) (any, *ErrorObj)

func (s *Server) Run(ctx context.Context) error {
    for {
        msg, err := s.transport.Read()
        // 解析 → 路由 → 执行 → 响应
    }
}
```

MCP 内置方法（不需要 tools 层）：
- `ping` → pong
- `tools/list` → 返回已注册的工具列表 (name, description, inputSchema)

**验收条件：**
- [ ] `ping` 请求正确返回
- [ ] `tools/list` 返回 5 个工具的 schema
- [ ] 未知 method 返回标准错误
- [ ] context 取消时优雅退出

---

### M3: Tools 实现

```
文件清单:
├── internal/mcp/
│   └── tools.go          5 个工具 handler
```

**wiki_build**

```go
// handlers.build → app.RunBuild → format as MCP content
```

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| source | string | 否 | 源路径 |
| output | string | 否 | 输出目录 |
| level | int | 否 | 1/2/3 |

**wiki_read**

```go
// handlers.read → os.ReadFile → return content
// 路径安全：校验 path 不包含 ../ 等穿越
```

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| path | string | 是 | 页面相对路径 |

**wiki_list**

```go
// handlers.list → os.ReadDir 递归 → tree structure
```

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| dir | string | 否 | 子目录路径 |
| depth | int | 否 | 递归深度（默认 1） |

**wiki_add**

```go
// handlers.add → os.WriteFile（原子写入）
// 路径安全校验（禁止 ../）
// overwrite=false 时检查文件已存在则报错
```

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| path | string | 是 | 页面相对路径 |
| content | string | 是 | Markdown 内容 |
| overwrite | bool | 否 | 是否覆盖（默认 false） |

**wiki_stats**

```go
// handlers.stats → storage.ReadMeta → count pages/dirs/tags
```

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| 无 | | | |

**验收条件：**
- [ ] `tools/call` + 参数 → 正确结果
- [ ] 参数缺失 → 标准错误
- [ ] 路径穿越（`../`）→ 拒绝
- [ ] wiki_read 不存在的页面 → 404 错误

---

### M4: CLI + Serve

```
文件清单:
├── cmd/baize-wiki/main.go    添加 mcp 子命令
├── internal/app/serve.go     MCP Server 启动用例
```

**CLI 命令：**

```bash
baize-wiki mcp [wiki-dir] [flags]

# 默认 stdio 模式
baize-wiki mcp ./wiki

# TCP 模式
baize-wiki mcp ./wiki --transport tcp --addr :8080
```

**app/serve.go：**

```go
func RunServe(ctx context.Context, wikiDir, transport string, addr string) error {
    // 验证 wikiDir 存在
    // 创建 transport (stdio / tcp)
    // 创建 server 并注册工具
    // 注册 signal handler（SIGINT/SIGTERM）
    // server.Run(ctx)
}
```

**验收条件：**
- [ ] `baize-wiki mcp --help` 显示帮助
- [ ] `baize-wiki mcp ./wiki` 启动 stdio 模式
- [ ] SIGINT 时优雅退出（关闭 transport，清理资源）

---

### M5: E2E 测试

```
文件清单:
├── internal/mcp/server_test.go    单元测试
├── internal/mcp/tools_test.go      工具测试
├── e2e_mcp_test.go                 集成测试
```

**单元测试：**

| 测试 | 说明 |
|------|------|
| `TestProtocolMarshal` | Request/Response 序列化正确 |
| `TestProtocolErrors` | 错误对象格式正确 |
| `TestTransportStdio` | stdio 读写正确 |
| `TestToolList` | tools/list 返回完整 schema |
| `TestToolBuild` | wiki_build 参数解析 + 调用 |
| `TestToolRead` | wiki_read 正常 + 不存在 |
| `TestToolListDir` | wiki_list 正常 + 深层 |
| `TestToolAdd` | wiki_add 新建 + 覆盖 |
| `TestToolStats` | wiki_stats 返回统计 |
| `TestPathSecurity` | 路径穿越被拒绝 |

**集成测试：**

```go
func TestMCPE2E(t *testing.T) {
    // 1. 准备测试 Wiki 目录
    // 2. 启动 MCP Server (goroutine)
    // 3. 通过 pipe 发送 JSON-RPC 请求
    // 4. 读取响应并验证
    // 5. 依次测试所有工具
}
```

---

### M6: Dockerfile

```dockerfile
# Build stage
FROM golang:1.24-alpine AS build
WORKDIR /src
COPY . .
RUN CGO_ENABLED=0 go build -o /bin/baize-wiki ./cmd/baize-wiki

# Runtime stage
FROM scratch
COPY --from=build /bin/baize-wiki /baize-wiki
ENTRYPOINT ["/baize-wiki"]
```

**注意**：Dockerfile 由于本地无 Docker 环境，在 CI 中验证。

---

## 6. 依赖关系

```
M1 (Protocol+Transport) ──▶ M2 (Server Core) ──▶ M3 (Tools)
                                                      │
                                                      ▼
                                               M4 (CLI+Serve)
                                                      │
                                                      ▼
                                               M5 (E2E Tests)
                                                      │
                                                      ▼
                                               M6 (Dockerfile)
```

- M1、M2、M3 可逐步推送，不需要一次全部完成再测试
- M5 需要 M4 就绪后才能跑 E2E（但工具单元测试可在 M3 完成后就开始）
- M6 与 M1-M5 无依赖，可随时并行

---

## 7. 代码规范（补充）

- MCP 工具名称用 `snake_case`，与 MCP 协议惯例一致
- JSON-RPC 请求/响应使用标准库 `encoding/json`
- 工具 handler 签名统一：`func(ctx context.Context, params json.RawMessage) (any, *ErrorObj)`
- 路径安全：所有用户输入的路径必须 clean + 校验是否越界
- stdio 模式下错误日志走 stderr，不影响 stdout 的数据流
- Transport 接口设计为可 mock，方便测试

---

## 8. 风险与缓解

| 风险 | 影响 | 缓解措施 |
|------|------|---------|
| JSON-RPC 协议兼容性 | Agent 无法连接 | 严格遵循 MCP 2025-03-26 规范，参考官方实现 |
| 路径穿越安全漏洞 | 读取/写入任意文件 | 所有路径做 Clean + 前缀检查，单元测试覆盖 |
| 并发请求乱序 | 响应与请求不匹配 | stdio 模式通常串行，必要时加 ID 校验 |
| TCP 模式运维复杂 | 端口冲突、防火墙 | Phase 2 TCP 作为备选，stdio 为默认推荐 |
| 本地无 Docker 环境 | Dockerfile 无法测试 | CI 中验证构建，手动测试在部署环境做 |
