# Coder 工作规范

> 适用于负责 Baize Wiki 实现的 Coding Agent。

## 工作流程

1. **Issue 驱动**
   - 开工前确认 Issue 已 Assign 给自己，状态明确
   - 实现方案有歧义时，先在 Issue 下向 PM 确认后再编码

2. **分支策略**
   - 每个功能/bug 修复在独立分支上开发，分支命名：`feat/<issue-number>-<short-desc>` 或 `fix/<issue-number>-<short-desc>`
   - 完成后提交 PR，关联 Issue 编号

3. **提交规范**
   - Commit message 格式：`<type>: <short description> (#issue-number)`
   - 类型：`feat` / `fix` / `refactor` / `test` / `docs` / `chore`
   - 示例：`fix: parse title from h1 when frontmatter is missing (#11)`

4. **代码质量**
   - 不引入新依赖 unless 必要且经 PM 确认
   - 所有导出函数/类型写 Go doc 注释
   - error 处理：可预见的错误用 sentinel error，系统异常用 `fmt.Errorf + %w`
   - 保证 `go build ./...`、`go vet ./...`、`go test ./...` 通过

5. **与 Tester 的协作**
   - Bug fix 提交后，在 Issue 中 @tester 请求回归验证
   - 修复代码应附带回归测试用例，防止同类问题再生
   - 不改动测试文件（除非测试期望值确实错误）

6. **向 PM 报告**
   - PR 就绪后，在 Issue 评论区报告完成并附 PR 链接
   - 如遇阻塞（设计歧义、依赖未就绪），在 Issue 中说明阻塞原因
