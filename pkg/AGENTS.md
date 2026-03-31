<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-03-31 | Updated: 2026-03-31 -->

# pkg

## Purpose
公共可导出包目录，存放可被外部项目引用的工具库和接口定义。

## Key Files
| File | Description |
|------|-------------|
| 当前为空 | 待添加公共接口和工具 |

## Subdirectories
| Directory | Purpose |
|-----------|---------|
| `interfaces/` | 插件接口定义（待创建） |
| `utils/` | 通用工具函数（待创建） |

## For AI Agents

### Working In This Directory
- 此目录下的包可被外部项目 import
- 新增内容需考虑 API 稳定性和文档完善
- **当前状态**: 目录为空，待后续开发填充

### Planned Contents
| Package | Description |
|---------|-------------|
| `interfaces/auth.go` | AuthProvider 接口定义 |
| `interfaces/block.go` | BlockType 接口定义 |
| `interfaces/notify.go` | NotificationChannel 接口定义 |
| `utils/crypto.go` | 密码哈希、UUID 生成 |
| `utils/slug.go` | Slug 生成工具 |

<!-- MANUAL: -->