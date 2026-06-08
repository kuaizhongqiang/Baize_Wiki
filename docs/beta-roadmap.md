# Baize Wiki — Beta 路线图

> 当前: v0.1.0-alpha
> 目标: v1.0.0-beta
> 状态: 框架规划中

---

## Alpha 遗留补完

v0.1.0-alpha 文档中规划但未交付的项目，Beta 版一并补齐：

- **G1 — Dockerfile + Docker 集成测试** ✅ 已完成
- **G2 — 真实语义 Embedding API** ✅ RemoteEmbedder + LM Studio 配置完成
- **G3 — MCP Resources / Prompts** ✅ 4 个协议方法可用
- **G4 — 大项目性能优化（1000+ 文件）** ✅ 测试数据已生成 (~925 文件)

> Watch 文件监听模式已正式移除（代码变更由 CodeGraph 负责，非 Baize Wiki 职责）。

## 战略方向

### D — AI 增强与 Token 优化 ✅ (已完成)

**核心模型：** 三层递归输出，每层是对上一层的更高维度总结。

- 🥉 **Level 1 — 机械层**：目录索引 + 原文
- 🥈 **Level 2 — 摘要层**：逐页摘要、关键词、实体提取
- 🥇 **Level 3 — 图谱层**：跨页关系发现、架构分层、graph.json、MCP wiki_graph

**已实现：** `--catalog-level 2` 编目流水线、Local/Remote 双后端、敏感过滤、增量检测、`wiki_read` 分层交付、智能搜索片段、`wiki_graph` 查询。

详见 [docs/directions/d-ai-enhancement.md](docs/directions/d-ai-enhancement.md)

---

### C — 模板系统 + 主题定制 ⏸️ (Beta 后)

自定义 `_index.md` 模板，HTML 主题渲染。

**价值：** 输出美化，适应不同场景

---

### E — 插件系统 ⏸️ (Beta 后)

自定义 Parser / Generator 插件，第三方扩展能力。

**价值：** 生态基础

---

### F — 导入/导出/集成 ⏸️ (Beta 后)

Obsidian / Notion / 飞书等知识库互通。

**价值：** 生态扩展（待研究）
