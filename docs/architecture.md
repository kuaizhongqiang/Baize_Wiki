# Baize Wiki — 架构设计文档

> 版本: v2
> 状态: 已完成 v0.1.0-alpha，规划 Beta 路线图

## 1. 产品定位

**Baize Wiki** 是一个面向 AI Agent 的 Wiki 生成与使用工具。

- 直接用户始终是 **AI Agent**，而非人类
- 提供 **CLI**、**MCP Server**、**容器化** 三种集成方式
- 以 **单二进制** 形式分发，零运行时依赖

### 设计原则

| 原则 | 说明 |
|------|------|
| **工具而非平台** | 不做 Web UI、不做实时协作、不做权限系统 |
| **文件优先** | Wiki 的存储和交换格式是文件系统 + Markdown |
| **渐进增强** | 核心功能不依赖外部服务，向量化是可选项 |
| **约定优于配置** | 合理的默认行为，配置只用来覆盖默认值 |

---

## 2. 整体架构

```
┌────────────────────────────────────────────────────┐
│                   集成层 (Interfaces)                │
│  ┌────────────────┐ ┌────────────┐ ┌────────────┐  │
│  │  CLI (cobra)   │ │ MCP Server  │ │  Docker    │  │
│  │  baize-wiki    │ │  (stdio)    │ │  Container │  │
│  └───────┬────────┘ └─────┬──────┘ └──────┬─────┘  │
└──────────┼────────────────┼───────────────┼────────┘
           │                │               │
           ▼                ▼               ▼
┌────────────────────────────────────────────────────┐
│                 应用层 (App Layer)                   │
│  ┌──────────┐ ┌──────────┐ ┌────────────────────┐  │
│  │  构建器   │ │ 检索器    │ │  Wiki 管理         │  │
│  │  Builder │ │  Searcher │ │  Manager (CRUD)    │  │
│  └─────┬────┘ └────┬─────┘ └─────────┬──────────┘  │
└────────┼───────────┼─────────────────┼─────────────┘
         │           │                 │
         ▼           ▼                 ▼
┌────────────────────────────────────────────────────┐
│                 核心域 (Core Domain)                │
│  ┌────────┐ ┌────────┐ ┌────────┐ ┌────────────┐  │
│  │ Scanner │ │ Parser │ │  Linker │ │ Generator  │  │
│  │ (扫描)  │ │ (解析)  │ │ (链接)  │ │ (生成)     │  │
│  └────────┘ └────────┘ └────────┘ └────────────┘  │
│  ┌────────┐ ┌────────┐ ┌────────┐ ┌────────────┐  │
│  │ Index   │ │ Storage │ │ Vector  │ │   Config   │  │
│  │ (索引)  │ │ (存储)  │ │ (向量)  │ │ (配置)     │  │
│  └────────┘ └────────┘ └────────┘ └────────────┘  │
└────────────────────────────────────────────────────┘
```

### 层间依赖规则

- **集成层** → 调用 **应用层**，不直接调用核心域
- **应用层** → 编排 **核心域** 的多个组件完成业务用例
- **核心域** → 只依赖 Go 标准库和少数精选第三方包，不依赖任何集成
- 箭头方向即依赖方向：上层依赖下层，下层不感知上层

---

## 3. Go 包结构

```
baize-wiki/
├── cmd/
│   └── baize-wiki/
│       └── main.go              # 入口：flag 解析 >> 调用 app layer
│
├── internal/
│   ├── app/                     # 应用层：用例编排
│   │   ├── build.go             #   wiki build usecase
│   │   ├── search.go            #   wiki search usecase
│   │   ├── serve.go             #   mcp serve usecase
│   │   ├── init.go              #   baize.yaml 初始化
│   │   ├── info.go              #   wiki 信息查询
│   │   └── mcp.go               #   MCP 模式入口
│   │
│   ├── core/                    # 核心域：领域逻辑
│   │   ├── scanner/             #  文件扫描器
│   │   ├── parser/              #  文档解析器
│   │   ├── linker/              #  交叉链接计算
│   │   ├── generator/           #  Wiki 生成器
│   │   ├── index/               #  全文搜索引擎 (bleve)
│   │   ├── storage/             #  持久化
│   │   ├── model/               #  领域模型 (纯数据结构)
│   │   └── vector/              #  向量存储 + 混合检索
│   │
│   ├── mcp/                     # MCP 协议实现
│   │   ├── server.go            #    MCP Server 主循环
│   │   ├── tools.go             #    工具定义 & 路由 (6 工具)
│   │   ├── protocol.go          #    JSON-RPC 消息类型
│   │   └── transport.go         #    stdio / TCP 传输
│   │
│   └── config/                  # 配置加载
│       └── config.go            #    读取 `baize.yaml`
│
├── pkg/
│   └── baize/                   # 公开库 API（供 Go 程序嵌入）
│
├── go.mod / go.sum
├── Makefile
├── .goreleaser.yaml
├── .github/workflows/ci.yml
├── .baizeignore
└── baize.yaml
```

---

## 4. 输出级别 (Level) 系统

Baize Wiki 的输出结构分为三个复杂度级别，由 `--level` 参数控制。

| Level | 结构 | 目录数 | 页面数 | 适用场景 |
|-------|------|--------|--------|---------|
| **1 (Flat)** | 单层目录，高聚合 | 1 | 1-10 | Agent 快速了解全貌，一次性读完 |
| **2 (Structured)** | 双层目录，按主题分组 | 3-5 | 每目录 3-10 | Agent 按领域查阅 |
| **3 (Deep)** | 多层目录，细粒度 | 完整深度 (max 3) | 每节独立 | Agent 深度检索 |

### Level 1 — Flat (单层文件)

```
wiki/
├── _index.md                    # Wiki 概览
├── 架构设计.md                  # 按顶层分类聚合
├── 开发指南.md
├── API 参考.md
└── 运维手册.md
```

- 源文件的深层路径被拍平，顶层目录名作为文件标题前缀
- 同一分类下的多文件内容合并到一个页面

### Level 2 — Structured (结构化)

```
wiki/
├── 架构设计/
│   ├── _index.md
│   ├── 系统设计.md
│   └── 数据模型.md
├── 开发指南/
│   ├── _index.md
│   ├── 环境搭建.md
│   └── 测试指南.md
└── API 参考/
    ├── _index.md
    ├── REST API.md
    └── SDK 接口.md
```

- 按第一级路径分类建立子目录
- 每个文件保持独立

### Level 3 — Deep (深度)

```
wiki/
├── 架构设计/
│   ├── 系统设计/
│   │   ├── 模块划分.md
│   │   └── 通信协议.md
│   └── 数据模型/
│       ├── 核心实体.md
│       └── 关系定义.md
├── 开发指南/
│   └── ...
└── API 参考/
    └── ...
```

- 最大深度 3 层
- 源文件中的子节（sub-section）可独立成页

### 实现策略

Phase 1 采用最简单实现：**基于源目录结构**决定输出层级。

```
源文件路径: docs/guide/installation.md

Level 1: wiki/guide-installation.md     (拍平，前缀加父目录名)
Level 2: wiki/guide/installation.md     (一级子目录)
Level 3: wiki/guide/installation.md     (完整路径，与 Level 2 相同)
```

深度超过 level 限制的路径，向上合并到最近的有效层级。

### Level 1 合并算法 (Phase 1 实现)

Level 1 的核心是将多个源文件合并为少量的聚合页面。算法如下：

```
输入: Pages[]  (解析后的页面列表)
输出: MergedPages[]  (合并后的大页面列表)

步骤:
1. 分类 (Grouping)
   每个 Page 按以下优先级确定其分类 Key:
     a) Frontmatter 中的 category 字段 (如有)
     b) 源路径的第一级目录名 (如 docs/guide/a.md → "guide")
     c) 以上皆无 → "未分类" (Uncategorized)

2. 合并排序 (Ordering)
   同一分类下的 Pages 排序:
     ① Frontmatter.weight 从小到大
     ② 文件名按字典序（weight 相同时）

3. 内容合并 (Merging)
   对每个分类:
     a) 生成合并页面的 Title = 分类名 (如 "guide")
     b) 从各 Page 中收集 Tags 取并集
     c) Description = 第一个 Page 的 description (如有)
     d) Content = 按顺序拼接各 Page.Content，页间用 `---` 分隔
     e) 若合并后 Content > 50KB: 按 Page 拆分为多个文件
        (命名: guide.md, guide-1.md, guide-2.md ...)

4. 边界处理
   单分类下只有一个 Page → 直接输出，不合并
   所有 Pages 的 category 各不相同 → 每个独立输出
   空输入 → 不输出 (Generator 自行处理)
```

限制：Phase 1 不做智能去重或语义合并，仅为简单拼接。

---

## 5. 扫描策略

Baize Wiki 的扫描策略分为 **默认模式** 和 **全扫描模式**。

### 5.1 默认模式（Phase 1）

```
[源目录]
    │
    ▼
忽略规则匹配 (.baizeignore) ──▶ 跳过
    │
    ▼
扩展名过滤: 只保留 .md / .mdx 文件 ──▶ 其他跳过
    │
    ▼
读取前 512 字节 → 二进制探测
    │
    ├── 含 null 字节? ──▶ 跳过
    ├── 非 UTF-8 序列? ──▶ 跳过
    │
    ▼
Markdown 解析 (frontmatter + 内容)
```

默认只扫描 `**/*.md` 和 `**/*.mdx`，避免将 `go.mod`、`vendor/`、`node_modules/` 等非文档文件混入 Wiki。

### 5.2 全扫描模式 (`--scan-all`)

```
baize-wiki build --scan-all ./project
```

启用全扫描后，行为回归"能扫尽扫"：
- 所有非二进制文本文件都被纳入
- 代码文件支持注释提取
- 需配合 `.baizeignore` 排除 `vendor/`、`node_modules/` 等

### 5.3 二进制探测算法

Phase 1 默认模式只扫 .md/.mdx，二进制探测作为安全检查。`--scan-all` 启用后，其他文本文件类型识别规则逐步生效。

### 二进制探测算法

```go
// isBinary checks if content appears to be binary data
func isBinary(data []byte) bool {
    // 检查前 512 字节
    checkLen := min(len(data), 512)
    for i := 0; i < checkLen; i++ {
        if data[i] == 0 { // null byte → binary
            return true
        }
    }
    // 检查是否有效的 UTF-8
    return !utf8.Valid(data[:checkLen])
}
```

---

## 6. 核心数据流

### 6.1 Wiki 构建 (build)

```
                  读取 baize.yaml (可选)
                         │
[Source Dir] ──▶ Scanner ──▶ [File List] ──▶ Parser ──▶ [Pages]
                                                         │
                                                    Generator ──▶ [Wiki Dir]
                                                         │
                                                    [.baize/meta.json]
```

1. **Scanner** 递归扫描源目录，应用二进制探测跳过非文本文件，产出 `FileInfo` 列表
2. **Parser** 读取各文件，提取 frontmatter（如有），解析内容，产出 `Page` 列表
3. **Generator** 根据 `--level` 参数将 Pages 组织为目录树，生成各层级的 `_index.md`
4. **Storage** 写入 `meta.json` 到 `.baize/` 目录

> Phase 1 不含 Linker 和 Indexer，后续 Phase 逐步加入。

### 6.2 Wiki 查询 (search)

```
[Query] ──▶ Searcher ──▶ Index ──▶ [Results with snippets]
```
> 已实现。详情见 [index/index.go](../internal/core/index/index.go)。

### 6.3 MCP 服务 (serve)

```
[AI Agent] ◀──▶ stdio JSON-RPC ◀──▶ MCP Server ◀──▶ App Layer ◀──▶ Core Domain
```
> 已实现。详情见 [mcp/](../internal/mcp/)。

---

## 7. 项目演进路线

### v0.1.0-alpha — 已完成 (5 Phases)

| Phase | 焦点 | 状态 |
|-------|------|:----:|
| **1** | CLI MVP: 扫描 → 解析 → 按 level 生成 Wiki | ✅ |
| **2** | MCP Server (stdio/TCP, 6 tools) | ✅ |
| **3** | 全文检索 (bleve) + 代码注释提取 | ✅ |
| **4** | 向量化 + 语义搜索 (Hybrid BM25/向量) | ✅ |
| **5** | `[[wiki-link]]` 交叉链接 + 反向链接 | ✅ |

### Beta — 规划中

参见 [beta-roadmap.md](beta-roadmap.md) 框架路线图。

- **D** — AI 增强与 Token 优化（Level 1/2/3 全套） ✅ 已完成
- **G1** — Dockerfile + Docker 集成测试 📋 待做
- **G2** — 真实语义 Embedding API 📋 待做
- **G3** — MCP Resources / Prompts 📋 待做
- **G4** — 大项目性能优化（1000+ 文件） 📋 待做
- **C** — 模板系统 + 主题定制 📋 待定
- **E** — 插件系统 📋 待定
- **F** — 导入/导出/生态集成 🔍 待研究

> Watch 模式已移除（代码变更由 CodeGraph 负责）。

---

## 8. 技术决策记录

| 决策 | 选择 | 理由 |
|------|------|------|
| 语言 | Go | 单二进制、标准库完善、交叉编译友好、容器镜像小 |
| 配置加载 | viper | 支持 YAML + 环境变量 + flag 多层合并, Go 社区标准 |
| CLI 框架 | cobra | Go 社区事实标准 |
| Markdown 解析 | goldmark | Go 最快最标准的 MD 解析器，支持扩展 |
| YAML 解析 | gopkg.in/yaml.v3 | 标准选择，frontmatter 解析需要 |
| 二进制探测 | 标准库 utf8.Valid | 零依赖，轻量可靠 |
| MCP 协议 | 自实现（基于 JSON-RPC） | 协议简单，需要完全控制 |
| 全文索引 | bleve | Go 原生全文搜索引擎 |
| 测试 | 标准 testing + testify | 社区标准 |
| 构建 | Makefile + goreleaser | 自动化构建、发布 |

---

## 9. 非功能性需求

| 指标 | 目标 |
|------|------|
| 二进制大小 | < 15MB（静态编译） |
| 构建速度 | 10万文件 < 5秒 |
| 搜索响应 | < 100ms |
| MCP 启动 | < 1秒 |
| 内存占用 | 空闲 < 10MB，构建 < 100MB |
| 平台支持 | Linux / macOS / Windows |
