# Baize Wiki — 测试方案

> 目的: 验证 v0.1.0-alpha → Beta 全链路功能正常

---

## 1. 测试样本

从 `F:/Project/` 下抽取 C# 项目源码作为输入。每个项目保持相对路径。

```bash
bash scripts/extract-cs-samples.sh
```

输出到 `testdata/cs-samples/<项目名>/<路径>/xxx.cs`

已生成的 1000 个 lorem ipsum 文件也随时可用（`testdata/large-project/`）。

---

## 2. 测试层级

| 层级 | 命令 | 预期 |
|:----:|------|------|
| 🥉 低 | `build --level 1` | 平面目录，原文输出，无编目 |
| 🥈 中 | `build --level 2 --catalog-level 2` | 结构化目录，每页有摘要+关键词 |
| 🥇 高 | `build --level 3 --catalog-level 2` | 深度目录 + 编目 + 知识图谱 |

---

## 3. 测试场景

### 3.1 Local 编目（无 LLM）

```bash
baize-wiki build testdata/large-project --output ./wiki-l3 --level 3 --catalog-level 2
```

- 速度预期: 1000 页 / ~3 秒
- 产出: 摘要（首段提取）+ 关键词 + 知识图谱
- 验证: `info --stats`、`search`、`wiki-test/.baize/graph.json`

### 3.2 Remote 编目（Qwen3.5-9B）

需要 LM Studio 运行中（`localhost:1234`），且代码支持 `catalog_backend: remote` 配置。

```bash
# 编目时每页调一次 LLM，生成高质量摘要
baize-wiki build ./src --catalog-level 2 --catalog-backend remote
```

- 速度预期: 每页约 1-3 秒（取决于模型速度）
- 产出: LLM 生成的摘要 + 关键词 + 实体
- 验证: 摘要质量明显优于首段提取

### 3.3 增量构建

```bash
# 第一次全量
baize-wiki build ./src --output ./wiki --catalog-level 2

# 第二次（只变了一两个文件）
baize-wiki build ./src --output ./wiki --catalog-level 2 --incremental
```

- 期望: 未变更文件跳过编目，只处理变动文件

### 3.4 向量搜索

1. 先 build 生成向量索引（`vector.mode: local` 或 `remote`）
2. 搜索: `baize-wiki search "关键词" --semantic`

### 3.5 MCP 工具

启动 MCP server 后，用任意 MCP 客户端连接:

| 工具 | 功能 | 验证方式 |
|:----|------|---------|
| `wiki_build` | 构建 Wiki | 传 source/output/catalog_level |
| `wiki_read` | 读页面 | depth=0/1/2, section=id |
| `wiki_search` | 搜索 | query + max_tokens |
| `wiki_graph` | 查知识图谱 | operation=entities/relations/layers |
| `wiki_list` | 目录浏览 | dir/depth |

---

## 4. 验证清单

- [ ] 编译通过: `go build ./cmd/baize-wiki`
- [ ] 全部测试: `go test ./...`
- [ ] Level 1/2/3 输出结构正确
- [ ] 编目后每页有摘要+关键词
- [ ] 知识图谱 graph.json 有节点+层
- [ ] 搜索结果能命中
- [ ] MCP 7 个工具都注册
- [ ] --incremental 不崩溃
- [ ] --scan-all 扫 .cs 文件

---

## 5. 当前状态

| 项目 | 状态 |
|:-----|:----:|
| 测试样本（1000 lorem） | ✅ 已生成 |
| 测试样本（C# 源码） | 🚧 `scripts/extract-cs-samples.sh` 需运行 |
| Local 编目 | ✅ 通过 |
| Remote 编目 | 🚧 需要 `--catalog-backend` 参数 + LM Studio |
| 知识图谱 | ✅ 通过 |
| 增量模式 | ⚠️ 基础实现，跨运行需持久化 |
| 向量搜索 | ✅ LocalEmbedder 通过 / LM Studio 待验证 |
| MCP 工具 | ✅ 7 个注册 |
