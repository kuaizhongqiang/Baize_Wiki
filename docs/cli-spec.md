# Baize Wiki — CLI 规范

> 版本: v1 (草案)
> 状态: 待审计
> 框架: cobra

## 1. 命令总览

```
baize-wiki — AI Agent 的 Wiki 工具

用法:
  baize-wiki [command] [flags]

可用命令:
  init      初始化一个 Wiki 项目（生成 baize.yaml）
  build     从源码目录构建/更新 Wiki
  info      查看 Wiki 或页面信息
  search    搜索 Wiki 内容
  mcp       启动 MCP Server 模式
  help      查看帮助

全局标志:
  -c, --config string   配置文件路径（默认 ./baize.yaml）
  -v, --verbose         输出详细日志
      --version         显示版本信息
```

### 输出格式说明

默认输出是人类友好的控制台文本（含 Emoji 状态标记）。所有命令支持 `--json` / `-j` 标志切换为 JSON 输出，供 AI Agent 解析。

设计考量：
- CLI 模式主要使用场景是**开发测试和调试**，人类可读输出更实用
- AI Agent 通过 CLI 使用时应附加 `--json` 标志获取结构化结果
- AI Agent 的**主要集成方式**是 MCP Server（Phase 2），通过 JSON-RPC 通信，格式问题自然解决

---

## 2. 子命令详情

### 2.1 `baize-wiki init`

初始化一个 Wiki 项目。在目标目录生成 `baize.yaml` 配置模板。

```
用法:
  baize-wiki init [source] [flags]

参数:
  source    源文件/目录路径（默认 ./）

标志:
  -o, --output string   输出目录（默认 ./wiki）
  -n, --name string     Wiki 名称（默认取目录名）
  -f, --force           覆盖已有配置
```

**示例：**
```bash
baize-wiki init ./docs --output ./wiki --name "项目文档"
# → 生成 ./baize.yaml（如果已存在则报错）
# → 提示 "运行 'baize-wiki build' 生成 Wiki"

baize-wiki init ./docs -f
# → 强制覆盖已有 baize.yaml
```

**生成 baize.yaml 模板：**
```yaml
# 由 Baize Wiki 自动生成
name: "项目文档"
scan:
  paths:
    - ./docs
  exclude: []
  max_size: 10485760
output:
  dir: "./wiki"
  level: 2
  clean: false
organize:
  by: directory
features:
  draft: false
```

**退出码：** `0` 成功, `1` 配置已存在且未传 `-f`, `2` 路径错误

---

### 2.2 `baize-wiki build`

核心命令：扫描源目录 → 解析文档 → 生成 Wiki。

```
用法:
  baize-wiki build [source] [flags]

参数:
  source    源文件/目录路径（覆盖 baize.yaml 中的 scan.paths）

标志:
  -o, --output string     输出目录（覆盖 baize.yaml output.dir）
  -l, --level int         输出复杂度: 1=平面, 2=结构化, 3=深度 (默认 2)
  -w, --watch             监听文件变更，自动重新构建 (Phase 2+)
      --draft             包含标记为 draft 的页面
  -q, --quiet             静默模式（仅输出错误和摘要）
```

**示例：**
```bash
baize-wiki build
# → 读取 baize.yaml，扫描 ./docs，输出到 ./wiki

baize-wiki build ./src/content -o ./public
# → 覆盖配置中的源路径和输出目录

baize-wiki build --level 3
# → 以 Level 3（深度）结构生成 Wiki
```

**构建流程：**
```
1. 加载配置（baize.yaml + flag 覆写）
2. 验证源路径存在
3. Scanner 扫描源目录 → 文件列表（跳过二进制）
4. Parser 解析每个文件 → Page 列表
5. Generator 按 --level 生成目录树 + _index.md
6. Storage 写入页面文件和 meta.json
```

**输出示例：**
```
✓ 扫描完成: 找到 24 个文件 (跳过 3 个二进制, 忽略 5 个)
✓ 解析完成: 24/24 成功
✓ 生成完成: 输出到 ./wiki (24 页面, 8 目录, Level 2)
```

**退出码：** `0` 成功, `1` 构建失败, `2` 源路径无效, `3` 无有效文件

---

### 2.3 `baize-wiki info`

查看 Wiki 或页面信息。

```
用法:
  baize-wiki info [path] [flags]

参数:
  path     Wiki 目录（默认 ./wiki）或页面路径

标志:
  -t, --tree       以树形显示 Wiki 目录结构
  -s, --stats      显示统计信息（页面数、标签数、链接数）
  -j, --json       JSON 格式输出（供 Agent 解析）
```

**示例：**
```bash
baize-wiki info --tree
# wiki/
# ├── _index.md
# ├── 架构设计/
# │   ├── _index.md
# │   ├── 数据模型.md
# │   └── 接口规范.md
# └── 开发指南/
#     ├── _index.md
#     └── 快速开始.md

baize-wiki info --stats --json
# {"page_count":24,"total_words":15800,...}

baize-wiki info ./wiki/architecture/data-model.md
# Title: 数据模型规范
# Tags: [data-model, api]
# Links: 3
# Backlinks: 5
# Updated: 2026-06-05 14:30
```

**退出码：** `0` 成功, `1` Wiki 不存在, `2` 页面不存在

---

### 2.4 `baize-wiki search`

搜索 Wiki 内容（需要 Phase 3 全文索引）。

```
用法:
  baize-wiki search <query> [flags]

参数:
  query    搜索关键词

标志:
  -w, --wiki string     Wiki 目录（默认 ./wiki）
  -l, --limit int       返回结果数（默认 10）
  -t, --tags strings    按标签筛选
  -j, --json            JSON 格式输出
```

**示例：**
```bash
baize-wiki search "数据模型" --limit 5
# 找到 3 个结果:
# 1. 架构设计/数据模型.md (匹配: 标题)
#    ...定义了 Page、Wiki 等核心数据结构...
# 2. 开发指南/快速开始.md (匹配: 内容)
#   ...运行 build 命令生成数据模型...

baize-wiki search "scanner" --tags go -j
# [{"path":"...", "title":"...", "snippet":"...", "score":0.95}]
```

**退出码：** `0` 找到结果, `1` 无结果 / 未建索引, `2` 查询为空

---

### 2.5 `baize-wiki mcp`

启动 MCP Server 模式（详见 [mcp-spec.md](mcp-spec.md)）。

```
用法:
  baize-wiki mcp [wiki-dir] [flags]

参数:
  wiki-dir    Wiki 目录路径（默认 ./wiki）

标志:
  -t, --transport string   传输方式: stdio | tcp（默认 stdio）
  -a, --addr string        TCP 监听地址（默认 :8080）
```

**示例：**
```bash
baize-wiki mcp ./wiki
# → 启动 stdio MCP Server

baize-wiki mcp -t tcp -a :8080 ./wiki
# → 启动 TCP MCP Server，监听 :8080
```

---

### 2.6 `baize-wiki help`

```
baize-wiki help [command]
```

---

## 3. 环境变量

| 变量 | 作用 | 默认 |
|------|------|------|
| `BAIZE_CONFIG` | 配置文件路径 | `./baize.yaml` |
| `BAIZE_WIKI_DIR` | Wiki 目录（作为 flag 默认值） | `./wiki` |
| `BAIZE_LOG_LEVEL` | 日志级别 | `info` |
| `BAIZE_NO_COLOR` | 禁用彩色输出 | `false` |

---

## 4. 返回格式 (JSON 模式)

所有命令支持 `--json` / `-j` 标志，输出 JSON 供 Agent 解析。

```json
// baize-wiki build --json 输出示例
{
  "success": true,
  "duration_ms": 2340,
  "summary": {
    "total_files": 24,
    "parsed": 24,
    "pages": 24,
    "directories": 8
  },
  "errors": [],
  "warnings": []
}
```

```json
// 错误输出
{
  "success": false,
  "error": {
    "code": "ERR_SOURCE_NOT_FOUND",
    "message": "源路径不存在",
    "detail": "/path/to/docs"
  }
}
```

---

## 5. 版本与升级

```bash
baize-wiki --version
# baize-wiki version 0.1.0 (commit abc1234, built 2026-06-05)
```

版本号遵循 [SemVer](https://semver.org/)：
- Phase 1: `v0.1.0-alpha` — MVP (首次发布)
- Phase 2: `v0.2.x` — MCP 集成
- Phase 3: `v0.3.x` — 检索
- Phase 4+: `v0.4.x` — 向量化
- 稳定版: `v1.0.0`
