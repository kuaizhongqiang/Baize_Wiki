# 实施计划

> 目标: Level 2/3 真正可用的编目流水线，产出对 AI Agent 有价值的"处理+生成"内容。

---

## 核心问题

现在 Local 编目只是"前 200 字搬运"，没有实际价值。需要让 LLM 真正参与处理：

1. **概念目录** — 理解项目后重新组织，而非 Mirror 文件路径
2. **精准摘要** — LLM 读代码后写 50 token 压缩
3. **实体关系** — 跨文件识别类/接口/依赖 → graph.json 有边

---

## 实施步骤

### Step 1: 配置文件化（1 天）

让 catalog/vector 后端完全由 `baize.yaml` 控制，不再硬编码。

```yaml
catalog:
  profile: 性价比           # 急速 | 性价比 | 本地
  level: 2
  backend: remote
  endpoint: http://localhost:1234/v1
  model: qwen/qwen3.5-9b

vector:
  profile: 性价比
  backend: remote
  endpoint: http://localhost:1234/v1
  model: text-embedding-baai-bge-m3-568m:2
```

CLI 参数: `--profile`, `--catalog-backend`, `--catalog-endpoint`, `--catalog-model`

### Step 2: Remote 编目跑通（1 天）

Qwen3.5 逐页处理源码 → 产出摘要+关键词+实体。

验证输出质量，确保:
- 摘要不是前 200 字，而是 LLM 理解后的压缩
- 实体不是标题凑数，而是真实 class/interface/module
- 多页之间有跨页关系

### Step 3: 概念目录（2 天）

不是 Mirror 文件路径，而是 LLM 理解后的概念重组:

```
当前:  api/chapter-1/page-1.md   ← 文件路径
目标:  架构设计/核心模块.md      ← 概念分组
```

实现方式: 编目时多一轮 LLM 调用，分析所有页面摘要 → 输出概念分类。

### Step 4: 输出模板（1 天）

每页输出格式从"原文"改为"结构化 markdown":

```markdown
# 标题

**摘要:** LLM 写的 50 字压缩

**关键词:** keyword1, keyword2

**实体:** ClassA(class), InterfaceB(interface)

## 原文
...
```

### Step 5: Level 3 图谱补完（2 天）

graph.json 的边不为空，支持:
- 跨页依赖关系（A.cs 定义了 → B.cs 使用了）
- 架构分层自动识别
- wiki_graph MCP 工具返回有效数据

---

## 优先级

| 步骤 | 价值 | 工作量 | 优先 |
|:----:|:----:|:------:|:----:|
| Step 1 配置文件化 | 高（解耦所有后续工作） | 小 | 🥇 1 |
| Step 2 Remote 编目 | 高（首次看到 LLM 产出） | 小 | 🥇 2 |
| Step 3 概念目录 | 高（核心价值） | 中 | 🥇 3 |
| Step 4 输出模板 | 中（改善展示） | 小 | 🥈 4 |
| Step 5 Level 3 图谱 | 高（知识网络） | 中 | 🥈 5 |

---

## 设计原则

1. **配置驱动，不写死** — 后端选择全由 baize.yaml + CLI 控制
2. **渐进验证** — 每一步都能独立看到产出，不阻塞
3. **兜底降级** — 无 LLM 时首段提取依然可用
