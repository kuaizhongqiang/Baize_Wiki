## Beta 发布

### 新增功能

- **入库编目流水线** — `--catalog-level 2/3`，支持 Local（首段提取）和 Remote（Qwen3.5/DeepSeek）双后端
- **配置文件化** — `--profile speed|balanced|local` + `--catalog-backend` + 环境变量全覆盖
- **概念目录** — `--catalog-level 3` 时自动按概念分组，非 Mirror 文件路径
- **知识图谱** — .baize/graph.json + wiki_graph MCP 工具
- **智能输出模板** — 每页结构化 markdown，摘要+关键词在前
- **增量构建** — `--incremental` 模式
- **MCP Resources/Prompts** — 4 个新协议方法
- **Dockerfile** — 多阶段构建
- **真实 Embedding** — 支持 LM Studio bge-m3 / OpenAI API

### 编目级别对比

| --catalog-level | 产出 | 需要 LLM |
|:---:|:-----|:--------:|
| 0 | 无编目 | 否 |
| 2 | 每页摘要+关键词+实体 | 可选 |
| 3 | 摘要+概念分类+知识图谱 | 可选 |

### Qwen3.5 编目输出样例

```markdown
> **FirstPersonController 是一个基于 Unity Input System 和 CharacterController 的
第一人称控制器，负责处理玩家移动、冲刺、跳跃、重力及旋转逻辑。**

`第一人称控制器` `玩家移动` `跳跃重力` `Cinemachine` `地面检测`
```

### 配置文件示例

```yaml
profile: balanced
catalog:
  level: 2
  backend: remote
  endpoint: http://localhost:1234/v1
  model: qwen/qwen3.5-9b
vector:
  mode: remote
  endpoint: http://localhost:1234/v1
  model: text-embedding-baai-bge-m3-568m:2
```

### 下载

提供 linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, windows/amd64 二进制。
