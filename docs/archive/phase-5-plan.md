# Baize Wiki — Phase 5 实施计划

> 版本: v1
> 状态: 草案
> 焦点: 交叉链接 + `[[wiki-link]]` 解析
> 预估工时: ~3-5 天（单人开发）

## 1. Phase 5 范围

**目标**：解析 `[[wiki-link]]` 语法，计算页面间的交叉引用和反向链接（backlinks），并在 Wiki 输出和信息展示中体现链接关系。

**包含**：
- ✅ `[[wiki-link]]` 语法解析（从 Markdown 内容中提取）
- ✅ 链接解析：将 `[[页面名]]` 映射到目标页面路径
- ✅ 反向链接（Backlink）自动计算
- ✅ 外部链接和资源链接检测
- ✅ Build 管线集成
- ✅ CLI `info` 命令显示链接/反向链接数
- ✅ `_index.md` 可选显示反向链接
- ✅ 测试

**不包含**：
- ❌ 悬空链接（dangling link）检测报告（后续增强）
- ❌ 链接自动补全/建议（后续增强）
- ❌ 跨 Wiki 链接（后续增强）

---

## 2. 架构

```
Parser (解析 [[link]])             Linker (计算交叉引用)
  ┌──────────────────┐              ┌──────────────────┐
  │ Markdown 解析时    │              │ 输入: Pages[]     │
  │ 提取 [[xxx]] 语法  │  ───Pages──▶ │ Link(ctx, pages)  │
  │ 存入 Page.Links[] │              │ 输出: 填充         │
  └──────────────────┘              │ Page.Links[]      │
                                    │ Page.Backlinks[]  │
                                    └────────┬─────────┘
                                             │
                                    ┌────────▼─────────┐
                                    │ Generator 可选     │
                                    │ _index.md 显示链接  │
                                    │ info 命令显示       │
                                    └──────────────────┘
```

### 数据流

```
Build 流水线 (更新):

Scanner → Parser (含 [[link]] 提取)
  → Linker.Link(pages)  ← Phase 5 新增
  → Generator.Generate(wiki, pages)
  → index.Build(pages)
  → VectorStore.Store(pages)
```

---

## 3. 核心设计

### 3.1 `[[wiki-link]]` 解析

解析规则：

```
[[页面名]]              → 链接到同目录下的页面
[[目录/页面名]]        → 链接到指定路径的页面
[[页面名|显示文本]]     → 带自定义显示文本的链接
[[#锚点]]              → 同页面内锚点链接（Phase 5 简单处理）
```

提取算法（在 Parser 中）：

```go
// ExtractWikiLinks extracts [[wiki-link]] references from content.
func ExtractWikiLinks(content string) []LinkRef {
    // 正则: \[\[([^\[\]]+?)(?:\|([^\[\]]+?))?\]\]
    // 匹配 [[target]] 和 [[target|text]] 两种形式
    // 返回列表: [{target, text}]
}
```

### 3.2 链接解析（Linker）

```go
// Linker 计算页面间的交叉引用。
type Linker struct{}

// Link 解析所有页面的 [[wiki-link]]，计算 Links 和 Backlinks。
func (l *Linker) Link(ctx context.Context, pages []*Page) error {
    // 1. 为所有页面建立 path → page 的索引映射
    // 2. 遍历每个页面，解析其内容中的 [[link]]
    // 3. 对每个 [[link]] 尝试匹配目标页面
    // 4. 填充 Page.Links 和 Page.Backlinks
}
```

#### 链接解析策略

```
输入的 [[页面名]]  →  匹配策略:
  1. 精确匹配: path == "页面名"
  2. 标题匹配: page.Title == "页面名"
  3. 文件名匹配: filepath.Base(path) == "页面名"
  4. 模糊匹配（Phase 5 不做，后续增强）

匹配失败 → 链接标记为 dangling（后续显示警告）
```

### 3.3 Link 模型（更新 Page）

```go
// Link 表示页面间的一个交叉引用。
type Link struct {
    SourceID   string   `json:"source_id"`    // 源页面 ID
    TargetID   string   `json:"target_id"`    // 目标页面 ID (空=dangling)
    TargetPath string   `json:"target_path"`  // 目标页面路径
    Text       string   `json:"text"`         // 链接显示文本
    Type       LinkType `json:"type"`         // internal | external | resource | auto
}

type LinkType string

const (
    LinkInternal LinkType = "internal"  // [[wiki-link]]
    LinkExternal LinkType = "external"  // https:// 外部链接
    LinkResource LinkType = "resource"  // ./image.png 资源
    LinkAuto     LinkType = "auto"      // 自动检测（暂不实现）
)
```

---

## 4. 文件清单

```
internal/core/linker/
├── linker.go          Linker 实现（解析 + 匹配 + backlink 计算）
└── linker_test.go     测试

internal/core/parser/
├── wikilink.go        [[link]] 解析提取 + 测试
├── wikilink_test.go

internal/core/model/
├── page.go            更新: 添加 Link/LinkType struct（已有占位）

internal/app/
├── build.go           更新: 在 parser 和 generator 之间插入 Linker
├── info.go            更新: 显示链接/反向链接数

internal/core/generator/
├── generator.go       更新: _index.md 可选显示 backlinks
```

---

## 5. 里程碑

| 里程碑 | 内容 | 预估 |
|--------|------|------|
| **M1: Link 模型** | 定义 Link/LinkType 结构体，更新 Page | Day 1 |
| **M2: `[[link]]` 解析** | 正则提取 [[target]] 和 [[target\|text]] | Day 1-2 |
| **M3: Linker** | 页面匹配 + Links 填充 + Backlinks 计算 | Day 2-3 |
| **M4: Build 集成** | 在 Pipeline 中插入 Linker | Day 3 |
| **M5: CLI/MCP 增强** | info 显示链接 + _index.md 增强 | Day 4 |
| **M6: 测试** | 单元 + 集成测试 | Day 4-5 |

---

## 6. 代码规范

- `[[link]]` 解析使用标准库 `regexp`，不引入新依赖
- 链接匹配策略：精确 path → 标题 → 文件名（依次降级）
- 匹配失败不报错，TargetID 留空即可
- Links 和 Backlinks 在 Generator 之前填充完毕
- 反向链接数可在 `_index.md` 中以 "被 N 个页面引用" 显示

---

## 7. 风险与缓解

| 风险 | 影响 | 缓解措施 |
|------|------|---------|
| `[[link]]` 匹配歧义（多个页面同名） | 链接到错误页面 | 按 path → title → file 依次精确匹配，首个匹配为准 |
| 大 Wiki 的 Linker 性能 | Build 变慢 | Phase 5 不做优化，预期 < 100ms/万页 |
| Markdown 内容含 `[[not-a-link]]` 误匹配 | 错误解析 | 只在 `.md`/`.mdx` 文件中解析 `[[link]]` |
