# Baize Wiki (白泽维基)

> 文档入库编目，为 AI 优化每次查询成本

[![Go Version](https://img.shields.io/badge/Go-1.25+-blue)](https://go.dev/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Status](https://img.shields.io/badge/status-v1.0.0--beta-green)](https://github.com/kuaizhongqiang/Baize_Wiki/releases)
[![Release](https://img.shields.io/github/v/release/kuaizhongqiang/Baize_Wiki)]()

**Baize Wiki** 是一个面向 AI Agent 的文档编目与知识库工具。源文件入库时经过三层处理——机械层（目录索引）、摘要层（LLM 摘要+关键词）、图谱层（概念分类+知识图谱）——让 AI 每次查询都以最少 token 获取最有价值的内容。

### 核心特性

- **入库编目** — 源文件入库时 LLM 自动理解内容，产出摘要、关键词、实体，而非全文搬运
- **三级编目** — `--catalog-level 0/2/3` 控制编目深度：无/摘要+关键词/概念分类+知识图谱
- **多后端支持** — `--catalog-backend local|remote`，支持本地首段提取或 Qwen3.5/DeepSeek
- **配置化** — baize.yaml + `--profile speed|balanced|local` + 环境变量全覆盖
- **三级输出** — `--level 1/2/3` 控制目录结构，从平面文件到深度目录
- **全文搜索** — 基于 bleve 的关键词检索，支持标签筛选
- **语义搜索** — 向量化 Embedding + BM25 混合检索（Hybrid Search）
- **增量构建** — `--incremental` 模式，只处理改动文件
- **AI Agent 优先** — CLI + MCP Server（7 工具 + Resources/Prompts）
- **单二进制分发** — 零运行时依赖，即下即用

---

## 状态

**v1.0.0-beta** — Alpha 五个 Phase + Beta 全部完成。

| 阶段 | 功能 | 状态 |
|:-----|:-----|:----:|
| Alpha 1-5 | CLI MVP + MCP + 搜索 + 向量 + 交叉链接 | ✅ |
| Beta D1 | 配置文件化（profile/backend/env） | ✅ |
| Beta D2 | Remote 编目（Qwen3.5 摘要+关键词+实体） | ✅ |
| Beta D3 | 概念目录（`--catalog-level 3`） | ✅ |
| Beta D4 | 输出模板（结构化 markdown） | ✅ |
| Beta D5 | Level 3 知识图谱 + wiki_graph MCP | ✅ |
| Beta G1 | Dockerfile | ✅ |
| Beta G2 | 真实 Embedding（bge-m3 + Local/Remote） | ✅ |
| Beta G3 | MCP Resources/Prompts 扩展 | ✅ |
| Beta G4 | 1000+ 文件性能测试数据 | ✅ |

详见 [Beta 路线图](docs/beta-roadmap.md)。

---

## 快速开始

```bash
# 下载二进制 (从 Releases 页面)
# 或本地编译
go build -o bin/baize-wiki ./cmd/baize-wiki

# 初始化配置
baize-wiki init ./source --name "我的 Wiki"

# Level 2: 结构化目录 + 首段提取编目（最快，无需 LLM）
baize-wiki build ./source --output ./wiki --level 2 --catalog-level 2

# Level 3: 完整编目 + 概念分类 + 知识图谱（需本地 LLM 或 API）
baize-wiki build ./source --output ./wiki --level 3 --catalog-level 3 --catalog-backend remote

# 搜索内容
baize-wiki search "关键词"
baize-wiki search "关键词" --semantic   # 语义搜索

# 启动 MCP Server（供 Claude Code 等 Agent 使用）
baize-wiki mcp ./wiki

# 浏览 Wiki 结构
baize-wiki info ./wiki --tree
```

### 编目级别

| Level | 产出 | 需要 LLM | 用途 |
|:-----:|:-----|:--------:|:-----|
| `0` | 无编目 | ❌ | 纯目录+原文 |
| `2` | 每页摘要+关键词+实体 | 可选 | AI 快速判断页面内容 |
| `3` | 摘要+关键词+实体+概念分类+图谱 | 可选 | 完整知识网络 |

### 后端选择

| 方式 | 示例 |
|:-----|:------|
| **CLI 参数** | `--catalog-backend remote --catalog-level 2` |
| **预设档位** | `--profile balanced`（catalog:远程API + 向量:本地） |
| **环境变量** | `BAIZE_CATALOG_ENDPOINT=http://localhost:1234/v1` |
| **配置文件** | baize.yaml 中 `catalog.backend: remote` |

### 编目输出样例（Qwen3.5 远程编目真实 C# 代码）

```markdown
> **FirstPersonController 是一个基于 Unity Input System 和 CharacterController 的
第一人称控制器，负责处理玩家移动、冲刺、跳跃、重力及旋转逻辑。它支持 Cinemachine
虚拟摄像机跟随，提供地面检测与超时机制以优化楼梯交互。**

`第一人称控制器` `玩家移动` `跳跃重力` `Cinemachine` `地面检测`
```

---

## 编目配置

### baize.yaml

```yaml
profile: balanced     # speed(全远程) | balanced(混合) | local(全本地)

catalog:
  level: 2            # 0 | 2 | 3
  backend: remote
  endpoint: http://localhost:1234/v1
  model: qwen/qwen3.5-9b

vector:
  mode: local
  endpoint: http://localhost:1234/v1
  model: text-embedding-baai-bge-m3-568m:2
```

### 环境变量

- `BAIZE_CATALOG_ENDPOINT` — 编目 LLM API 地址
- `BAIZE_CATALOG_MODEL` — 编目模型名
- `BAIZE_CATALOG_API_KEY` — 编目 API Key
- `BAIZE_CATALOG_BACKEND` — local | remote
- `BAIZE_VECTOR_ENDPOINT` — 向量 API 地址
- `BAIZE_VECTOR_MODEL` — 向量模型名
- `BAIZE_VECTOR_API_KEY` — 向量 API Key
- `BAIZE_VECTOR_MODE` — local | remote

---

## 项目结构

```
baize-wiki/
├── cmd/baize-wiki/          # CLI 入口（cobra）
├── internal/
│   ├── core/                # 核心域
│   │   ├── model/           # 数据结构（Page, Wiki, Config）
│   │   ├── scanner/         # 文件扫描 + 二进制探测
│   │   ├── parser/          # Markdown/纯文本解析
│   │   ├── linker/          # 交叉链接计算
│   │   ├── catalog/         # 编目流水线（摘要/实体/概念分类）
│   │   ├── graph/           # 知识图谱引擎
│   │   ├── generator/       # Wiki 输出生成
│   │   ├── index/           # 全文搜索引擎（bleve）
│   │   ├── vector/          # Embedding + 混合检索
│   │   └── storage/         # 文件读写 + 元数据持久化
│   ├── app/                 # 应用层用例编排
│   ├── config/              # 配置加载（viper + 环境变量）
│   └── mcp/                 # MCP Server（7 工具 + Resources/Prompts）
├── docs/                    # 设计文档
├── scripts/                 # 测试数据生成脚本
├── testdata/                # 测试数据
├── Dockerfile               # 多阶段构建
└── .env.example             # 环境变量配置样例
```

---

## 集成方式

### CLI 直接使用

```bash
# 构建 Wiki（Level 3 完整编目）
baize-wiki build ./project --level 3 --catalog-level 3 --catalog-backend remote --scan-all

# 全文搜索
baize-wiki search "数据模型" --limit 5

# 语义搜索（需先 build 生成向量索引）
baize-wiki search "数据库配置" --semantic

# 增量构建
baize-wiki build ./project --incremental --catalog-level 2
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

MCP 提供 7 个工具：`wiki_build`、`wiki_read`（支持 depth/section/max_tokens）、`wiki_list`、`wiki_add`、`wiki_stats`、`wiki_search`、`wiki_graph`。

---

## 搜索能力

| 层级 | 方式 | 命令 |
|:-----|:-----|:------|
| 全文搜索 | BM25 关键词匹配（bleve） | `baize-wiki search "关键词"` |
| 语义搜索 | bge-m3 向量 + 余弦相似度 | `baize-wiki search "关键词" --semantic` |
| 混合搜索 | BM25 + 向量加权融合 | `baize-wiki search "关键词" --semantic --hybrid-weight 0.5` |

搜索结果支持 `max_tokens` 控制响应长度，支持 `depth` 控制内容深度。

---

## 设计文档

| 文档 | 说明 |
|:-----|:------|
| [架构总览](docs/architecture.md) | 整体架构、Go 包结构、技术决策 |
| [数据模型](docs/data-model.md) | 核心结构体、配置格式、接口合约 |
| [CLI 规范](docs/cli-spec.md) | 命令定义、参数、退出码、JSON 输出格式 |
| [MCP 规范](docs/mcp-spec.md) | MCP 工具定义、JSON-RPC 协议细节 |
| [Beta 路线图](docs/beta-roadmap.md) | Beta 路线图与状态 |
| [D 方向详细设计](docs/directions/d-ai-enhancement.md) | AI 增强与 Token 优化设计 |
| [测试方案](docs/test-plan.md) | 全链路测试方案 |

---

## License

MIT License © 2026 kuaizhongqiang
