# Baize Wiki (白泽维基)

> 面向 AI Agent 的 Wiki 生成与使用工具

[![Go Version](https://img.shields.io/badge/Go-1.22+-blue)](https://go.dev/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Status](https://img.shields.io/badge/status-alpha-orange)]()

**Baize Wiki** 是一个专为 AI Agent 设计的 Wiki 工具。它能扫描指定路径下的所有文档，自动解析内容，并按可配置的复杂度级别（1/2/3）生成结构化的 Wiki 文件。

### 核心特性

- **能扫尽扫** — 自动探测并跳过二进制文件，其余全部纳入 Wiki
- **三级输出** — `--level 1/2/3` 控制输出结构，从平面文件到深度目录
- **AI Agent 优先** — 提供 CLI 和 MCP Server 两种集成方式
- **单二进制分发** — 零运行时依赖，即下即用

---

## 状态

当前处于 **Phase 1 (MVP) 开发阶段**，核心链路即将完成：

```
源目录 → 扫描器 → 解析器 → 生成器 → Wiki 输出 (MD 文件)
```

后续规划：MCP Server → 全文检索 → 向量化 → 交叉链接

详见 [docs/architecture.md](docs/architecture.md) 和 [docs/phase-1-plan.md](docs/phase-1-plan.md)。

---

## 快速开始

```bash
# 下载二进制 (TODO: 发布后可用)
# 或本地编译
go build -o bin/baize-wiki ./cmd/baize-wiki

# 初始化配置
baize-wiki init ./docs --name "我的 Wiki"

# 构建 Wiki (Level 2 结构化)
baize-wiki build ./docs --output ./wiki --level 2

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
├── cmd/baize-wiki/     # CLI 入口
├── internal/
│   ├── core/           # 领域模型与核心逻辑
│   │   ├── model/      # 数据结构 (Page, Wiki, Config)
│   │   ├── scanner/    # 文件扫描 + 二进制探测
│   │   ├── parser/     # Markdown / 纯文本解析
│   │   ├── generator/  # Wiki 生成 (Level 1/2/3)
│   │   └── storage/    # 文件读写
│   ├── app/            # 应用层用例
│   ├── config/         # 配置加载
│   └── mcp/            # MCP 协议 (Phase 2)
├── pkg/baize/          # 公开 API
├── docs/               # 设计文档
├── testdata/           # 测试数据
└── examples/           # 使用示例
```

---

## 集成方式

### CLI 直接使用

```bash
baize-wiki build ./docs --level 2 --output ./wiki --json
```

### MCP Server (Phase 2)

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

### Docker (Phase 2)

```bash
docker run -v ./docs:/docs baize-wiki build /docs --output /wiki
```

---

## 设计文档

| 文档 | 说明 |
|------|------|
| [架构总览](docs/architecture.md) | 整体架构、Go 包结构、决策记录 |
| [数据模型](docs/data-model.md) | 核心结构体、配置格式、接口合约 |
| [CLI 规范](docs/cli-spec.md) | 命令定义、参数、退出码 |
| [MCP 规范](docs/mcp-spec.md) | MCP 工具定义、协议细节 |
| [Phase 1 计划](docs/phase-1-plan.md) | 实施里程碑、任务分解 |

---

## License

MIT License © 2026 kuaizhongqiang
