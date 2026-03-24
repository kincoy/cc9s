# cc9s 项目路线图

> 本文件由 project-lead agent 维护，记录项目阶段状态和关键决策。

## 当前状态

- **发布状态**: `v0.1.3` 待发布
- **产品状态**: 当前主干功能已覆盖项目浏览、会话浏览、会话恢复、详情、日志、搜索、命令模式、多选删除、skills / agents 资源浏览和 Phase 8 UI 重构
- **最近更新**: 2026-03-24
- **文档口径**:
  - `README.md` / `README_zh.md` 是对外功能入口
  - 本 roadmap 记录阶段演进和项目里程碑
  - `specs/001-cc9s-cli-tool/` 中的 plan/tasks/research 文档为历史实施记录
- **最近完成**:
  - ✅ Agent Resource Management 完成（agents 一级资源、列表/详情、Claude 对齐 Ready/Invalid 状态）
  - ✅ Skill Resource Management 完成（skills 一级资源、列表/详情、Ready/Invalid 状态）
  - ✅ Session Lifecycle v1 完成（Active / Idle / Completed / Stale 四态、detail 判定依据、header/search 口径统一）
  - ✅ Phase 8.1 完成（样式基础）
  - ✅ Phase 8.2 完成（布局重构）
  - ✅ Phase 8.3 完成（细节优化）
  - ✅ Phase 8.4 完成（文档更新和验证）

## 最近特性补充

### Agent Resource Management（2026-03-24）

**目标**: 把本地 file-backed Claude Code agents 纳入 cc9s 的统一资源模型，使其能够像 projects、sessions 和 skills 一样被浏览、查看和识别状态。

**结果**:
- 新增 `agents` 一级资源，可通过命令模式切换
- 覆盖项目级、用户级和 plugin 级 file-backed agents，built-in agents 保持不纳入 v1
- agents 视图复用现有 context 语义，支持 `all` 与项目上下文切换
- 列表显示作用域、名称、摘要和状态，支持搜索、排序和详情查看
- 详情面板展示路径、来源、作用域、模型、tools 和归一化 availability reasons
- 支持 `e` 直接打开 agent 定义文件
- `Ready / Invalid` 状态与 Claude Code 的 `claude agents` 识别结果对齐
- Claude 识别失败时，agents 资源页显示明确 load error，而不是伪造状态

**验证**:
- ✅ `go test ./...`
- ✅ `go build ./...`
- ✅ `go vet ./...`
- ✅ 已完成一轮真实 TUI 验证，覆盖 `:agents`、project context、invalid 样本可见性和 plugin agent ready 状态

---

### Skill Resource Management（2026-03-24）

**目标**: 把本地 Claude Code skills 纳入 cc9s 的统一资源模型，使其能够像 projects 和 sessions 一样被浏览、查看和识别状态。

**结果**:
- 新增 `skills` 一级资源，可通过命令模式切换
- 统一纳入项目级、用户级和 plugin 级可用资源，覆盖 `skills` 与 `commands`
- 支持目录型 `SKILL.md` skill 与单文件 markdown 资源
- skills 视图复用 context 语义，支持 `all` 与项目上下文切换
- 列表显示作用域、名称、类型、摘要和状态，支持搜索、排序和详情查看
- 支持 `e` 直接打开 skill / command 入口文件进行编辑
- `Ready / Invalid` 两态语义统一复用同一份本地扫描结果

**验证**:
- ✅ `go test ./...`
- ✅ `go build ./...`
- ✅ `go vet ./...`
- ✅ 已完成一轮实际交互修正，覆盖 context、搜索、排序、编辑和全局视图展示

---

### Session Lifecycle v1（2026-03-23）

**目标**: 把 session 状态从粗粒度 active/completed 语义升级为可解释的四态生命周期模型。

**结果**:
- 列表统一显示 `Active` / `Idle` / `Completed` / `Stale`
- `Stale` 明确定义为“session 已不再可靠”，而不是“更久没动”
- 详情面板复用现有入口，新增 lifecycle 状态与 2-4 条判定依据
- 搜索、header 汇总、进入/删除保护统一复用同一份 lifecycle snapshot

**验证**:
- ✅ `go test ./internal/claudefs/...`
- ✅ `go build ./...`
- ✅ `go vet ./...`
- ✅ quickstart/TUI 手工验证已覆盖列表、header、detail、状态搜索和现场 `Stale` 样本
- ⏳ 外部 `claude --resume` 接管后的返回画面链路在当前 TTY 验证环境中仍待补充确认

---

## 阶段定义

### Phase 0: 数据层探索

**目标**: 搞清楚 Claude Code 的真实数据结构，确认可提取的信息字段。

**任务**:
- [x] 扫描 `~/.claude/` 目录结构，绘制目录树
- [x] 解析 session `.jsonl` 文件格式，确认字段
- [x] 确认项目与会话的关联方式（路径映射）
- [x] 确认会话状态判断逻辑（活跃 vs 已完成）
- [x] 输出数据模型草案（Go struct 定义）

**验证标准**:
- ✅ 能用代码扫描出本地所有项目和会话（29 项目 / 530 会话）
- ✅ 数据模型覆盖 spec 中定义的 Key Entities 字段
- ✅ 关键假设（Assumptions 章节）已被验证或修正

**产出**: `docs/phase0-data-exploration.md`

**状态**: ✅ 完成（2026-03-20）

**审查记录**（2026-03-20）:
- ✅ 产物文档完整（docs/phase0-data-exploration.md）
- ✅ 目录结构清晰，包含实际数据规模统计（29 项目 / 530 会话）
- ✅ 数据模型覆盖 spec 中定义的所有 Key Entities 字段
- ✅ 关键假设已验证（5/5 项通过或标记为需处理）
- ✅ 提供了实现建议（懒加载、缓存、并发）
- ✅ 性能优化策略明确（500+ 会话、< 3s 启动）
- ⚠️ 发现 Git 分支字段可能缺失，已记录需处理边界情况
- **结论**: Phase 0 所有验证标准均已达标，可以进入 Phase 1

---

### Phase 1: 最小可运行骨架

**目标**: TUI 框架搭建，三段式布局空壳能正常渲染。

**任务**:
- [x] Go module 初始化，依赖引入
- [x] TUI 框架集成（Bubble Tea v2）
- [x] 四段式布局：Header / Breadcrumb / Body / Footer
- [x] 基础键盘事件：q 退出、? 帮助
- [x] 终端尺寸自适应（含 < 6 行降级处理）

**验证标准**:
- ✅ `cc9s` 启动后能看到四段式布局
- ✅ Header 显示 logo 和占位统计
- ✅ Footer 显示快捷键提示
- ✅ q 键正常退出
- ✅ 终端缩放时布局不崩溃
- ✅ go build / go vet 通过
- ✅ Go 代码审查 APPROVED

**状态**: ✅ 完成（2026-03-20）
**前置**: Phase 0

**踩坑记录**: 详见 `.claude/agent_docs/bubbletea-v2-pitfalls.md`

---

### Phase 2: 项目列表视图

**目标**: 显示真实的项目列表，支持键盘导航。

**任务**:
- [x] 数据扫描集成到 TUI
- [x] 项目列表渲染（表格形式）
- [x] j/k 和方向键导航
- [x] 列排序（按时间、名称、会话数、大小）
- [x] 选中行高亮
- [x] 空状态处理

**验证标准**:
- ✅ 启动后看到真实的项目列表
- ✅ 列包含：项目名、会话数、最近活跃时间、大小
- ✅ j/k 导航流畅
- ⏳ 500+ 会话场景下启动 < 3 秒（待用户验证）

**状态**: ✅ 完成（2026-03-21）
**前置**: Phase 1

**实现细节**:
- 新增 `internal/data/` 包：types.go（数据结构）、format.go（格式化工具）、scanner.go（目录扫描）
- 新增 `internal/ui/table.go`：自定义表格渲染（不依赖 Bubbles table）
- Model 扩展：异步数据加载、cursor 导航、排序支持
- 快捷键：j/k（上下）、g/G（首尾）、s/S（排序切换/反转）
- 所有任务（13 个）已完成，go build/vet 通过

---

### Phase 3: 会话列表 + 层级导航

**目标**: Enter 进入项目的会话列表，Esc 返回，面包屑导航。

**任务**:
- [x] 会话列表视图
- [x] Enter 进入 / Esc 返回
- [x] 面包屑导航（Breadcrumb）
- [x] Header 统计信息更新
- [x] Footer 快捷键动态切换
- [x] 返回时恢复选中位置

**验证标准**:
- ✅ Projects → Sessions 完整导航流程
- ✅ 面包屑正确显示当前路径
- ✅ 返回后光标位置保持

**状态**: ✅ 完成（2026-03-21）
**前置**: Phase 2

**实现细节**:
- 新增 Session 数据结构和会话扫描器（LoadProjectSessions）
- 新增 SessionListModel：独立的会话列表视图
- 新增 session_table.go：会话表格渲染（SESSION ID, STATUS, LAST ACTIVE, EVENTS, SIZE）
- AppModel 路由：监听 EnterProjectMsg 和 BackToProjectsMsg
- 面包屑动态显示："Projects > {项目名}"
- Footer 快捷键根据视图动态切换
- 光标位置保持：lastProjectCursor
- 所有任务（12 个）已完成，go build/vet 通过

**踩坑修复**:
- 活跃会话判断：sessions/*.json 文件名是 PID 而非 sessionID，需解析 JSON
- 列对齐：用 lipgloss.Width() 处理 Unicode 字符（●）和样式码
- 弹性列宽：SESSION ID 列自适应终端宽度

---

### Phase 4: 会话恢复

**目标**: 从 cc9s 直接进入 Claude Code 会话。

**任务**:
- [x] 暂停 TUI，让出终端控制（tea.ExecProcess）
- [x] 调用 `claude --resume <session-id>`
- [x] Claude Code 退出后恢复 TUI
- [x] 活跃会话检测与冲突提示（确认对话框）
- [x] Flash 消息系统（成功/错误反馈）
- [x] 清屏处理（进入/退出）

**验证标准**:
- ✅ 选中会话按 Enter，进入 Claude Code 交互
- ✅ 退出后自动回到 cc9s，会话状态已更新
- ✅ 活跃会话有冲突提示（y/n 确认对话框）

**状态**: ✅ 完成（2026-03-21）
**前置**: Phase 3

**实现细节**:
- 新增 ConfirmDialogModel：独立的确认对话框组件（可复用）
- Bubble Tea ExecProcess API：自动处理终端释放和恢复
- AppModel 集成对话框系统：showingDialog + confirmDialog
- Flash 消息系统：Footer 临时替换，成功 2s / 错误 5s
- 清屏处理：进入和退出时执行 \033[H\033[2J
- 活跃会话检测：Enter 键检查 session.IsActive
- y/n 快捷键确认（对标 k9s）
- 返回后刷新当前项目会话列表
- Ctrl+C 静默处理（不显示错误）
- 所有任务（12 个）已完成，go build/vet 通过

---

### Phase 5: 详情 + 日志

**目标**: d 键查看详情面板，l 键查看事件日志。

**任务**:
- [x] 详情面板（会话元数据 + 事件统计）
- [x] 日志视图（事件流、可滚动）
- [x] Esc 返回列表
- [x] 数据层：ParseSessionStats 和 ParseSessionLog
- [x] UI 层：DetailViewModel 和 LogViewModel
- [x] AppModel 路由集成

**验证标准**:
- ✅ d 键显示完整的会话元数据（悬浮窗，保留背景）
- ✅ l 键显示可读的事件日志（全屏）
- ✅ 日志支持 j/k/g/G 滚动
- ✅ 详情面板自适应宽度（60%-100%）

**状态**: ✅ 完成（2026-03-21）
**前置**: Phase 3

**实现细节**:
- **数据层**（4 任务）：
  - SessionStats 结构（元数据、对话统计、Token 统计、工具使用）
  - LogEntry + ToolCall 结构（按 turn 组织）
  - ParseSessionStats（单 pass 遍历 JSONL）
  - ParseSessionLog（分页加载、过滤噪音事件）
- **UI 层 - DetailViewModel**（5 任务）：
  - 详情面板 Model（异步加载统计）
  - View 方法：自适应宽度 `clamp(width*0.6, 60, 100)`
  - ViewBox 方法：叠加到背景（悬浮窗效果）
  - 5 个区域：标题、元数据、对话统计、工具使用（Top 5）、Token 统计
  - 消息：ShowDetailMsg / CloseDetailMsg / statsLoadedMsg
- **UI 层 - LogViewModel**（5 任务）：
  - 日志视图 Model（异步加载前 100 turn）
  - j/k 滚动，g/G 跳转
  - 按 turn 组织显示（Turn header + User + Assistant + Tools）
  - 全屏显示（覆盖背景）
  - 消息：ShowLogMsg / CloseLogMsg / logLoadedMsg
- **AppModel 路由集成**（4 任务）：
  - 扩展字段：showingDetail / detailView / showingLog / logView
  - 处理 ShowDetailMsg / CloseDetailMsg / ShowLogMsg / CloseLogMsg
  - renderBody 优先级：dialog > detail > log > help > session/project
  - 详情视图用 overlayDialog 叠加，日志视图全屏
- **SessionListModel 快捷键**（2 任务）：
  - d 键：发送 ShowDetailMsg
  - l 键：发送 ShowLogMsg
- **样式和 Footer**（3 任务）：
  - 新增样式：DetailTitleStyle, DetailSectionStyle, DetailLabelStyle, DetailValueStyle
  - 新增样式：LogTitleStyle, LogTurnHeaderStyle, LogUserStyle, LogAssistantStyle, LogToolStyle
  - Footer 新增：`<d> Detail  <l> Logs` 提示
- **辅助功能**：
  - 新增 FormatTime、FormatDuration、FormatNumber 格式化函数
  - 修复 ParseSessionStats 从前几行提取元数据（version 和 gitBranch 在第 2+ 行）

**总任务数**: 30 个（全部完成）
**编译状态**: ✅ go build / go vet 通过
**功能验证**: ✅ 用户确认通过

---

### Phase 6: 搜索 + 命令模式 + 全局会话 + 多选删除

**目标**: 补齐剩余核心交互功能。

**任务**:
- [x] / 搜索过滤（项目/会话/全局会话，实时过滤，Esc 退出恢复）
- [x] : 命令模式（:sessions 切全局会话，:projects 切回项目列表）
- [x] 全局会话视图（跨项目会话列表，6 列表格含 PROJECT）
- [x] Space 多选（绿色背景高亮，Footer 显示选中数量）
- [x] Ctrl+D 删除（单个 + 批量，活跃会话保护 alert）
- [x] 删除后同步刷新 project 列表
- [x] 确认对话框（confirm + alert 两种模式）
- [x] 全局会话视图支持 Enter/d/l（与普通会话视图功能对齐）

**验证标准**:
- ✅ / 输入关键词后列表实时过滤
- ✅ :sessions 能切换到全局会话视图
- ✅ 删除有确认流程，活跃会话不可删除
- ✅ 删除后返回 projects，SessionCount 已更新

**状态**: ✅ 完成（2026-03-22）
**前置**: Phase 4, Phase 5

**延后功能**:
- Context 架构（namespace 过滤）：类型和消息已移除，待后续独立实现

**实现细节**:
- 新增 GlobalSessionListModel：全局会话列表（搜索 + 多选 + 删除）
- 新增 global_session_table.go：6 列表格（PROJECT | SESSION ID | STATUS | LAST ACTIVE | EVENTS | SIZE）
- 新增 delete.go：DeleteSession / DeleteSessions / DeleteTarget
- ConfirmDialogModel 扩展 alert 模式（任意键关闭，红色边框）
- 新增 MultiSelectedStyle（绿色背景高亮选中行）
- 去掉 ✓ marker 列，改用纯颜色高亮
- Completed 状态补充 ○ 符号
- Bubble Tea v2 Space 键修复（"space" 而非 " "）
- 表格对齐修复（header 前导空格导致偏移）
- 删除后 projectList 同步刷新（scanProjectsCmd）

---

### Phase 7: 统一 Session 视图 + 体验优化

**目标**: 统一 Session 和 GlobalSession 为单一视图，通过 Context 控制显示范围；提升命令输入效率和上下文感知。

**任务**:
- [x] Context 类型定义（ContextAll / ContextProject）
- [x] 统一 SessionListModel（使用 GlobalSession + Context 过滤）
- [x] 统一表格渲染（showProjectColumn 控制 5/6 列）
- [x] 移除 ScreenGlobalSessions，简化路由（只有 Projects 和 Sessions）
- [x] Enter 进入项目时自动设置 ContextProject
- [x] :context <name|all> 命令 + 0 快捷键
- [x] 删除 global_session_list.go 和 global_session_table.go
- [x] Footer 增加排序和 context 提示
- [x] 命令模式 Tab 补全（命令名 + :context 项目名）
- [x] inline suggestion（输入 `pro` 显示灰色 `jects`，zsh-autosuggestions 风格）
- [x] Header Context 标识（All Projects / 项目名 替代数字计数）
- [x] SUMMARY 列（提取第一条用户消息，弹性宽度，空值显示 `-`）
- [x] Detail 面板新增 Session ID 字段和 Summary 区域（1000 字符限制）

**验证标准**:
- ✅ go build / go vet 通过
- ✅ 项目视图 → Enter → 5 列会话 → Esc 返回 → :sessions → 6 列 → :context cc9s 过滤
- ✅ `:pro` Tab → `projects `，连续 Tab 循环
- ✅ `:context cc` Tab → 项目名循环
- ✅ Header 显示 `All Projects` / 项目名
- ✅ SUMMARY 列单行显示，Detail 面板完整展示

**状态**: ✅ 完成（2026-03-22）
**前置**: Phase 6

**实现细节**:
- Context 过滤链：allSessions → applyContext() → contextSessions → applySearchFilter() → sessions
- Screen 枚举简化为 Projects 和 Sessions 两种
- SwitchContextMsg 消息驱动 context 切换
- 统一 loadAllSessionsCmd（始终加载全部，内存过滤）
- 代码复用度大幅提升，删除 ~280 行重复代码
- Tab 补全在 `handleCommandInput` 中拦截 `tab` 按键，匹配候选列表循环切换
- 自定义 `renderCommandLine()` 替代 textinput.View()，prompt + 用户输入(白色) + suggestion(灰色)
- 数据层 ExtractSessionSummary 提取第一条用户消息，清洗不截断，UI 层各自截断
- 列顺序：SESSION ID(固定12) | SUMMARY(弹性) | STATUS | LAST ACTIVE | EVENTS | SIZE

---

### Phase 8: UI 重构（对标 k9s）

**目标**: 纯 UI 重构，提升视觉层次、布局骨架和交互表达，对标 k9s 的专业终端体验。

**任务**:
- [x] Phase 8.1 - 样式基础（色彩系统模块化）
  - [x] 创建 colors.go, border.go, text.go, status.go
  - [x] 重构 styles.go
  - [x] ColorNormal 从 #A0A0A0 提升到 #BBBBBB
  - [x] 选中行背景改为 aqua (#00D7FF)
  - [x] 更新所有文件的样式引用
- [x] Phase 8.2 - 布局重构（Header/表格/Footer）
  - [x] Header 加边框（ThickBorder）
  - [x] Breadcrumb 当前路径 aqua 高亮（后删除）
  - [x] 删除 Breadcrumb 层（避免信息重复）
  - [x] Panel title 嵌入顶部边框（方案 A：手动绘制）
  - [x] 表头改成连续背景 bar
  - [x] Footer 状态机（FooterContext）
  - [x] 修复 Bug 1: Panel header 不连续且未对齐（外层 Style 缺背景色）
  - [x] 修复 Bug 2: Sessions (context=all, 7列) 没有边框
  - [x] LAST ACTIVE 列表头和数据行统一右对齐
- [x] Phase 8.3 - 细节优化（CommandBar 边框、Flash 背景色、80×24 自适应）
  - [x] CommandBar 边框颜色区分（: aqua, / seagreen）
  - [x] Flash 消息背景色（成功绿底白字，错误红底白字）
  - [x] Projects 表格 width<100 隐藏 SIZE 列
  - [x] Sessions 表格 width<100 隐藏 EVENTS 和 SIZE 列
  - [x] 修复 session_table sepCount 计算错误
- [x] Phase 8.4 - 文档更新和验证
  - [x] 更新 roadmap 和 MEMORY.md
  - [x] 更新 bubbletea-v2-pitfalls.md（新增第 12、13 条）

**验证标准**:
- ✅ go build / go vet 通过
- ✅ Panel title 嵌入顶部边框（`┏━ Projects(29) ━━━┓`）
- ✅ 表头连续（一整条深色 bar）
- ✅ 表格铺满 panel（无右侧空白）
- ✅ LAST ACTIVE 列右对齐
- ✅ Footer 视图驱动（根据状态动态切换）
- ✅ CommandBar 边框颜色区分
- ✅ Flash 消息背景色
- ✅ 80×24 表格列自适应

**状态**: ✅ 已完成（2026-03-22）
**前置**: Phase 7

**实现细节**:
- 方案 A：手动绘制 ThickBorder 边框（`┏ ━ ┓ ┃ ┗ ┛`）
- Panel title 格式：`Projects(N)` / `Sessions(context)(N)`
- 删除 Breadcrumb 层（Panel title 已提供 context 信息）
- Footer 状态机：10 种状态的快捷键映射

**收尾说明**:
- Phase 8 中记录的渲染问题已在实现过程中修复
- 如后续出现新的 UI 问题，请创建新的 bug 文档，不要复用已完成阶段的临时问题列表

---

## 关键决策记录

| 日期 | 决策 | 原因 |
|------|------|------|
| 2026-03-20 | 采用 6 阶段迭代开发 | 避免瀑布式一次性 plan，每阶段可独立验证 |
| 2026-03-20 | Phase 0 优先做数据探索 | 数据结构是所有后续工作的基础，且最可能有意外 |
| 2026-03-22 | Phase 7 统一 Session 视图 + Context | 对标 k9s namespace 模式，减少重复代码，新资源类型（skills）可复用 |
| 2026-03-22 | Phase 8 UI 重构（对标 k9s） | 提升专业度、面板感、信息密度，删除 Breadcrumb 避免重复 |
| 2026-03-22 | Panel title 嵌入边框（方案 A） | 对标 k9s 风格，手动绘制边框以实现 title 嵌入效果 |

## 风险追踪

| 风险 | 影响 | 状态 |
|------|------|------|
| Claude Code 数据格式可能跨版本不一致 | 数据解析需要兼容处理 | 待 Phase 0 验证 |
| 会话恢复可能需要特殊的终端处理 | Phase 4 可能比预期复杂 | 待 Phase 4 验证 |
| 500+ 会话的性能表现 | 可能需要提前考虑懒加载 | 待 Phase 2 验证 |
