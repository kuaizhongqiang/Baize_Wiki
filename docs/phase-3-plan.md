# Baize Wiki — Phase 3 实施总结

> 版本: v1（复盘记录）
> 焦点: 全文检索 + 代码注释提取
> 状态: ✅ 已完成 (2026-06-07)
> 实现: #38-#43, 合并到 main (5c8a9f9)

## 1. Phase 3 范围

**目标**：为 Wiki 添加全文检索能力，支持 AI Agent 通过关键词搜索 Wiki 内容，同时支持代码注释提取。

**包含**：
- ✅ 全文索引引擎（基于 bleve）
- ✅ Build 管线集成（自动建索引）
- ✅ CLI `search` 命令
- ✅ MCP `wiki_search` 工具
- ✅ `--scan-all` 全扫描模式
- ✅ 代码注释提取（20+ 语言）
- ✅ 测试覆盖

**不包含**（按设计范围，留给后续 Phase）：
- ❌ 向量语义搜索 (→ Phase 4)
- ❌ 混合检索（BM25 + 向量）(→ Phase 4)

---

## 2. 架构

```
baize-wiki build
  → Scanner (支持 --scan-all)
  → Parser (含代码注释提取)
  → Generator → Wiki 文件
  → Index.Build() → .baize/index.bleve   ← Phase 3 新增

baize-wiki search <query>
  → Index.Search() → 结果排序 → 输出

MCP wiki_search
  → Agent 对话 → tool_wiki_search → Index.Search()
```

### 包结构

```
internal/core/index/
├── index.go          bleve 全文索引引擎 (Build/Search/Close)
└── index_test.go     索引测试

internal/core/parser/
├── comments.go       代码注释提取 (20+ 语言)
└── comment_test.go   注释提取测试

internal/app/
└── search.go         CLI search 命令实现

cmd/baize-wiki/
└── main.go           注册 search 命令
```

---

## 3. 核心实现

### 3.1 全文索引引擎 (bleve)

**文件**: `internal/core/index/index.go` (203 行)

使用 `github.com/blevesearch/bleve/v2` 实现：

| 方法 | 功能 | 参数 |
|------|------|------|
| `NewIndex(path)` | 创建或打开索引 | 索引路径 |
| `Build(ctx, pages)` | 批量建索引 | 页面列表 |
| `Search(ctx, query, opts)` | 关键词搜索 | 查询 + 选项 |
| `Close()` | 释放资源 | - |

**索引文档结构**：

```go
type doc struct {
    Path    string   // keyword, stored
    Title   string   // text, stored
    Content string   // text, stored
    Tags    []string // keyword, stored
}
```

### 3.2 Build 集成

**文件**: `internal/app/build.go` (增量 20 行)

Generator 完成后自动调用索引构建：

```go
// 5. Build full-text index (non-fatal on failure)
indexPath := filepath.Join(absOutput, ".baize", "index.bleve")
idx, err := index.NewIndex(indexPath)
if err == nil {
    idx.Build(ctx, pages)
    idx.Close()
}
```

索引路径：`<wiki-dir>/.baize/index.bleve`

### 3.3 CLI search 命令

**文件**: `internal/app/search.go` (100 行)

```bash
baize-wiki search <query> [flags]

Flags:
  -w, --wiki string     Wiki 目录 (默认 ./wiki)
  -l, --limit int       返回结果数 (默认 10)
  -t, --tags strings    按标签筛选
  -c, --with-content    返回全文
  -j, --json            JSON 格式输出
```

### 3.4 MCP wiki_search

**文件**: `internal/mcp/tools.go` (增量 57 行)

AI Agent 通过 `wiki_search` 工具搜索 Wiki：

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| query | string | 是 | 搜索关键词 |
| tags | string[] | 否 | 标签筛选 |
| limit | int | 否 | 最大结果数 (默认 10) |
| include_content | bool | 否 | 返回全文 |

### 3.5 --scan-all 全扫描

**文件**: `cmd/baize-wiki/main.go`, `internal/app/build.go`

```bash
baize-wiki build --scan-all ./project
```

启用后扫描所有文本文件（不限于 .md/.mdx），适合含代码注释的代码库。

### 3.6 代码注释提取

**文件**: `internal/core/parser/comments.go` (158 行)

从代码文件中提取顶层文档注释，支持 20+ 语言：

| 语言 | 注释语法 |
|------|---------|
| Go | `//`, `/* */` |
| Python | `#` |
| JavaScript/TypeScript | `//`, `/* */` |
| Rust | `//`, `/* */` |
| Java/Swift/Kotlin | `//`, `/* */` |
| YAML/TOML | `#` |

提取算法：

```
读取文件内容 → 逐行扫描
  ├─ 行注释 (//, #) → 收集
  ├─ 块注释 (/* */) → 收集块内文本
  └─ 遇到非注释代码 → 停止
返回 → 文件头部注释块
```

---

## 4. 依赖

| 依赖 | 用途 | 版本 |
|------|------|------|
| `github.com/blevesearch/bleve/v2` | 全文搜索引擎 | v2.6.0 |

---

## 5. 代码规范

- 索引构建失败不阻塞 Wiki 输出（warning 级别）
- 单文件索引错误不影响其他文件（batch 模式）
- 所有导出类型和方法有 Go doc 注释
- 与 Phase 1/2 保持一致的错误处理风格

---

## 6. 里程碑完成情况

| 里程碑 | Issue | 状态 |
|--------|-------|------|
| M1: 全文索引引擎 | #38 | ✅ |
| M2: Build 集成 | #39 | ✅ |
| M3: CLI search 命令 | #40 | ✅ |
| M4: MCP wiki_search | #41 | ✅ |
| M5: 代码注释提取 + scan-all | #42 | ✅ |
| M6: 测试 | #43 | ✅ |

---

## 7. 已知限制

1. bleve 中文分词使用内置 analyzer（非专用中文分词），对混合中英文搜索效果可能不如专门的搜索引擎
2. 代码注释提取只提取文件头部注释块，不提取函数/方法级注释
3. `--scan-all` 模式下代码文件以纯文本索引，未做 AST 级别的语义提取
4. 无增量索引优化——每次 build 都会重建完整索引
