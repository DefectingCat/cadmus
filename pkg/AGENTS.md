<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-04-09 | Updated: 2026-04-09 -->

# pkg - 公共可导出工具包

## Purpose

`pkg/` 目录存放可被外部项目导入的公共 Go 包。这些是稳定、可复用的工具库，提供与 Cadmus 核心业务解耦的通用功能。

## Subdirectories

| Directory | Purpose |
|-----------|---------|
| `utils/` | 通用工具函数 (see `utils/AGENTS.md`) |

## For AI Agents

### 可导出的公共 API

#### pkg/utils

```go
import "cadmus/pkg/utils"
```

| Function | Description |
|----------|-------------|
| `InitTimestamps(created, updated time.Time) (time.Time, time.Time)` | 初始化实体时间戳。零值自动填充为当前时间，非零值保持不变。用于 Repository 层创建/更新实体时统一处理 `CreatedAt`/`UpdatedAt` 字段。 |
| `NormalizeTime(t time.Time) time.Time` | 规范化单个时间字段。零值返回当前时间，否则返回原值。用于单独处理时间字段的场景。 |

### 使用示例

```go
import (
    "time"
    "cadmus/pkg/utils"
)

// 创建实体时初始化时间戳
entity := Post{
    Title: "新文章",
    // CreatedAt 和 UpdatedAt 为零值
}
entity.CreatedAt, entity.UpdatedAt = utils.InitTimestamps(entity.CreatedAt, entity.UpdatedAt)

// 或者链式赋值
post := &Post{Title: "文章"}
post.CreatedAt, post.UpdatedAt = utils.InitTimestamps(post.CreatedAt, post.UpdatedAt)
```

### Working In This Directory

- **添加新工具包**: 在 `pkg/` 下创建新目录，确保包名小写、导出函数大写
- **测试**: `go test ./pkg/...` 或 `make test`
- **构建验证**: `go build ./pkg/...`

### Architecture Notes

- **无状态设计**: 所有工具函数均为纯函数，无内部状态
- **零值处理**: 统一处理 `time.Time` 零值场景，避免数据库插入时的意外行为
- **无外部依赖**: `pkg/` 包不应依赖 `internal/` 或其他业务逻辑代码

<!-- MANUAL: 新增公共包时更新此文件 -->
