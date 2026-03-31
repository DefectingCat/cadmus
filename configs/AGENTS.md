<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-03-31 | Updated: 2026-03-31 -->

# configs

## Purpose
配置文件目录，存放应用配置模板。

## Key Files
| File | Description |
|------|-------------|
| `config.example.yaml` | 配置文件模板示例 |
| `README.md` | 配置说明文档 |

## For AI Agents

### Working In This Directory
- 实际配置文件 `config.yaml` 不应提交到仓库（包含敏感信息）
- 新增配置项需同步更新 `config.example.yaml` 和 `README.md`

### Configuration Sections
| Section | Description |
|---------|-------------|
| `server` | 服务端口、主机配置 |
| `database` | PostgreSQL 连接参数 |
| `redis` | Redis 缓存配置 |
| `jwt` | JWT 密钥、过期时间 |
| `email` | SMTP 邮件发送配置 |

<!-- MANUAL: -->