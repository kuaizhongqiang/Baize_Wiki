# Baize Wiki — Phase 4 实施计划

> 版本: v1
> 状态: 草案
> 焦点: 向量化 + 语义搜索
> 预估工时: ~1 周（单人开发）

## 1. Phase 4 范围

**目标**：为 Wiki 添加向量化语义搜索能力，与 Phase 3 的全文检索形成混合检索（Hybrid Search），提升 AI Agent 的检索准确率。

**包含**：
- ✅ Vector Store 接口与实现（本地存储 + 余弦相似度搜索）
- ✅ 嵌入（Embedding）生成引擎（本地默认 + 可选 API）
- ✅ 混合检索（BM25 + 向量相似度加权融合）
- ✅ CLI `--semantic` 标志 / MCP `wiki_search` 升级
- ✅ 构建管线集成（Build 时生成向量索引）
- ✅ 配置支持（`baize.yaml` 向量相关选项）
- ✅ 测试

**不包含**：
- ❌ RAG 增强生成（留给后续）
- ❌ 跨文档语义聚类（留给后续）
- ❌ 外部向量数据库集成（Milvus/Pinecone 等）

---

## 2. 架构

```
                   ┌───────────────────┐
                   │    Hybrid Searcher │
                   │  (加权融合排序)     │
                   └──┬────────────┬───┘
                      │            │
              ┌───────▼──┐   ┌────▼────────┐
              │ BM25     │   │  Vector      │
              │ (Phase 3)│   │  Search      │
              │ bleve    │   │  (Phase 4)   │
              └──────────┘   └──────┬───────┘
                                    │
                           ┌────────▼────────┐
                           │  Embedder       │
                           │  (本地/API)      │
                           └────────┬────────┘
                                    │
                           ┌────────▼────────┐
                           │  Vector Store   │
                           │  .baize/vectors/ │
                           └─────────────────┘
```

### 混合检索评分公式

```
score = α * BM25_score + (1-α) * vector_score

α ∈ [0, 1] 由 --hybrid-weight 控制
α=1.0 → 纯全文检索（Phase 3 兼容）
α=0.0 → 纯语义检索
α=0.5 → 均衡模式（默认）
```

---

## 3. 核心设计

### 3.1 Vector Store 接口

```go
// VectorStore 存储和检索向量嵌入。
type VectorStore interface {
    // Store 存储一个页面的向量嵌入。
    Store(ctx context.Context, pageID string, embedding []float32) error

    // Search 返回与查询向量最相似的页面。
    Search(ctx context.Context, embedding []float32, limit int) ([]VectorResult, error)

    // Close 释放资源。
    Close() error
}

type VectorResult struct {
    PageID    string  `json:"page_id"`
    Path      string  `json:"path"`
    Title     string  `json:"title"`
    Score     float64 `json:"score"` // 余弦相似度
}

// Embedder 将文本转换为向量嵌入。
type Embedder interface {
    // Embed 将文本转为向量。
    Embed(ctx context.Context, text string) ([]float32, error)

    // Dim 返回向量维度。
    Dim() int
}
```

### 3.2 本地 Embedder（默认）

使用基于词频的轻量向量表示，零外部依赖：

```go
type LocalEmbedder struct {
    dim    int          // 默认 256 维
    vocab  map[string]int
}

// 实现原理：
// 1. 分词（简单空格 + 标点分割，含 CJK 字符）
// 2. 构建词频向量（类似 TF 但不做 IDF，避免需要语料库）
// 3. 输出为固定维度的哈希特征向量（特征哈希 / Feature Hashing）
```

**为什么选 Feature Hashing：**
- 零外部依赖（不需要模型文件、不需要 API）
- 固定维度输出（256 维），适合存储和比较
- 对中文也有基本效果（按字/词哈希）
- 虽然不如深度学习 embedding 准确，但作为默认方案够用

### 3.3 API Embedder（可选）

支持通过配置切换到外部 Embedding API：

```yaml
vector:
  mode: api                     # local | api
  provider: openai              # openai | anthropic
  api_key: "${OPENAI_API_KEY}"  # 环境变量注入
  model: text-embedding-3-small # 模型名
```

### 3.4 混合检索器

```go
type HybridSearch struct {
    bm25    *index.Index      // Phase 3 索引
    vector  VectorStore
    embed   Embedder
    weight  float64           // α 值，BM25 权重
}

func (h *HybridSearch) Search(ctx context.Context, query string, opts SearchOpts) ([]SearchResult, error) {
    // 1. BM25 搜索 → BM25 结果集
    // 2. 向量搜索 → Embed query → Vector 结果集
    // 3. 融合: score = α * bm25_score + (1-α) * vector_score
    // 4. 排序返回 TopN
}
```

---

## 4. 文件清单

```
文件:
├── internal/core/vector/
│   ├── store.go          VectorStore 接口 + 内存实现
│   ├── store_test.go
│   ├── embedder.go       Embedder 接口 + LocalEmbedder 实现
│   ├── embedder_test.go
│   └── hybrid.go         混合检索器
├── internal/core/vector/embed_api.go     API Embedder（可选）
├── internal/app/build.go                更新：构建时生成向量索引
├── internal/app/search.go               更新：支持 --semantic 标志
├── internal/mcp/tools.go                更新：wiki_search 支持向量
├── internal/core/model/config.go        更新：VectorConfig
└── internal/core/index/index.go         更新：BM25 搜索暴露给 HybridSearch
```

---

## 5. 里程碑

| 里程碑 | 内容 | 预估 |
|--------|------|------|
| **M1: Vector Store** | 接口定义 + 内存实现 + 持久化 + 测试 | Day 1-2 |
| **M2: Local Embedder** | Feature Hashing 向量化 + Embedder 接口 + 测试 | Day 2-3 |
| **M3: Hybrid Search** | 混合检索器 + 加权融合 + 测试 | Day 3-4 |
| **M4: 构建管线集成** | Build 时生成向量索引 + 配置加载 | Day 4-5 |
| **M5: CLI + MCP 更新** | `--semantic` 标志 + wiki_search 升级 | Day 5 |
| **M6: 测试** | 单元 + 集成 + E2E | Day 6-7 |

---

## 6. 数据流

### Build 时的向量索引构建

```
build sourceDir
  → Scanner → file list
  → Parser → pages
  → Generator → wiki output
  → index.Build(pages) → .baize/index.bleve          (Phase 3)
  → Embedder.Embed(page.Title + page.Content) → vector  (Phase 4 新增)
  → VectorStore.Store(pageID, vector)                  (Phase 4 新增)
  → persist → .baize/vectors/                          (Phase 4 新增)
```

### 搜索时的混合检索

```
baize-wiki search "关键词" --semantic
  → 混合检索:
    1. BM25 搜索 bleve → BM25 结果（含 BM25 分数）
    2. Embed query → query vector
    3. VectorStore.Search(queryVector) → 向量结果（含余弦相似度）
    4. 加权融合: score = α * bm25_score + (1-α) * vec_score
    5. 排序返回 TopN
```

---

## 7. 配置变更

```yaml
# baize.yaml 新增
features:
  draft: false
  vector: true             # 是否启用向量索引

vector:
  mode: local              # local | api
  # API mode (可选)
  provider: openai         # openai | anthropic
  api_key: ""              # 建议用环境变量 BAIZE_VECTOR_API_KEY
  model: text-embedding-3-small

# CLI flag 新增
# --semantic: 搜索时启用语义检索（默认基于 BM25）
# --hybrid-weight: BM25 权重 (0.0-1.0, 默认 0.5)
```

---

## 8. 代码规范

- Vector Store 和 Embedder 均面向接口编程，便于替换实现
- LocalEmbedder 不引入外部依赖，使用标准库 + 简单哈希
- 向量持久化使用 gob 或 JSON 编码，存储在 `.baize/vectors/`
- 混合检索器接受 α 参数，α=1.0 时退化为纯 BM25 搜索
- 向量索引构建失败不阻塞 build（同 Phase 3 索引策略）
- 所有新增 public 类型和函数写 Go doc 注释

---

## 9. 风险与缓解

| 风险 | 影响 | 缓解措施 |
|------|------|---------|
| Feature Hashing 语义效果差 | 搜索结果不准确 | 作为默认方案，提供 API Embedder 作为升级路径 |
| 向量维度高导致内存占用大 | 百万级页面时 OOM | Phase 4 目标中小型 Wiki（< 10 万页），必要时后续加磁盘映射 |
| 外部 Embedding API 延迟 | 搜索变慢 | API Embedder 为可选项；默认本地方案零延迟 |
| 混合检索参数α难调优 | 搜索结果不如预期 | 默认 0.5 均衡，用户可通过 flag 微调 |
