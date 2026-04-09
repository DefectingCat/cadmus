<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-04-09 | Updated: 2026-04-09 -->

# pkg/utils - 公共时间工具包

## Purpose

`pkg/utils` 提供可复用的时间处理工具函数，专注于 `time.Time` 零值的规范化处理。这些工具函数用于实体创建/更新时统一处理时间戳字段，避免数据库操作中的意外行为。

## Key Files

| File | Description |
|------|-------------|
| `time.go` | 时间处理工具函数：`InitTimestamps` 和 `NormalizeTime` |
| `time_test.go` | 完整的单元测试覆盖 |

## Public API

### InitTimestamps

```go
func InitTimestamps(created, updated time.Time) (time.Time, time.Time)
```

初始化实体时间戳。零值自动填充为当前时间，非零值保持不变。

**使用场景**：Repository 层创建或更新实体时批量处理 `CreatedAt` 和 `UpdatedAt` 字段。

**使用示例**：
```go
import "cadmus/pkg/utils"

entity := Post{Title: "新文章"}
entity.CreatedAt, entity.UpdatedAt = utils.InitTimestamps(entity.CreatedAt, entity.UpdatedAt)
```

### NormalizeTime

```go
func NormalizeTime(t time.Time) time.Time
```

规范化单个时间字段。零值返回当前时间，否则返回原值。

**使用场景**：单独处理某个时间字段的场景。

**使用示例**：
```go
import "cadmus/pkg/utils"

publishedAt := utils.NormalizeTime(post.PublishedAt)
```

## For AI Agents

### 何时使用

- 创建新实体需要设置时间戳时 → 使用 `InitTimestamps`
- 单独规范化某个时间字段时 → 使用 `NormalizeTime`
- 需要确保时间字段不为零值时 → 两个函数均可

### 测试验证

```bash
# 运行 utils 包测试
go test ./pkg/utils/...

# 或使用 make
make test
```

### 设计原则

- **无状态**：所有函数均为纯函数，无内部状态
- **零值安全**：统一处理 `time.Time` 零值，避免数据库插入时的意外行为
- **无外部依赖**：仅依赖标准库 `time` 包
- **测试覆盖**：所有函数均有完整单元测试

<!-- MANUAL: 新增工具函数时更新此文件 -->
