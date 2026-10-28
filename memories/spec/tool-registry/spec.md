---
brief: XML 工具的注册、解析与分发规范。
capability: tool-registry
---

# 工具注册表

## Purpose

为 Agent 的 XML 工具提供**单一事实来源**。所有工具的标签名、展示名、
prompt 文档、结果标签集中声明于一处，解析器与 UI 由此派生，
避免"文档说一套、代码做一套"。

## Requirements

### Requirement: 工具集中注册

系统 SHALL 在统一的工具注册表中集中登记所有 XML 工具，
每个工具声明标签名、展示名、描述、prompt 文档、结果标签等元数据。

**Priority**: P0
**Rationale**: 分散声明会导致解析器标签、UI 标签、prompt 文档三处不一致。

#### Scenario: 新增工具

- GIVEN 需要新增一个 XML 工具
- WHEN 在注册表中登记该工具并补齐对应的 prompt 文档
- THEN 解析器的指令标签与结果标签自动生效，无需改动解析器

#### Scenario: 查询工具元数据

- GIVEN 已注册的工具集合
- WHEN 按标签名查询工具
- THEN 返回对应的注册项；未注册时返回空

### Requirement: 解析标签自动派生

系统 SHALL 从工具注册表派生指令标签与结果标签集合，
不得在解析器中手写标签清单。

**Priority**: P0
**Rationale**: 手写清单会随工具增减而漂移，导致解析遗漏。

#### Scenario: 标签集合与注册表一致

- GIVEN 注册表中登记了 N 个工具
- WHEN 检查指令标签集合
- THEN 其内容与注册表的标签集合完全一致

### Requirement: 统一执行分发

系统 SHALL 通过统一入口分发工具执行，
使自动触发与手动执行走同一路径。

**Priority**: P0
**Rationale**: 两条路径若各自实现，行为会分叉。

#### Scenario: 自动与手动执行一致

- GIVEN 同一个工具指令块
- WHEN 分别由自动触发与用户手动执行
- THEN 两者的执行逻辑与结果格式完全相同

## 约束

- 注册表只做声明，不承载业务逻辑
- 解析器与执行器的实现与注册表解耦，避免循环依赖