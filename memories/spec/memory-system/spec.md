---
brief: 记忆文档的读写、检索与持久化规范。
capability: memory-system
---

# 记忆系统

## Purpose

把项目关键决策、约束与经验沉淀为可检索的记忆文档，
让 Agent 跨会话保持上下文一致性，避免重复踩坑。

## Requirements

### Requirement: 记忆路径约束

系统 SHALL 只允许读写 `memories/` 目录及其子路径，越界路径必须被拒绝。

**Priority**: P0
**Rationale**: 防止误写项目其他文件。

#### Scenario: 拒绝越界路径

- GIVEN 一个指向 `memories/` 之外的路径
- WHEN 调用记忆写入
- THEN 返回错误，文件不被创建或修改

#### Scenario: 接受子目录路径

- GIVEN 路径 `memories/repo/workflow.md`
- WHEN 调用记忆写入
- THEN 正常写入

### Requirement: 记忆文档头部摘要

系统 SHALL 要求每篇记忆文档以 YAML front matter 的 `brief` 字段开头，
用于在 prompt 中快速识别内容，避免注入正文。

**Priority**: P0
**Rationale**: prompt 只注入清单，无 brief 的文件无法被识别。

#### Scenario: 无 brief 的文件被跳过

- GIVEN `memories/` 下存在一个无 front matter 的 .md 文件
- WHEN 扫描记忆文件清单
- THEN 该文件不出现在清单中

### Requirement: 记忆检索

系统 SHALL 提供 `memory_search` 指令，基于关键词检索记忆文档，
支持跨工作区（当前工作区优先，其他工作区按阈值过滤后追加）。

**Priority**: P1
**Rationale**: 记忆数量增长后，人工翻阅不可行。

#### Scenario: 当前工作区优先

- GIVEN 当前工作区与其他工作区均有相关记忆
- WHEN 执行检索
- THEN 当前工作区的命中排在前面

#### Scenario: 写入后立即可检索

- GIVEN 刚写入一篇新记忆
- WHEN 立即执行检索
- THEN 新记忆可被命中（索引按需重建）

### Requirement: 工作区物理落盘

系统 SHALL 把记忆写入工作区 `memories/` 目录（唯一事实来源），并保持检索索引同步。

**Priority**: P0
**Rationale**: 工作区目录是版本控制跟踪的事实来源，需与索引一致。

#### Scenario: 写入后物理文件存在

- GIVEN 一次成功的记忆写入
- WHEN 检查工作区 `memories/` 目录
- THEN 对应文件存在且内容一致

## 约束

- 记忆文档单条不超过 5 行，优先使用列表与陈述句
- 同一主题文件超过 50 行时考虑拆分