# Baize Wiki — MCP 协议规范

> 版本: v1 (草案)
> 状态: 待审计

## 1. 概述

Baize Wiki 提供 **MCP (Model Context Protocol)** 接口，让 AI Agent 能动态地查询、浏览和操作 Wiki 内容。

### 协议版本
- MCP 协议版本: 2025-03-26
- 传输方式: stdio（默认） / TCP（可配置）
- 消息格式: JSON-RPC 2.0

AI Agent（如 Claude Code、Cline）通过以下方式集成：

**方式 1 — stdio（推荐，MCP 客户端自动拉起）：**
```json
{
  "mcpServers": {
    "baize-wiki": {
      "command": "/usr/local/bin/baize-wiki",
      "args": ["mcp", "/path/to/wiki"],
      "env": {}
    }
  }
}
```

**方式 2 — TCP（适用于 Docker 部署）：**
```json
{
  "mcpServers": {
    "baize-wiki": {
      "url": "http://localhost:8080/mcp"
    }
  }
}
```

---

## 2. Tools 定义

### 2.1 wiki_build

构建/更新 Wiki。

```json
{
  "name": "wiki_build",
  "description": "从源目录构建或更新 Wiki。扫描指定路径下的文档并生成结构化的 Wiki 输出。",
  "inputSchema": {
    "type": "object",
    "properties": {
      "source": {
        "type": "string",
        "description": "源文件或目录路径（默认使用 baize.yaml 配置）"
      },
      "output": {
        "type": "string",
        "description": "Wiki 输出目录（默认使用 baize.yaml 配置）"
      },
      "config": {
        "type": "string",
        "description": "配置文件路径（默认 ./baize.yaml）"
      },
      "level": {
        "type": "integer",
        "description": "输出复杂度: 1 | 2 | 3（默认使用 baize.yaml 配置）"
      }
    },
    "required": []
  }
```

**输出示例：**
```json
{
  "content": [{
    "type": "text",
    "text": "{\"success\":true,\"duration_ms\":2340,\"summary\":{\"pages\":24,\"directories\":8,\"links\":47}}"
  }]
}
```

---

### 2.2 wiki_search

搜索 Wiki 内容。

```json
{
  "name": "wiki_search",
  "description": "在 Wiki 中搜索相关内容。支持关键词搜索和标签筛选。返回匹配的页面列表，含上下文摘要。",
  "inputSchema": {
    "type": "object",
    "properties": {
      "query": {
        "type": "string",
        "description": "搜索关键词"
      },
      "tags": {
        "type": "array",
        "items": {"type": "string"},
        "description": "按标签筛选（可选）"
      },
      "limit": {
        "type": "integer",
        "description": "最大返回数（默认 10）",
        "default": 10
      },
      "include_content": {
        "type": "boolean",
        "description": "是否返回全文内容（默认 false，仅返回摘要）",
        "default": false
      }
    },
    "required": ["query"]
  }
}
```

**输出示例：**
```json
{
  "content": [{
    "type": "text",
    "text": "[{\"path\":\"架构设计/数据模型.md\",\"title\":\"数据模型规范\",\"score\":0.95,\"snippet\":\"...定义了 Page、Wiki 等核心数据结构...\",\"tags\":[\"data-model\",\"api\"]}]"
  }]
}
```

---

### 2.3 wiki_read

读取 Wiki 页面内容。

```json
{
  "name": "wiki_read",
  "description": "读取 Wiki 中一个页面的完整 Markdown 内容。",
  "inputSchema": {
    "type": "object",
    "properties": {
      "path": {
        "type": "string",
        "description": "页面路径（相对于 Wiki 根目录，如 \"架构设计/数据模型.md\"）"
      },
      "format": {
        "type": "string",
        "enum": ["markdown", "html", "text"],
        "description": "返回格式（默认 markdown）",
        "default": "markdown"
      }
    },
    "required": ["path"]
  }
}
```

**输出示例：**
```json
{
  "content": [{
    "type": "text",
    "text": "# 数据模型规范\n\n> 版本: v1\n\n## 核心数据结构\n\n### Page\n\n```go\n...```"
  }]
}
```

---

### 2.4 wiki_list

列出 Wiki 目录结构。

```json
{
  "name": "wiki_list",
  "description": "浏览 Wiki 的目录结构。可查看根目录或指定子目录下的页面列表。",
  "inputSchema": {
    "type": "object",
    "properties": {
      "dir": {
        "type": "string",
        "description": "目录路径（相对 Wiki 根，默认 \"\" 即根目录）"
      },
      "depth": {
        "type": "integer",
        "description": "递归深度（默认 1，-1 为无限）",
        "default": 1
      },
      "include_pages": {
        "type": "boolean",
        "description": "是否包含页面列表（默认 true）",
        "default": true
      }
    },
    "required": []
  }
}
```

**输出示例：**
```json
{
  "content": [{
    "type": "text",
    "text": "{\"path\":\"\",\"type\":\"directory\",\"children\":[{\"name\":\"架构设计\",\"type\":\"directory\",\"children\":[{\"name\":\"数据模型.md\",\"type\":\"page\",\"title\":\"数据模型规范\"},{\"name\":\"接口规范.md\",\"type\":\"page\",\"title\":\"接口规范\"}]},{\"name\":\"开发指南\",\"type\":\"directory\",\"children\":[{\"name\":\"快速开始.md\",\"type\":\"page\",\"title\":\"快速开始\"}]}]}"
  }]
}
```

---

### 2.5 wiki_add

添加/更新 Wiki 页面。

```json
{
  "name": "wiki_add",
  "description": "向 Wiki 中添加新页面或更新已有页面。AI Agent 可以通过此工具将新知识写入 Wiki。",
  "inputSchema": {
    "type": "object",
    "properties": {
      "path": {
        "type": "string",
        "description": "页面路径（相对于 Wiki 根，如 \"开发指南/调试技巧.md\"）"
      },
      "content": {
        "type": "string",
        "description": "Markdown 格式的页面内容"
      },
      "tags": {
        "type": "array",
        "items": {"type": "string"},
        "description": "标签列表"
      },
      "overwrite": {
        "type": "boolean",
        "description": "是否覆盖已有页面（默认 false，false 时若存在则报错）",
        "default": false
      }
    },
    "required": ["path", "content"]
  }
}
```

---

### 2.6 wiki_stats

获取 Wiki 统计信息。

```json
{
  "name": "wiki_stats",
  "description": "获取 Wiki 的整体统计信息，包括页面数、标签数、交叉引用数、最后更新时间等。",
  "inputSchema": {
    "type": "object",
    "properties": {},
    "required": []
  }
}
```

**输出示例：**
```json
{
  "content": [{
    "type": "text",
    "text": "{\"name\":\"我的 Wiki\",\"page_count\":24,\"directory_count\":8,\"total_links\":127,\"dangling_links\":3,\"tags\":[\"architecture\",\"go\",\"api\"],\"updated_at\":\"2026-06-05T14:30:00Z\",\"wiki_path\":\"/path/to/wiki\"}"
  }]
}
```

---

## 3. 错误处理

所有工具返回 MCP 标准错误格式：

```json
{
  "isError": true,
  "content": [{
    "type": "text",
    "text": "{\"code\":\"ERR_PAGE_NOT_FOUND\",\"message\":\"页面不存在\"}"
  }]
}
```

| 错误码 | HTTP 类比 | 说明 |
|--------|----------|------|
| `ERR_WIKI_NOT_FOUND` | 404 | Wiki 目录不存在或未构建 |
| `ERR_PAGE_NOT_FOUND` | 404 | 指定路径的页面不存在 |
| `ERR_SOURCE_NOT_FOUND` | 400 | 构建时源路径无效 |
| `ERR_INVALID_PATH` | 400 | 路径包含非法字符 |
| `ERR_PAGE_EXISTS` | 409 | 页面已存在且 overwrite=false |
| `ERR_BUILD_FAILED` | 500 | Wiki 构建过程失败 |
| `ERR_INTERNAL` | 500 | 内部错误 |

---

## 4. 标准化 Prompts (可选)

MCP 规范支持定义 Prompts，这里列出建议预定义的 prompt 模板，方便 Agent 快速上手使用该 Wiki：

```
prompt: wiki-guide
description: 如何有效使用 Baize Wiki 的指南
messages:
  - role: system
    content: >
      你正在使用一个 Baize Wiki。你可以：
      1. 用 wiki_search 搜索知识
      2. 用 wiki_list 浏览目录
      3. 用 wiki_read 阅读具体页面
      4. 用 wiki_add 向 Wiki 写入新知识
      5. 用 wiki_build 从源码重新生成 Wiki
      6. 用 wiki_stats 查看整体概览
      
      Wiki 内容以目录结构组织，每篇页面是 Markdown 文件。
      当你获得新知识时，应该用 wiki_add 将其写入 Wiki 以便后续使用。
```

---

## 5. 传输协议详情

### 5.1 stdio 传输

- 启动：`baize-wiki mcp <wiki-dir>`
- 从 stdin 读取 JSON-RPC 请求
- 写入 stdout 返回 JSON-RPC 响应
- stderr 输出日志，不影响数据流
- 日志级别由 `BAIZE_LOG_LEVEL` 控制

### 5.2 TCP 传输（容器场景）

- 启动：`baize-wiki mcp --transport tcp --addr :8080 <wiki-dir>`
- 端点：`POST /mcp`
- Content-Type: `application/json`
- 请求体：JSON-RPC 2.0 请求
- 响应体：JSON-RPC 2.0 响应

---

## 6. 资源定义 (Resources)

MCP 规范支持 Resources（只读数据源）。Baize Wiki 可暴露：

| Resource URI | 说明 | MIME |
|-------------|------|------|
| `baize://wiki/meta` | Wiki 元信息 | `application/json` |
| `baize://wiki/tree` | 目录树 | `application/json` |
| `baize://wiki/pages/{path}` | 特定页面内容 | `text/markdown` |
| `baize://wiki/stats` | 统计信息 | `application/json` |

Phase 2 实现，当前先聚焦 Tools。
