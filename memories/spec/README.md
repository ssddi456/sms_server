---
brief: 规范真相源目录（memories/spec/）的定位、目录约定与写作规范。
---

# 规范真相源（memories/spec/）

> 本目录承载「系统**当前**如何工作」的权威描述。
> 与 `memories/` 下其他文件的区别：其他文件记录**决策与经验**（为什么这样选、踩过什么坑），
> 本目录记录**行为契约**（现在必须满足什么）。

## 一、目录结构

```
memories/spec/
├── README.md                  # 本文件
├── <capability>/              # 一个能力 = 一个目录
│   ├── spec.md                # 主规范（唯一权威）
│   └── design.md              # （可选）已确立的实现模式
└── changes/                   # 进行中的变更
    ├── <change-id>/
    │   ├── proposal.md        # 为什么做 / 做什么 / 影响面
    │   ├── specs/
    │   │   └── <capability>.md  # Delta Spec（增量）
    │   └── meta.json          # 状态元数据
    └── archive/               # 已完成变更
        └── YYYY-MM-DD-<change-id>/
```

## 二、命名约定

| 项 | 约定 | 示例 |
|---|---|---|
| `capability` | kebab-case，**名词** | `tool-registry`、`memory-system` |
| `change-id` | kebab-case，**动词开头** | `add-spec-validate`、`update-tool-registry` |
| 动词前缀 | `add-` / `update-` / `remove-` / `refactor-` | — |

字符集限制：仅 `[a-z0-9-]`，长度 ≤ 64，禁止 `..` 与前导 `-`。

## 三、主规范格式

```markdown
---
brief: 一句话说明该能力
capability: <capability>
---

# <能力名>

## Purpose
（这个能力存在的理由，1-3 句）

## Requirements

### Requirement: <需求名>
系统 SHALL <可验证的行为声明>。

**Priority**: P0
**Rationale**: （为什么需要它）

#### Scenario: <场景名>
- GIVEN <前置条件>
- WHEN <触发动作>
- THEN <可观察的结果>
```

## 四、写作规范

| # | 规则 | 级别 |
|---|---|---|
| 1 | 每个 `### Requirement:` 至少包含 1 个 `#### Scenario:` | ERROR |
| 2 | 每个 Scenario 必须含 GIVEN / WHEN / THEN 三行 | ERROR |
| 3 | Requirement 正文必须含 RFC 2119 关键词（`SHALL` 或 `MUST`） | ERROR |
| 4 | Scenario 标题使用 4 个井号（`####`） | WARN |
| 5 | Requirement 标题使用 3 个井号（`###`） | WARN |

**核心原则**：规范描述**可观察的行为契约**，而非**内部实现方案**。
- ✅ 「系统 SHALL 在合并前校验规范格式」
- ❌ 「在 spec_service.go 的 ApplyChange 函数里调用 Validate」

实现细节放 `design.md` 或 `docs/`。

## 五、与相邻目录的边界

| 目录 | 管什么 | 不管什么 |
|---|---|---|
| `memories/spec/` | **当前行为契约**（现在必须满足什么） | 决策过程、历史沿革 |
| `memories/repo/` | 仓库级决策与经验 | 行为契约 |
| `memories/project/` | 项目主题记忆 | 行为契约 |
| `docs/` | 设计文档、调研报告 | 权威行为契约 |

**关键约束**：同一事实**只允许有一处权威描述**。若某行为已在 `spec/` 中定义，
其他位置应引用而非复述。

## 六、使用方式

- **读**：直接用 `read_memory`（路径如 `memories/spec/<capability>/spec.md`）
- **检索**：`memory_search` 可命中（规范位于 `memories/` 下，自动进入检索索引）
- **写**：可直接编辑，或通过变更流程（`changes/`）提交增量后再合并