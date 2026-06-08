# Baize Wiki — Beta 路线图

> 当前: v0.1.0-alpha
> 目标: v1.0.0-beta
> 状态: 框架规划中

---

## 战略方向

### C — 模板系统 + 主题定制

自定义 `_index.md` 模板，HTML 主题渲染。

**价值：** 输出美化，适应不同场景

---

### D — AI 增强与 Token 优化 ⬅️ (当前焦点)

**核心模型：** 三层递归输出，每层是对上一层的更高维度总结。

- 🥉 **Level 1 — 机械层**：目录索引 + 原文，不做理解
- 🥈 **Level 2 — 摘要层**：每页独立摘要 + 关键词
- 🥇 **Level 3 — 图谱层**：跨页关系 + 架构图 + 依赖分析

层间兼容，渐进增强。后端（Local/Remote）只是实现手段。

**价值：** 编目质量决定查询命中率，决定总 token 成本。一次入库的收益在每次查询时复利。

详见 [docs/directions/d-ai-enhancement.md](docs/directions/d-ai-enhancement.md)

---

### E — 插件系统

自定义 Parser / Generator 插件，第三方扩展能力。

**价值：** 生态基础

---

### F — 导入/导出/集成

Obsidian / Notion / 飞书等知识库互通。

**价值：** 生态扩展（待研究）
