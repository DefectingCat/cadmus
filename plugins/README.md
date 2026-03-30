# Cadmus 插件目录

此目录存放 Cadmus 平台的插件。

## 插件结构

每个插件应包含：
```
plugin-name/
├── plugin.yaml    # 插件元数据
├── main.go        # 插件入口（可选）
└── hooks/         # 钩子实现
```