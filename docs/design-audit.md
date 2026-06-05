# Baize Wiki — 设计阶段审计报告（第二版）

> 审计者: CodeBuddy Code
> 审计日期: 2026-06-05（第二版）
> 审计范围: `docs/` 下全部 6 份设计文档 + `README.md`
> 项目阶段: 设计迭代后，零代码实现

---

## 0. 执行摘要

经过第一轮审计后，`architecture.md`、`data-model.md`、`phase-1-plan.md` 三份文档有实质性更新，体现了对审计意见的响应。但本轮审计发现了 **3 类不可接受的问题**：

1. **第一轮标记为"阻塞"的问题，0 个被修复** — 包括 README 不同步、增量构建策略、Linker 问题
2. **更新本身引入了新的一致性矛盾** — `cli-spec.md` 的构建流程与新的架构陈述直接冲突
3. **新设计的 Level 系统存在关键逻辑漏洞** — Level 1 的内容合并策略未定义

**综合结论：设计准备度从 70% 退步到 65%**。不是因为文档变差了，而是因为暴露出的问题比第一轮审计时更多了。在修复所有"阻塞"和"严重"级问题之前，不建议开始编码。

---

## 1. 变更追踪：什么改了，什么没改

| 文档 | 变更内容 | 对应第一轮审计意见 |
|------|---------|-----------------|
| `architecture.md` | +Level 输出系统（第 4 节）、+扫描策略及二进制探测（第 5 节）、数据流移除 Linker、路线图调整 | 部分响应了"扫描策略不完整" |
| `data-model.md` | 精简 Config（移除 include/pretty_urls/index/backlinks）、移除 Linker/Link 类型 | 响应了"模型含有 Phase 1 用不到的字段" |
| `phase-1-plan.md` | 升级到 v2：Scope 重构（移除 Linker/`[[wiki-link]]`）、M3/M4/M5 重写、依赖精简 | 响应了"Phase 1 范围过大" |
| `cli-spec.md` | +`--level` flag、`init` 模板更新 | 局部响应 |
| `design-overview.md` | **未变更** | — |
| `mcp-spec.md` | **未变更** | — |
| `README.md` | **未变更** | ❌ 第一轮标记为"必须修复" |

**第一轮 10 项修复建议的执行率：0/3 阻塞 + 1/4 高优 + 0/3 中优 = 10%。** 这个执行率不可接受。

---

## 2. 第一轮阻塞问题（均未修复）

### 2.1 [阻塞] README 与设计文档矛盾 — 完全未处理

第一轮审计第 3.1 节的 3 项矛盾：

| 矛盾 | 状态 | 影响 |
|------|------|------|
| "架构形态尚未确定" vs 已确定 Go 技术栈 | **未修复** | 任何新加入者读到 README 都会被误导 |
| `src/` 目录 vs `cmd/`+`internal/` 布局 | **未修复** | 仓库里 `src/` 是个空目录，令人困惑 |
| "可能的形态"列了 5 种 vs 已确定 CLI+MCP 双模式 | **未修复** | README 和设计文档在讲两个故事 |

**这 3 项中每一行的修复工作量都不超过 15 分钟。没有被修复只有一种解释：要么没读审计报告，要么不认同审计结论。如果是后者，请在 README 中明确说明"我们决定保留 README 的开放性表述"，而不是留下一份矛盾的文档。**

### 2.2 [阻塞] `[[wiki-link]]` 解析方案 — 被绕过而非解决

Phase 1 计划将 `[[wiki-link]]` 解析推迟到 Phase 5。这个决策本身是合理的——范围控制。但有以下问题：

- **`architecture.md` 第 5 节路线图将 Phase 5 标注为"交叉链接 + 生态集成"**，但 `[[wiki-link]]` 不仅仅是一个"生态"功能——它是 Wiki 的核心价值主张。将链接功能推到最后一个 Phase，意味着从 Phase 1 到 Phase 4 生成的 Wiki 页面之间没有任何交叉引用。**这本质上是在生成一个"文件夹里的 Markdown 文件集合"，而不是一个 Wiki。**
- **风险并没消失，只是被推迟了**：到 Phase 5 时，Parser、Generator、Storage 可能已经做了不利于链接引入的假设，届时改造代价更大。

**建议：** 至少定义 `[[wiki-link]]` 在 Markdown 源文件中的存在形式（语法 + AST 节点），即使 Phase 1 不做解析。或者接受"Phase 1 输出的是文档目录而非 Wiki"并体现在产品定位中。

### 2.3 [阻塞] 增量构建策略 — 完全未处理

`--no-cache` 标志依然存在于 `cli-spec.md` 中，但没有任何文档解释缓存如何工作。

关键问题一个都没回答：
- 再次构建时是全量重建还是增量？`meta.json` 中的 `version` 递增策略是什么？
- 如果只改了 1 个文件，Generator 是否需要为全部 1000 个页面重新生成 `_index.md`？
- 如果我删除了源目录中的一个文件，第二次 build 会从 Wiki 输出中删除对应页面吗？

**要求：** 在 `phase-1-plan.md` 中明确定义 Phase 1 是全量构建还是增量构建。如果是全量构建，删除 `--no-cache` 标志（它不存在时默认行为就是"不缓存"）。如果是增量构建，定义文件变更检测策略。

---

## 3. 更新引入的新问题

### 3.1 [严重] `cli-spec.md` 构建流程与 `architecture.md` 矛盾

`architecture.md` 第 6.1 节明确声明：

> Phase 1 不含 Linker 和 Indexer，后续 Phase 逐步加入。

但 `cli-spec.md` 第 2.2 节的构建流程（第 115-125 行）仍然包含：

```
5. Linker 计算交叉链接
6. Generator 生成 Wiki 目录树 + _index.md
7. Indexer 构建全文索引 → .baize/
```

**这是两份设计文档之间的直接矛盾**。根据 architecture.md 的路线图，Linker 在 Phase 5，Indexer 在 Phase 3。`cli-spec.md` 的构建流程必须同步更新。

同样，`cli-spec.md` 的 JSON 输出示例（第 281 行）包含 `"links": 47, "dangling_links": 3`，如果 Linker 不在 Phase 1 中，这些字段不应该出现在输出中。

### 3.2 [严重] Section 模型不一致

`data-model.md` 第 2.4 节的 Section 结构体定义包含 `Children []Section`：

```go
type Section struct {
    Children  []Section  `json:"children"`   // 子段落
}
```

但 `phase-1-plan.md` M2（第 143 行）的 Section 定义是：

```
Section: ID, Level, Title, Content
```

**没有 `Children` 字段。** 如果 Section 没有子段落，标题树就只是一层扁平列表，失去了树形结构的意义。哪个定义是正确的？

### 3.3 [严重] Level 1 内容合并策略未定义

`architecture.md` 第 4 节定义 Level 1（Flat）时写道：

> 同一分类下的多文件内容合并到一个页面

但"合并"的算法完全没有明确定义：
- **合并边界**：什么算"同一分类"？按 frontmatter category？按顶层目录？如果文件既没 category 又不在同一目录下，合并逻辑是什么？
- **合并顺序**：多个文件合并到页面时，按什么顺序排列？按文件名排序？按 weight？如果用户调整了文件顺序，每次构建的输出会变化吗？
- **冲突处理**：如果两个文件有各自的 frontmatter（不同 tags、不同 description），合并后用什么？第一个文件的？合并后的 union？
- **边界情况**：一个 100KB 的文件和一个同样 100KB 的文件合并，结果页面可能达到 200KB，AI Agent 读取时内容截断了怎么办？

**要求：** 在写入一行代码之前，必须用伪代码或自然语言明确定义合并算法。一个无法被第三方独立实现的规范不是规范，只是想法。

### 3.4 [严重] 缺失反模式的显式否定：`organize.by: flat` vs `--level 1`

`data-model.md` 的 `organize.by` 枚举值包括 `flat`，而 `architecture.md` 的 Level 系统定义了 Level 1 为 "Flat" 模式。

这两个概念关系如何？
- `--level 1` 是 `organize.by: flat` 的同义词？
- 如果我传了 `--level 3 --organize flat`，哪个优先？
- 如果都不优先，这是否是一个合法的配置？如果是，它应该报错。

**要求：** 明确 `--level` 与 `organize.by` 的关系。推荐将 `organize.by` 在未来 Phase 的 tags 模式时才使用，Phase 1 中 `--level` 即为输出结构的唯一控制参数。

---

## 4. 设计逻辑缺陷

### 4.1 "能扫尽扫"策略的实际产出质量风险

`architecture.md` 第 5 节的扫描策略声称"能扫尽扫"——只过滤二进制文件，其他全部纳入。设想以下真实场景：

```bash
# 在一个典型的 Go 项目上运行
baize-wiki build . --level 2
```

扫描结果将包括：
- `go.mod`、`go.sum` → 显示为"go.mod"页面，内容是一堆版本号
- `vendor/github.com/.../*.go` → 生成几百个无意义的"源码页面"
- `node_modules/` → 如果不小心没配 `.baizeignore`，会生成数万个页面
- `.git/config`、`.git/HEAD` → Git 内部文件暴露为 Wiki 内容

**这是"能扫尽扫"策略在真实项目中的必然结果。** `phase-1-plan.md` 的风险表中有一条"二进制探测误判"，但真正的风险不是误判——是**非文档文本文件被误当作文档处理**。

**建议：**
- 默认只扫描 `**/*.md` 和 `**/*.mdx`（白名单模式）
- `--scan-all` 标志启用"能扫尽扫"模式
- 或者在输出中为不同文件类型打上来源标记，让 Agent 可以区分"这是文档"和"这是源代码"

### 4.2 空输出的定义缺失

以下场景的输出应该是什么：
- 空目录 → Scanner 返回空列表 → Generator 输入为空
- 目录下只有二进制文件 → 同上
- 目录下只有 `.baizeignore` 排除了的文件 → 同上

`generator_test.go` 的测试边界写的是"空 pages 列表 → 仅根 _index.md"，但生成的根 `_index.md` 内容是什么？"该 Wiki 还没有页面"？这是一个产品设计问题，但没有任何文档规定。

**要求：** 增加"空 Wiki"的产品行为规范，至少定义 CLI 输出的消息和退出码。

### 4.3 Frontmatter `yaml:",inline"` 的隐式缺陷

`data-model.md` 的 Frontmatter 结构体使用 `yaml:",inline"` 捕获自定义字段。这个设计有已知问题：

```go
type Frontmatter struct {
    Title   string            `yaml:"title"`
    Custom  map[string]any    `yaml:",inline"`
}
```

```yaml
---
title: 数据模型
title: 重复的值
---
```

`yaml:",inline"` 在 `gopkg.in/yaml.v3` 中的行为：如果 `Custom` 中出现了与具名字段（如 `title`）同名的 key，**具名字段会覆盖 map 中的同名 key**。但如果 YAML 中有重复 key，解析结果取决于实现。

更重要的是：如果用户在 frontmatter 中写 `type: blog`（一个没有被 Frontmatter 结构体定义的字段），会被 `Custom` 捕获。但如果未来某个版本将 `Type` 加入 Frontmatter 结构体，旧版本写入的 `type` 就会和新版本的具名字段发生覆盖冲突——**这是向后兼容性地雷。**

**建议：** 将 `Custom` 的 key 加上命名空间前缀（如 `x-`），或者在文档中明确声明"自定义字段可能与未来版本冲突"。

---

## 5. 一致性复查——仍然存在的问题

### 5.1 版本号：还是不一致

| 文档 | 表述 | 问题 |
|------|------|------|
| `cli-spec.md` | Phase 1 → v0.1.x | 暗示 v0.1.0 是一个正式版本号 |
| `phase-1-plan.md` | 打 tag `v0.1.0-alpha` | 明确标记为 alpha |

`v0.1.0` 和 `v0.1.0-alpha` 在 SemVer 中是不同的版本标识。如果使用 Go module，`v0.1.0-alpha` 作为伪版本在 `go get` 时行为与 `v0.1.0` 不同。**统一用一个。**

### 5.2 viper 依赖仍未在架构层面记录

`phase-1-plan.md` 使用 `github.com/spf13/viper`，`architecture.md` 的技术决策表中没有它。要么加进去并说明理由，要么去掉 viper 用 yaml.v3 直接反序列化。

### 5.3 MCP spec 中 `inputSchema` 缺少 `required` 字段

`mcp-spec.md` 中 6 个 Tools 的 `inputSchema`，只有 `wiki_search` 和 `wiki_add` 声明了 `required`。其余工具（`wiki_build`、`wiki_read`、`wiki_list`、`wiki_stats`）缺少 `required` 声明或将其明确设为空数组。

MCP 客户端在渲染 Tool Schema 时，`required` 字段决定了客户端是否会强制用户填写必填参数。缺失 `required` 可能导致工具调用时缺少关键参数。

---

## 6. Phase 1 计划深度审计

### 6.1 M1 依赖引入顺序风险

`phase-1-plan.md` M1 要求在 Day 1 就引入 `cobra` + `viper` 两个依赖。但在 M2-M5 的开发过程中，viper 实际上并不需要（配置可以 hardcode）。提前引入 viper 的风险：

- Day 1 的 `go get` 可能引入不需要的传递依赖
- viper 的配置合并逻辑可能比预期复杂，导致 M6 的 config 实现返工
- 如果后续决定替换 viper，Day 1 引入的依赖包袱已经存在

**建议：** M1 只引入 cobra，viper 在 M6 引入。或者在 M1 就做一个最小化 viper 集成的 Spike。

### 6.2 M5 `DirNode` 缺少序列化标记

`phase-1-plan.md` M5 的 DirNode 定义（第 307-313 行）：

```go
type DirNode struct {
    Name     string
    Path     string
    Pages    []*Page
    Children []*DirNode
    Depth    int
}
```

`Pages` 字段直接引用了 `*Page`（完整的 Page 对象而非路径字符串）。如果 DirNode 用于 JSON 序列化（例如 `info --tree --json`），Page 对象可能携带大量不需要的数据（Content 字段可能数 KB）。**当前定义会导致 `info --tree --json` 输出膨胀。**

**建议：** 为 DirNode 定义一个轻量的 PageRef（仅有 Title 和 Path）代替 `*Page`，或者在 Tree 构建时只保留必要字段。

### 6.3 M5 原子写入存在边缘情况

"先写 `.tmp` 文件再 `os.Rename`"的原子写入策略在多平台上有不同语义：
- Linux: `os.Rename` 在同一文件系统内是原子的 ✓
- macOS: 同上 ✓
- Windows: `os.Rename`（实际调用 `MoveFileEx`）如果目标文件已存在，**可能失败**（取决于文件是否被打开）

**建议：** 在文档中注明此限制，或在 Windows 上使用 `MoveFileEx` 的 `MOVEFILE_REPLACE_EXISTING` 标志（通过 `syscall`）。

### 6.4 测试边界缺少性能基线和回归防护

`phase-1-plan.md` 各里程碑的测试边界只覆盖了功能正确性，没有覆盖：
- Scanner 在 1 万 / 10 万文件下的性能基线
- Parser 在 1MB Markdown 文件下的内存占用
- Generator 在 1000 页面下的输出速度

**建议：** 在 M7 中加入 bench test，记录关键路径的性能基线到 CI。

---

## 7. 产品定位的根本矛盾

### 7.1 "AI Agent 的 Wiki 工具"但输出格式是人类友好的

所有设计文档都强调目标用户是 AI Agent，但：
- 生成的 `_index.md` 模板是面向人类阅读的（标题 + 列表 + 描述）
- CLI 的输出是面向人类阅读的（`✓ 扫描完成: 找到 24 个文件`）
- JSON 输出是事后加入的（所有命令的 `--json` 标志）

如果 AI Agent 是真用户，那么**默认输出应该是机器可读的结构化格式（如 JSON），人类友好的人类输出才是可选模式。** 当前设计是反过来的。

### 7.2 Wiki 的核心价值在 Phase 1 中被彻底剥离

由于 Linker 推迟到 Phase 5、全文搜索推迟到 Phase 3、Level 合并策略未定义，Phase 1 的实际产出是：

> 一个将 Markdown 文件扫描后复制到另一个目录并生成索引页的工具。

这和一个 `find . -name '*.md' -exec cp {} /output \;` 配合 `tree` 命令的区别在哪？Phase 1 的核心价值增量是什么？

**这个问题不需要在代码中回答，但必须在产品文档中回答**——它决定了 Phase 1 是否值得发布，以及用户（即使是 AI Agent）为什么应该使用 Baize Wiki 而不是一个简单的文件复制脚本。

---

## 8. 评分修正

| 维度 | 第一轮评分 | 本轮评分 | 变化原因 |
|------|----------|---------|---------|
| 架构合理性 | ★★★★☆ | ★★★☆☆ | Level 合并算法未定义、扫描策略实际产出质量有严重疑问 |
| 文档完整性 | ★★★★☆ | ★★★☆☆ | 新内容引入了不一致（cli-spec 构建流程 vs architecture） |
| 一致性 | ★★★☆☆ | ★★☆☆☆ | 旧矛盾未修复 + 新矛盾已产生，净恶化 |
| 可落地性 | ★★★★☆ | ★★★☆☆ | Phase 1 范围精简了，但 Level 1 合并策略模糊不可实现 |
| 风险可见性 | ★★☆☆☆ | ★★☆☆☆ | 新增了"全文本扫描"的风险但未设缓解措施 |

**综合：设计准备度从 70% → 65%。不是因为文档少了，而是审计要求的标准提高了。**

---

## 9. 强制执行清单

以下问题按严重程度分级。**"阻塞"级问题必须在进入编码阶段前修复。** 完成后请在 `docs/design-audit.md` 的下一版本中标注状态。

### 阻塞（Blocking）

| # | 问题 | 涉及文档 | 修复动作 |
|---|------|---------|---------|
| B1 | README 与设计文档矛盾 | `README.md` | 重写 README：删除"架构形态尚未确定"和"可能的形态"，替换为已确定的 Go 技术栈、CLI+MCP 演进路线 |
| B2 | `cli-spec.md` 构建流程包含 Linker/Indexer | `cli-spec.md` | 删除构建流程第 5 步(Linker)和第 7 步(Indexer)，删除 JSON 输出中的 links/dangling_links |
| B3 | Level 1 合并算法未定义 | `architecture.md` | 用伪代码定义：合并边界判定、合并顺序、冲突处理、最大页面大小 |
| B4 | Section 模型不一致 | `data-model.md` vs `phase-1-plan.md` | 统一 Section 定义，确认 Children 字段是否存在 |
| B5 | 增量构建策略未定义 | `phase-1-plan.md` + `cli-spec.md` | 明确 Phase 1 是全量构建，删除 `--no-cache` 标志 |

### 严重（Critical）

| # | 问题 | 涉及文档 | 修复动作 |
|---|------|---------|---------|
| C1 | 全文本扫描的产出质量风险 | `architecture.md` | 默认添加 `**/*.md` 白名单，`--scan-all` 启用全扫描 |
| C2 | `organize.by` 与 `--level` 关系未定义 | `data-model.md` | 明确 `--level` 是 Phase 1 唯一输出结构控制参数 |
| C3 | viper 依赖在架构决策表中缺失 | `architecture.md` | 加入 viper 的决策记录，或移除 viper 依赖 |
| C4 | Frontmatter `yaml:",inline"` 向后兼容风险 | `data-model.md` | 在文档中标注自定义字段的兼容性限制 |
| C5 | CLI 默认输出方向错误 | `cli-spec.md` | 重新评估默认输出格式（JSON vs 人类可读），至少补充设计说明 |

### 建议（Recommended）

| # | 问题 | 涉及文档 | 修复动作 |
|---|------|---------|---------|
| R1 | 版本号不统一 | `cli-spec.md` + `phase-1-plan.md` | 统一为 `v0.1.0-alpha` 或 `v0.1.0` |
| R2 | MCP `inputSchema` 缺少 `required` | `mcp-spec.md` | 为所有 Tool Schema 补充 `required` 字段 |
| R3 | M5 DirNode 引用完整 Page 导致序列化膨胀 | `phase-1-plan.md` | 定义轻量 PageRef 替代 |
| R4 | 测试缺少性能基线 | `phase-1-plan.md` | 在 M7 加入 bench test 计划 |
| R5 | Section.Children 是否存在 | `data-model.md` | 统一 Section 定义 |

---

> **最终总结：** 第一轮审计提出了 10 项改进建议，本轮只看到 3 项文档获得了实质性更新，7 项悬而未决，且更新本身引入了 2 项新的一致性问题。同时，新设计的 Level 系统在合并算法这一核心逻辑上留下了缺口——而这是 Phase 1 区别于"文件复制脚本"的唯一价值所在。建议优先级：**B1-B5 修复 → C1-C5 修复 → 重新审计 → 开始编码**。在执行率为 10% 的情况下跳过审计直接编码，不是勇气，是鲁莽。

---

## 附录：第三轮修复状态

> 修复者: Claude Code
> 日期: 2026-06-05

| 编号 | 问题 | 状态 | 变更 |
|------|------|------|------|
| B1 | README 与设计文档矛盾 | ✅ 已修复 | 重写 README: 明确 Go 技术栈、CLI+MCP 路线、Level 系统 |
| B2 | cli-spec 构建流程含 Linker/Indexer | ✅ 已修复 | 删除 Linker/Indexer 步骤，同步输出示例 |
| B3 | Level 1 合并算法未定义 | ✅ 已修复 | 用伪代码定义：Grouping → Ordering → Merging → Splitting (50KB) |
| B4 | Section 模型不一致 | ✅ 已修复 | phase-1-plan.md 补充 Children 字段，与 data-model.md 一致 |
| B5 | 增量构建策略未定义 | ✅ 已修复 | 明确 Phase 1 为全量构建，移除 --no-cache 标志 |
| C1 | 全文本扫描产出质量风险 | ✅ 已修复 | 默认只扫 .md/.mdx，--scan-all 启用全扫描（Phase 2+） |
| C2 | organize.by vs --level 未定义 | ✅ 已修复 | Phase 1 --level 唯一控制，organize.by 标注为 Phase 2+ |
| C3 | viper 缺失于架构决策表 | ✅ 已修复 | 加入架构决策记录 |
| C4 | Frontmatter yaml:",inline" 兼容性 | ✅ 已修复 | 在 data-model.md 添加兼容性警告 |
| C5 | CLI 默认输出方向 | ✅ 已补救 | 增加输出格式设计说明（Human 默认 + --json + MCP 为主） |
| R1 | 版本号不统一 | ✅ 已修复 | 统一为 v0.1.0-alpha |
| R2 | MCP inputSchema 缺 required | ✅ 已修复 | 全部 6 个 Tool Schema 补全 required |
| R3 | DirNode 引用完整 Page | ✅ 已修复 | 新增 PageRef 轻量结构体替代 |
| R4 | 测试缺少性能基线 | ✅ 已修复 | M7 增加 Scanner/Parser/Generator bench test 计划 |
| R5 | Section.Children | ✅ 已在 B4 修复 | — |
