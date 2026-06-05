# Baize Wiki — 数据模型规范

> 版本: v1 (草案)
> 状态: 待审计

## 1. Wiki 输出目录规范

Baize Wiki 的输出是一个标准文件系统目录树，以 Markdown 文件为载体。

```
wiki-output/
├── _index.md                 # Wiki 根索引页
├── _config.yaml              # Wiki 运行时元数据
│
├── <category>/               # 按语义分组的目录
│   ├── _index.md             #   分类索引页
│   ├── <page>.md             #   具体页面
│   └── ...
│
├── .baize/                   # Baize 内部数据目录
│   ├── meta.json             #   Wiki 元信息缓存
│   ├── index.bleve           #   全文索引 (Phase 3)
│   └── vectors/              #   向量数据 (Phase 4)
│
└── baize.lock                # 生成锁定文件，标识该目录为 Baize Wiki
```

### 规则

1. 每个目录下的 `_index.md` 是该目录的索引/概览页
2. 源文件中的目录结构映射到 Wiki 输出的目录结构
3. 扁平输入可通过配置按标签/分类自动组织
4. `.baize/` 目录是内部数据，不应被手动编辑

---

## 2. 核心 Go 数据结构

### 2.1 Wiki

```go
// Wiki 代表一个完整的 Wiki 知识库
type Wiki struct {
    ID          string    `json:"id"`           // 唯一标识 (UUID)
    Name        string    `json:"name"`          // 名称
    Description string    `json:"description"`   // 描述
    SourcePath  string    `json:"source_path"`   // 源文件路径
    OutputPath  string    `json:"output_path"`   // 输出目录路径
    Config      Config    `json:"config"`        // 配置
    PageCount   int       `json:"page_count"`    // 页面总数
    CreatedAt   time.Time `json:"created_at"`    // 创建时间
    UpdatedAt   time.Time `json:"updated_at"`    // 最后更新时间
    Version     int       `json:"version"`       // 版本号 (递增)
}
```

### 2.2 Page

```go
// Page 代表 Wiki 中的一个页面
type Page struct {
    ID         string     `json:"id"`            // 唯一标识 (hash of path)
    WikiID     string     `json:"wiki_id"`       // 所属 Wiki ID
    Path       string     `json:"path"`          // Wiki 内相对路径 (含 .md)
    Title      string     `json:"title"`         // 标题
    Content    string     `json:"content"`       // Markdown 原文
    HTML       string     `json:"-"`             // 渲染后的 HTML (运行时)
    Meta       Frontmatter `json:"meta"`         // frontmatter 元数据
    Sections   []Section  `json:"sections"`      // 文档结构 (heading 树)
    Tags       []string   `json:"tags"`          // 标签
    Links      []Link     `json:"links"`         // 本页发出的链接
    Backlinks  []Link     `json:"backlinks"`      // 引用本页的链接
    Depth      int        `json:"depth"`         // 目录深度
    Weight     int        `json:"weight"`        // 排序权重
    SourceFile string     `json:"source_file"`   // 源文件绝对路径
    UpdatedAt  time.Time  `json:"updated_at"`    // 最后修改时间
}
```

### 2.3 Frontmatter

```go
// Frontmatter 是 Markdown 文件的 YAML 元数据
type Frontmatter struct {
    Title       string            `yaml:"title" json:"title"`
    Description string            `yaml:"description,omitempty" json:"description,omitempty"`
    Tags        []string          `yaml:"tags,omitempty" json:"tags,omitempty"`
    Aliases     []string          `yaml:"aliases,omitempty" json:"aliases,omitempty"`
    Weight      *int              `yaml:"weight,omitempty" json:"weight,omitempty"`
    Draft       bool              `yaml:"draft,omitempty" json:"draft,omitempty"`
    Category    string            `yaml:"category,omitempty" json:"category,omitempty"`
    Custom      map[string]any    `yaml:",inline" json:"custom,omitempty"` // 自定义字段
}

> ⚠️ **兼容性说明**：`Custom` 通过 `yaml:",inline"` 捕获未在 Frontmatter 结构体中定义的字段。这意味着未来版本新增的具名字段（如 `type`、`date`）可能与旧版本写入的自定义字段发生冲突。自定义字段请避免使用未来可能保留的字段名。详见 [向后兼容策略](#-向后兼容策略)。
}
```

示例：
```yaml
---
title: Baize Wiki 架构设计
description: 系统架构和模块划分的详细文档
tags: [architecture, design, go]
weight: 1
category: 技术文档
---
```

### 2.4 Section

```go
// Section 表示文档中的一个标题段落，形成树状结构
type Section struct {
    ID        string     `json:"id"`         // anchor id
    Level     int        `json:"level"`      // heading 级别 (1-6)
    Title     string     `json:"title"`      // 标题文本
    Content   string     `json:"content"`    // 该段落下文本摘要
    Children  []Section  `json:"children"`   // 子段落
}
```

### 2.5 Link

```go
// Link 表示页面间的一个交叉引用
type Link struct {
    SourceID   string `json:"source_id"`    // 源页面 ID
    TargetID   string `json:"target_id"`    // 目标页面 ID
    TargetPath string `json:"target_path"`  // 目标页面路径
    Text       string `json:"text"`         // 链接文本
    Type       LinkType `json:"type"`       // 链接类型
}

type LinkType string

const (
    LinkInternal LinkType = "internal"  // [[wiki-link]] 内部链接
    LinkExternal LinkType = "external"  // https:// 外部链接
    LinkResource LinkType = "resource"  // ./image.png 资源引用
    LinkAuto     LinkType = "auto"      // 自动检测的页面引用
)
```

---

## 3. 配置模型

### 3.1 配置文件 (`baize.yaml`)

```yaml
# baize.yaml — Baize Wiki 项目配置
name: "我的 Wiki"              # Wiki 名称
description: "项目知识库"       # 可选描述

# 扫描配置
scan:
  paths:                     # 源路径（默认 ./）
    - ./docs
    - ./README.md
  exclude:                   # 排除模式 (gitignore 语法)
    - "node_modules/**"
    - "vendor/**"
    - ".git/**"
  max_size: 10485760         # 单文件最大字节 (10MB, 默认)
  # include 控制: "能扫尽扫", 不设 whitelist; 通过 exclude 排除

# 输出配置
output:
  dir: "./wiki"              # 输出目录
  level: 2                   # 输出复杂度: 1 | 2 | 3 (默认 2)
  clean: false               # 构建前是否清空

# Wiki 组织方式
organize:
  by: "directory"            # directory | flat | tags (Phase 2+)

# 功能开关
features:
  draft: false               # 是否包含草稿 (Phase 2+)

# MCP 服务 (Phase 2+)
# serve:
#   transport: stdio          # stdio | tcp
#   addr: ":8080"            # tcp 模式下监听地址
```

### 3.2 Go 结构

```go
type Config struct {
    Name        string         `yaml:"name" json:"name"`
    Description string         `yaml:"description" json:"description,omitempty"`
    Scan        ScanConfig     `yaml:"scan" json:"scan"`
    Output      OutputConfig   `yaml:"output" json:"output"`
    Organize    OrganizeConfig `yaml:"organize" json:"organize"`
    Features    FeatureConfig  `yaml:"features" json:"features"`
}

type ScanConfig struct {
    Paths   []string `yaml:"paths" json:"paths"`
    Exclude []string `yaml:"exclude" json:"exclude"`
    MaxSize int64    `yaml:"max_size" json:"max_size"`
}

type OutputConfig struct {
    Dir   string `yaml:"dir" json:"dir"`
    Level int    `yaml:"level" json:"level"` // 1 | 2 | 3
    Clean bool   `yaml:"clean" json:"clean"`
}

type OrganizeConfig struct {
    By string `yaml:"by" json:"by"` // directory | flat | tags (Phase 2+)
    // Phase 1: 输出结构由 output.level 唯一控制, organize.by 暂不生效
}

type FeatureConfig struct {
    Draft bool `yaml:"draft" json:"draft"`
}
```

---

## 4. 忽略规则 (`.baizeignore`)

语法兼容 `.gitignore`，按行读取，`#` 开头为注释：

```gitignore
# .baizeignore — 扫描时忽略的文件/目录
node_modules/
vendor/
.git/
*.exe
*.bin
.DS_Store
Thumbs.db
```

---

## 5. 内部元数据格式 (`.baize/meta.json`)

Wiki 构建完成后，将元信息写入 `.baize/meta.json`，供 MCP Server 或后续增量构建使用：

```json
{
  "id": "wiki_abc123",
  "name": "我的 Wiki",
  "description": "项目知识库",
  "source_path": "/home/user/project/docs",
  "output_path": "/home/user/project/wiki",
  "page_count": 42,
  "created_at": "2026-06-05T10:00:00Z",
  "updated_at": "2026-06-05T14:30:00Z",
  "version": 3,
  "config_hash": "sha256..." 
}
```

---

## 6. 错误模型

```go
// 领域错误分类
var (
    ErrWikiNotFound    = errors.New("wiki not found")     // 指定 wiki 目录不存在
    ErrPageNotFound    = errors.New("page not found")     // 指定页面不存在
    ErrSourceNotFound  = errors.New("source not found")   // 源目录不存在
    ErrInvalidConfig   = errors.New("invalid config")     // 配置错误
    ErrScanFailed      = errors.New("scan failed")        // 扫描失败
    ErrGenerateFailed  = errors.New("generate failed")    // 生成失败
    ErrEmptySource     = errors.New("empty source")       // 源目录无有效文件
)

type Error struct {
    Code    string `json:"code"`    // 机器可读错误码
    Message string `json:"message"` // 人类可读消息
    Detail  string `json:"detail,omitempty"` // 详细信息
    Err     error  `json:"-"`       // 原始 error
}
```

---

## 7. 数据流接口合约

核心域各组件通过 Go interface 解耦：

```go
// Scanner 扫描源目录，返回文件列表
type Scanner interface {
    Scan(ctx context.Context, root string, cfg ScanConfig) ([]FileInfo, error)
}

// Parser 解析文件内容为 Page
type Parser interface {
    Parse(ctx context.Context, file FileInfo) (*Page, error)
    ParseBatch(ctx context.Context, files []FileInfo) ([]*Page, error)
}

// Linker 计算页面间的交叉引用
type Linker interface {
    Link(ctx context.Context, pages []*Page) error
}

// Generator 将 Pages 生成 Wiki 目录
type Generator interface {
    Generate(ctx context.Context, wiki *Wiki, pages []*Page) error
}

// Indexer 构建和查询全文索引
type Indexer interface {
    Build(ctx context.Context, pages []*Page) error
    Search(ctx context.Context, query string, opts SearchOpts) ([]SearchResult, error)
}

// Storage 读写 Wiki 文件
type Storage interface {
    WritePage(ctx context.Context, path string, page *Page) error
    WriteIndex(ctx context.Context, wiki *Wiki) error
    ReadMeta(ctx context.Context, wikiDir string) (*Wiki, error)
    ReadPage(ctx context.Context, wikiDir string, path string) (*Page, error)
}
```
