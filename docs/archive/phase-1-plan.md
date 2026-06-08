# Baize Wiki — Phase 1 实施计划

> 版本: v2 (基于讨论更新)
> 状态: 待审计
> 预估工时: ~2-3 周（单人开发）

## 1. Phase 1 范围

**目标**：一个可用的 CLI 工具，能扫描指定路径下的所有文档 → 解析 → 按 Level 1/2/3 生成结构化的 Wiki 输出（文件夹 + MD 文件）。

**包含**：
- ✅ 基础项目脚手架 (Go module, Makefile, goreleaser)
- ✅ 配置加载 (`baize.yaml`)
- ✅ 文件扫描器 (Scanner — 全文本文件扫描 + 二进制探测)
- ✅ 文档解析器 (Parser — frontmatter + Markdown 标题结构)
- ✅ Wiki 生成器 (Generator — 支持 Level 1/2/3 输出)
- ✅ 文件存储 (Storage — 写入文件系统)
- ✅ CLI 命令: `init`, `build`, `info`
- ✅ 测试覆盖

**不包含**（留给后续 Phase）：
- ❌ `[[wiki-link]]` 交叉链接 (→ Phase 5)
- ❌ MCP Server (→ Phase 2)
- ❌ 全文搜索 (→ Phase 3)
- ❌ 代码注释提取 (→ Phase 3)
- ❌ 向量化 (→ Phase 4)
- ❌ 文件监听 watch mode (→ Phase 2)
- ❌ Docker 容器化 (→ Phase 2)

---

### 构建策略说明

Phase 1 采用 **全量构建**：

- 每次 `baize-wiki build` 都是完整重建
- 先清空输出目录，再重新生成全部文件
- 不存在增量/缓存逻辑
- `--no-cache` 标志不适用，已移除
- 对于大型项目（1000+ 文件），全量构建的性能优化留到后续 Phase

---

## 2. 核心链路

```
baize-wiki build ./source --level 2 --output ./wiki

step 1: 加载配置 (baize.yaml + flag 覆写)
step 2: Scanner — 遍历所有文件
           ├── .baizeignore 匹配? → 跳过
           ├── 二进制探测? → 跳过
           └── 其他 → 加入文件列表
step 3: Parser — 依次解析每个文件
           ├── .md → frontmatter + 标题结构解析
           ├── .txt/json/yaml → 纯文本读取
           └── 其他文本 → 纯文本读取
step 4: Generator — 按 --level 组织输出结构
           ├── Level 1: 拍平为一层，合并同主题内容
           ├── Level 2: 一层子目录（默认）
           └── Level 3: 深层目录（max 3 层）
step 5: Storage — 写入文件系统
           ├── wiki/xxx.md       (页面文件)
           ├── wiki/xxx/_index.md (目录索引)
           └── wiki/.baize/meta.json (元信息)
```

---

## 3. 里程碑

| 里程碑 | 内容 | 预估 |
|--------|------|------|
| **M1: 项目骨架** | Go module、包结构、Makefile、goreleaser、CLI 框架 | Day 1 |
| **M2: 核心模型** | model 包、config 包、error 定义 + 测试 | Day 2-3 |
| **M3: 扫描器** | scanner 包（全文件扫描 + 二进制探测 + 忽略规则）+ 测试 | Day 4-6 |
| **M4: 解析器** | parser 包（frontmatter + markdown + 纯文本）+ 测试 | Day 7-9 |
| **M5: 生成器** | generator（Level 1/2/3）+ storage + 测试 | Day 10-14 |
| **M6: CLI 集成** | init + build + info 命令串联、端到端测试 | Day 15-17 |
| **M7: 收尾** | README、示例、文档完善、goreleaser 发布 | Day 18-19 |

---

## 4. 详细任务分解

### M1: 项目骨架

```
文件清单:
├── go.mod
├── Makefile
├── .goreleaser.yaml
├── cmd/baize-wiki/main.go
├── internal/app/build.go        (stub)
├── internal/app/info.go         (stub)
└── internal/config/config.go    (stub)
```

任务：

1. **初始化 Go module**
   - `module github.com/kuaizhongqiang/baize-wiki`
   - Go 1.22+
   - 引入依赖: `github.com/spf13/cobra`, `github.com/spf13/viper`
   - go.sum 锁定版本

2. **Makefile**
   - `make build` — 编译到 `bin/baize-wiki`
   - `make test` — 全部测试
   - `make lint` — golangci-lint
   - `make clean` — 清理
   - `make cross` — 交叉编译: linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, windows/amd64

3. **goreleaser.yaml**
   - 目标平台: linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, windows/amd64
   - 压缩: tar.gz (Linux/macOS), zip (Windows)
   - 注入版本号和 commit hash 到 ldflags

4. **CLI main 入口**
   - cobra root command
   - `init` 子命令 (stub)
   - `build` 子命令 (stub)
   - `info` 子命令 (stub)
   - `--version` 标志 (ldflags)

5. **配置加载 stub**
   - 读取 `baize.yaml`
   - 支持 `--config` flag

**验收条件：**
- [x] `go build ./cmd/baize-wiki` 成功
- [x] `baize-wiki --help` 显示命令列表
- [x] `make test` 通过
- [x] `make cross` 产出 5 个平台的二进制
- [x] `goreleaser build --snapshot` 成功

---

### M2: 核心模型

```
文件清单:
├── internal/core/model/
│   ├── page.go           Page, Section, FileInfo 结构体
│   ├── wiki.go           Wiki 结构体
│   ├── config.go         Config 结构体 + 默认值
│   └── errors.go         领域错误定义
```

任务：

1. **Page 模型**
   - `Page`: ID(路径hash), Title, Content, Meta(Frontmatter), Sections, SourceFile, UpdatedAt
   - `Section`: ID, Level, Title, Content, Children (嵌套子段落, 树形结构)
   - `FileInfo`: Path, AbsPath, Size, Extension

2. **Wiki 模型**
   - `Wiki`: ID, Name, SourcePath, OutputPath, Config, PageCount, CreatedAt, UpdatedAt
   - 构造函数 `NewWiki()`

3. **Config 模型**
   - 完整 Config 结构体（按 data-model.md 最新版）
   - `DefaultConfig()` — Level 默认 2
   - `Merge(flags)` — flag 覆写合并
   - `Validate()` — Level 必须是 1/2/3

4. **错误定义**
   - 哨兵错误: ErrSourceNotFound, ErrPageNotFound, ErrEmptySource, ErrGenerateFailed
   - Error 结构体: Code, Message, Detail

**测试边界：**
- Config 默认值: Level=2, MaxSize=10MB
- Validate: Level=0 报错, Level=4 报错
- Page.ID 相同 path 生成相同 ID
- 错误类型断言正确

---

### M3: 扫描器

```
文件清单:
├── internal/core/scanner/
│   ├── scanner.go         Scanner 实现 (全文件扫描 + 二进制探测)
│   ├── rules.go           忽略规则解析
│   └── scanner_test.go
├── internal/core/scanner/testdata/
│   ├── sample.md
│   ├── sample.txt
│   ├── sample.json
│   ├── binary.bin         含 null 字节的测试文件
│   └── .baizeignore        测试忽略规则
```

任务：

1. **忽略规则引擎 (`rules.go`)**
   - 解析 `.baizeignore` (兼容 `.gitignore` 语法)
   - `Match(path)` → bool
   - 内置默认忽略: `.git/`, `node_modules/`, `vendor/`, `.DS_Store`, `Thumbs.db`

2. **二进制探测 (`scanner.go`)**
   - 读取前 512 字节
   - `isBinary(data)`: 含 null 字节 → true; 无效 UTF-8 → true
   - 使用 `bufio` 流式读取，避免大文件全读入内存

3. **文件扫描器 (`scanner.go`)**
   - `Scan(root string, cfg ScanConfig)` → `[]FileInfo`
   - 递归 `filepath.Walk`
   - 过滤链: 忽略规则 → 二进制探测 → 文件大小上限
   - 不设白名单: 所有非二进制文件都通过

**测试边界：**
- 空目录 → 空结果
- 纯目录结构（无文件）→ 空结果
- 混合文件: .md, .txt, .json, .go, .png, .mp3 → 只返回文本文件
- `.baizeignore` 正确排除
- >10MB 文件被过滤
- 符号链接被跳过
- 超大目录: context 超时

---

### M4: 解析器

```
文件清单:
├── internal/core/parser/
│   ├── parser.go          解析器主入口 + ParseBatch
│   ├── frontmatter.go     YAML frontmatter 提取
│   ├── markdown.go        Markdown 解析 (标题结构)
│   └── parser_test.go
├── internal/core/parser/testdata/
│   ├── basic.md
│   ├── with-frontmatter.md
│   ├── empty.md
│   ├── invalid-frontmatter.md
│   ├── sample.txt
│   └── sample.json
```

任务：

1. **Frontmatter 解析 (`frontmatter.go`)**
   - 提取 `---...---` 分隔的 YAML
   - 解析为 Frontmatter 结构体
   - 无 frontmatter → 空 Frontmatter，content 不变
   - 格式错误 → 记录警告，继续解析

2. **Markdown 解析 (`markdown.go`)**
   - 使用 `github.com/yuin/goldmark`
   - 提取标题 (`# ~ ######`) → Section 树
   - Section 的 Content 字段存放该节正文摘要（前 200 字）
   - 不解析 `[[link]]`（Phase 5 再做）

3. **非 .md 文件处理**
   - `.txt` / 无扩展名文本 → 文件名作为标题，全文作为 Content
   - `.json` / `.yaml` / `.toml` → 文件名作为标题，原文作为 Content
   - `.go` / `.py` / `.js` 等代码 → 文件名作为标题，原文作为 Content
   - 不尝试解析代码注释（Phase 3 再做）

4. **ParseBatch**
   - 并发解析（errgroup + 信号量限制并发数）
   - 单个文件解析失败不中断整体流程

**测试边界：**
- 标准 Markdown → Sections 树正确
- Frontmatter 完整 → 元数据正确
- 无 frontmatter → 正常解析
- 无效 frontmatter → 返回警告，不崩溃
- 空文件 → 空 Page（标题=文件名）
- `.txt`, `.json`, `.go` 文件 → 正确读取
- 二进制文件（误传入）→ 跳过并告警

---

### M5: 生成器

```
文件清单:
├── internal/core/generator/
│   ├── generator.go        生成编排主流程
│   ├── levels.go           Level 1/2/3 目录树构建
│   └── generator_test.go
├── internal/core/storage/
│   ├── writer.go           写入文件系统
│   └── reader.go           读取文件系统
```

任务：

**5.1 Level 路由 (`levels.go`)**

根据 `--level` 决定输出结构：

```
Level 1 (Flat):
  LevelBuilder.Flat(pages) → tree
  逻辑: 将所有页面按顶层路径分类
  输出: wiki/xxx.md (拍平, 前缀加父目录名)
  合并: 同分类的多个小文件合并为一个

Level 2 (Structured):
  LevelBuilder.Structured(pages) → tree
  逻辑: 取第一级路径作为子目录
  输出: wiki/category/page.md
  深度超过2层? 向上合并到第2层

Level 3 (Deep):
  LevelBuilder.Deep(pages) → tree
  逻辑: 完整目录树, 但限制 maxDepth=3
  输出: wiki/a/b/page.md
  深度超过3层? 截断到3层
```

**5.2 目录树结构**

```go
type PageRef struct {
    Title string `json:"title"` // 页面标题
    Path  string `json:"path"`  // 相对路径
}

type DirNode struct {
    Name     string     `json:"name"`     // 目录名
    Path     string     `json:"path"`     // 相对路径
    Pages    []PageRef  `json:"pages"`    // 该目录下的页面（轻量引用）
    Children []*DirNode `json:"children"` // 子目录
    Depth    int        `json:"depth"`    // 当前深度
}
```

**5.3 Generator (`generator.go`)**

```
Generator.Generate(wiki, pages) → error
  1. LevelBuilder.Build(pages, level) → DirNode 根节点
  2. 遍历 DirNode 树:
     for 每个目录节点:
       生成 _index.md (目录内页面列表 + 子目录列表)
       Storage.WritePage(path + "_index.md", content)
     for 每个 Page:
       Storage.WritePage(page.Path, page.Content)
  3. Storage.WriteMeta(wiki) → .baize/meta.json
```

**5.4 Index 模板 (`_index.md`)**

```markdown
# {目录名}

> 共 {N} 个页面

## 页面

- [{标题}]({相对路径})
- ...

## 子目录

- [{目录名}]({路径}/)
- ...
```

Phase 1 不做模板自定义，内置即可。

**5.5 Storage (`writer.go`)**

- `WritePage(path, content)`: 确保父目录存在 → 写入 `.md`
- `WriteMeta(wiki)`: 写入 `.baize/meta.json` 为 JSON
- `ReadMeta(wikiDir)`: 从 `.baize/meta.json` 读取
- 原子写入: 先写 `.tmp` 文件再 `os.Rename`

**测试边界：**
- 空 pages 列表 → 仅根 _index.md
- Level 1: 所有页面在根目录，前缀正确
- Level 2: 一层子目录，结构正确
- Level 3: 完整深度（max 3 层）
- 源目录深度 < level → 按实际深度输出
- _index.md 内容正确（标题、链接不空）
- 原子写入: 中间崩溃不产生残缺文件
- 重复 build → 幂等

---

### M6: CLI 集成

```
文件清单:
├── cmd/baize-wiki/main.go      完整实现
├── internal/app/build.go       build 用例编排
├── internal/app/info.go        info 用例编排
├── internal/app/app.go         共享工具
├── internal/config/config.go   完整实现 + flag 绑定
├── testdata/                   E2E 测试数据
│   ├── source/
│   │   ├── 快速开始.md           (带 frontmatter)
│   │   ├── 架构设计.md
│   │   ├── guide/
│   │   │   ├── 安装指南.md
│   │   │   └── 使用教程.md
│   │   ├── sample.txt
│   │   └── secret.key           (二进制文件, 应被跳过)
│   └── .baizeignore
└── e2e_test.go                 端到端测试
```

任务：

1. **config 加载完整实现**
   - viper 读取 `baize.yaml`
   - 环境变量覆写: `BAIZE_*`
   - Flag 绑定: `--level`, `--output`, `--config`, `--verbose`, `--json`

2. **`baize-wiki init [source]`**
   - 在当前目录生成 `baize.yaml` 模板
   - `--output`, `--name`, `--level` 参数
   - `--force` 覆盖已有

3. **`baize-wiki build [source]`**
   - 串联流程: Config → Scanner → Parser → Generator → Storage
   - `--level` 控制输出结构
   - `--json` 输出结构化结果（供 Agent 解析）
   - 进度输出: 文件数 / 页面数 / 耗时

4. **`baize-wiki info`**
   - `--tree`: 读取 Wiki 目录，显示树形结构
   - `--stats`: 从 `.baize/meta.json` 读取
   - `--json`: 统一 JSON 输出

5. **端到端测试**
   - 准备 testdata 源目录（含 .md, .txt, .go, .bin）
   - 分别运行 `--level 1/2/3` 验证输出结构
   - 验证二进制文件被跳过
   - 验证 `_index.md` 正确生成
   - 验证 JSON 输出可解析

---

### M7: 收尾

1. README: 安装、快速开始、Level 说明、MCP 配置预览
2. `examples/basic/` 示例项目
3. `.baizeignore` 默认文件
4. GitHub Actions: CI (lint + test + build)
5. **Benchmark 测试**: 记录关键路径性能基线
   - Scanner: 1千 / 1万 / 10万文件扫描耗时
   - Parser: 1MB Markdown 文件解析内存占用
   - Generator: 100 / 1000 页面输出速度
   - 基线数据写入 `docs/benchmarks.md`
6. 打 tag `v0.1.0-alpha`

---

## 5. 依赖清单

```
# 核心依赖
github.com/spf13/cobra          # CLI 框架
github.com/spf13/viper           # 配置读取
github.com/yuin/goldmark         # Markdown 解析 (取标题结构)
gopkg.in/yaml.v3                 # YAML frontmatter 解析

# 开发依赖
github.com/stretchr/testify      # 测试断言
golang.org/x/sync/errgroup       # 并发控制
```

---

## 6. 代码规范

- 遵循 Go 标准项目布局
- `internal/` 不对外暴露，`pkg/` 有兼容性保证
- 领域错误用 sentinel errors，系统错误用 `fmt.Errorf + %w`
- 日志: `slog` 标准库，通过 context 传递
- 交互式输出（进度条等）写入 stderr，stdout 仅输出 `--json` 结果
- 测试: 表格驱动 + testify/assert
- 文件名和目录名不做假定：Windows 路径用 `filepath.Join`

---

## 7. 风险与缓解

| 风险 | 影响 | 缓解措施 |
|------|------|---------|
| 二进制探测误判（如含 null 的文本文件） | 漏扫 | 增加可配置的白名单扩展名排除 |
| 超大目录扫描性能 | 慢 | filepath.Walk 本身高效，可后续加并发 |
| Level 聚合逻辑复杂度 | 返工 | Phase 1 用最简单的"基于目录结构"方案 |
| Windows 路径兼容 | 跨平台 bug | 全用 `filepath` / `path/filepath`，CI 加 Windows runner |
| 非 UTF-8 编码文件 | 乱码 | 尝试检测 BOM，否则按 UTF-8 处理并告警 |
