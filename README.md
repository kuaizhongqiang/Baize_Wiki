# Baize Wiki (白泽维基)

> 面向 AI Agent 的 Wiki 生成与使用工具

[![Go Version](https://img.shields.io/badge/Go-1.24+-blue)](https://go.dev/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Status](https://img.shields.io/badge/status-v0.1.0--alpha-orange)]()
[![Release](https://img.shields.io/github/v/release/kuaizhongqiang/Baize_Wiki)]()

**Baize Wiki** 是一个专为 AI Agent 设计的 Wiki 工具。它能扫描指定路径下的所有文档，自动解析内容，生成带全文搜索、语义搜索和交叉引用的结构化 Wiki，并通过 CLI 和 MCP Server 两种方式供 AI Agent 使用。

### 核心特性

- **能扫尽扫** — 自动探测并跳过二进制文件，支持 `.md`/`.mdx` 及 `--scan-all` 全文本扫描
- **三级输出** — `--level 1/2/3` 控制输出结构，从平面文件到深度目录
- **全文搜索** — 基于 bleve 的关键词检索，支持标签筛选
- **语义搜索** — 向量化 Embedding + BM25 混合检索（Hybrid Search）
- **交叉引用** — `[[wiki-link]]` 自动解析 + 反向链接（Backlinks）
- **代码注释提取** — 支持 20+ 编程语言的文档注释提取
- **AI Agent 优先** — 提供 CLI 和 MCP Server 两种集成方式
- **单二进制分发** — 零运行时依赖，即下即用

---

## 状态

**v0.1.0-alpha** — 核心功能全部完成：

| Phase | 功能 | 状态 |
|-------|------|------|
| 1 | CLI MVP：扫描 → 解析 → 按 Level 1/2/3 生成 Wiki | ✅ |
| 2 | MCP Server：5 个工具（build/read/list/add/stats） | ✅ |
| 3 | 全文搜索（bleve）+ 代码注释提取 + `--scan-all` | ✅ |
| 4 | 向量语义搜索 + 混合检索（BM25 + Vector） | ✅ |
| 5 | `[[wiki-link]]` 交叉链接 + 反向链接 | ✅ |

---

## 快速开始

```bash
# 下载二进制 (从 Releases 页面)
# 或本地编译
go build -o bin/baize-wiki ./cmd/baize-wiki

# 初始化配置
baize-wiki init ./docs --name "我的 Wiki"

# 构建 Wiki (Level 2 结构化)
baize-wiki build ./docs --output ./wiki --level 2

# 搜索内容
baize-wiki search "关键词"
baize-wiki search "关键词" --semantic   # 语义搜索

# 启动 MCP Server（供 Claude Code 等 Agent 使用）
baize-wiki mcp ./wiki

# 浏览 Wiki 结构
baize-wiki info ./wiki --tree
```

### 输出级别

| Level | 结构 | 说明 |
|-------|------|------|
| `1` | 平面文件 | 所有内容聚合为 1-10 个 MD 文件，适合快速概览 |
| `2` | 结构化 | 按主题分 3-5 个子目录，适合按领域查阅 |
| `3` | 深度 | 完整目录树（最多 3 层），适合深度检索 |

---

## 项目结构

```
baize-wiki/
├── cmd/baize-wiki/          # CLI 入口（cobra）
├── internal/
│   ├── core/                # 核心域
│   │   ├── model/           # 数据结构（Page, Wiki, Config, Link）
│   │   ├── scanner/         # 文件扫描 + 二进制探测 + .baizeignore
│   │   ├── parser/          # Markdown/纯文本解析 + [[link]] 提取
│   │   ├── linker/          # 交叉链接解析 + Backlinks 计算
│   │   ├── generator/       # Wiki 输出生成（Level 1/2/3）
│   │   ├── index/           # 全文搜索引擎（bleve）
│   │   ├── vector/          # 向量语义搜索 + 混合检索
│   │   └── storage/         # 文件读写 + 元数据持久化
│   ├── app/                 # 应用层用例编排
│   ├── config/              # 配置加载（viper）
│   └── mcp/                 # MCP Server（JSON-RPC 2.0）
├── docs/                    # 设计文档
├── testdata/                # 测试数据
└── .baizeignore             # 扫描忽略规则
```

---

## 集成方式

### CLI 直接使用

```bash
# 构建 Wiki
baize-wiki build ./docs --level 2 --output ./wiki --json

# 全文搜索
baize-wiki search "数据模型" --limit 5

# 语义搜索（需先 build 生成向量索引）
baize-wiki search "数据库配置" --semantic

# 全文本扫描（含代码注释）
baize-wiki build --scan-all ./project --output ./wiki
```

### MCP Server

配置到 Claude Code / Cline 的 MCP settings 中：

```json
{
  "mcpServers": {
    "baize-wiki": {
      "command": "baize-wiki",
      "args": ["mcp", "/path/to/wiki"]
    }
  }
}
```

MCP 提供 6 个工具：`wiki_build`、`wiki_read`、`wiki_list`、`wiki_add`、`wiki_stats`、`wiki_search`。

---

## 搜索能力

Baize Wiki 提供三层递进搜索：

| 层级 | 方式 | 命令 |
|------|------|------|
| 全文搜索 | BM25 关键词匹配（bleve） | `baize-wiki search "关键词"` |
| 语义搜索 | Feature Hashing 向量 + 余弦相似度 | `baize-wiki search "关键词" --semantic` |
| 混合搜索 | BM25 + 向量加权融合 | `baize-wiki search "关键词" --semantic --hybrid-weight 0.5` |

---

## 设计文档

| 文档 | 说明 |
|------|------|
| [架构总览](docs/architecture.md) | 整体架构、Go 包结构、技术决策 |
| [数据模型](docs/data-model.md) | 核心结构体、配置格式、接口合约 |
| [CLI 规范](docs/cli-spec.md) | 命令定义、参数、退出码、JSON 输出格式 |
| [MCP 规范](docs/mcp-spec.md) | MCP 工具定义、JSON-RPC 协议细节 |
| [Phase 1 计划](docs/phase-1-plan.md) | CLI MVP 实施记录 |
| [Phase 2 计划](docs/phase-2-plan.md) | MCP Server 实施记录 |
| [Phase 3 计划](docs/phase-3-plan.md) | 全文搜索 + 代码注释实施记录 |
| [Phase 4 计划](docs/phase-4-plan.md) | 向量搜索 + 混合检索实施记录 |
| [Phase 5 计划](docs/phase-5-plan.md) | 交叉链接 + Backlinks 实施记录 |

---

## License

MIT License © 2026 kuaizhongqiang
