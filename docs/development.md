# 开发环境指南

本文档详细说明如何使用 Docker 进行 CentralHub 项目的开发。

## 目录

1. [开发模式](#开发模式)
2. [热重载开发](#热重载开发)
3. [配置管理](#配置管理)
4. [常见工作流](#常见工作流)
5. [故障排查](#故障排查)

---

## 开发模式

CentralHub 提供两种 Docker 开发模式：

### 1. 标准开发模式（生产类似）

使用生产环境相同的镜像，适合测试生产环境行为。

```bash
# 启动
make dev

# 查看日志
make logs

# 停止
make stop
```

**特点**：
- ✅ 与生产环境一致
- ✅ 镜像优化（106MB）
- ❌ 代码修改需要重新构建

### 2. 热重载开发模式（推荐）

使用 Air 工具实现代码热重载，代码修改后自动重启应用。

```bash
# 启动热重载环境
make dev-hot

# 查看日志
make dev-hot-logs

# 停止
make dev-hot-down
```

**特点**：
- ✅ 代码自动重载
- ✅ 快速开发迭代
- ✅ 实时查看修改效果
- ❌ 镜像较大（包含开发工具）

---

## 热重载开发

### Air 配置

热重载使用 [Air](https://github.com/air-verse/air) 工具，配置文件位于 `.air.toml`。

#### 主要配置

```toml
[build]
  cmd = "go build -o ./tmp/main ."
  bin = "./tmp/main"
  delay = 1000  # 延迟1秒后重启
  exclude_dir = ["assets", "tmp", "vendor", "logs", "docs"]
  include_ext = ["go", "tpl", "tmpl", "html"]
```

### 工作原理

1. **监控文件变化**：Air 监控 `.go` 文件的变化
2. **自动编译**：检测到变化后自动编译应用
3. **自动重启**：编译成功后自动重启应用
4. **查看日志**：实时输出应用日志

### 开发流程

```bash
# 1. 启动热重载环境
make dev-hot

# 2. 修改代码
# 编辑任意 .go 文件

# 3. 自动重载
# Air 自动检测变化、编译、重启

# 4. 查看效果
curl http://localhost:8080/health

# 5. 查看日志
make dev-hot-logs
```

### 文件监控

Air 会监控以下文件类型的变化：
- `.go` - Go 源代码
- `.html` - HTML 模板
- `.tpl`, `.tmpl` - 其他模板文件

排除目录：
- `tmp/` - 临时构建目录
- `vendor/` - 依赖目录
- `logs/` - 日志目录
- `docs/` - 文档目录

---

## 配置管理

### 环境配置文件

项目提供多个环境的配置文件：

```
config/
├── config.go           # 配置加载逻辑
├── config.dev.yaml     # 开发环境
├── config.test.yaml    # 测试环境
├── config.yaml.example # 配置模板
└── README.md          # 配置文档
```

### 使用配置文件

#### 方式一：复制开发配置

```bash
cp config/config.dev.yaml config.yaml
```

#### 方式二：使用 Make 命令

```bash
make setup-config
```

#### 方式三：自定义配置

```bash
cp config.yaml.example config.yaml
# 编辑 config.yaml，填入你的配置
```

### 配置优先级

1. 命令行参数 `-config=/path/to/config.yaml`
2. 默认配置文件 `config.yaml`
3. 代码中的默认值

---

## 常见工作流

### 工作流 1: 新功能开发

```bash
# 1. 启动热重载环境
make dev-hot

# 2. 创建新功能分支
git checkout -b feature/new-feature

# 3. 编写代码
# 修改 .go 文件，Air 自动重载

# 4. 测试功能
curl http://localhost:8080/your-endpoint

# 5. 查看日志确认
make dev-hot-logs

# 6. 提交代码
git add .
git commit -m "Add new feature"
```

### 工作流 2: Bug 修复

```bash
# 1. 启动热重载环境
make dev-hot

# 2. 复现 Bug
curl http://localhost:8080/buggy-endpoint

# 3. 查看日志
make dev-hot-logs

# 4. 修复代码
# 编辑相关文件

# 5. 验证修复
curl http://localhost:8080/buggy-endpoint

# 6. 查看日志确认
make dev-hot-logs
```

### 工作流 3: API 测试

```bash
# 1. 启动环境
make dev-hot

# 2. 使用 curl 测试
curl -X POST http://localhost:8080/create \
  -H "Content-Type: application/json" \
  -d '{"domain": "example.com"}'

# 3. 查看应用日志
make dev-hot-logs

# 4. 查看数据库数据
make shell-db
# 在 mongosh 中查询数据
db.domains.find()
```

### 工作流 4: 数据库交互

```bash
# 1. 启动环境
make dev-hot

# 2. 连接数据库
make shell-db

# 3. 在 mongosh 中操作
use centralhub
db.domains.find()
db.domains.insertOne({name: "test.com"})

# 4. 在应用中验证
curl http://localhost:8080/query
```

---

## 故障排查

### 问题 1: 热重载不工作

**症状**：修改代码后应用没有重启

**解决方案**：

```bash
# 1. 检查 Air 是否运行
docker-compose -f docker-compose.yml -f docker-compose.dev.yml logs app

# 2. 确认文件修改在监控范围内
# 查看 .air.toml 中的 include_ext

# 3. 重启环境
make dev-hot-down
make dev-hot
```

### 问题 2: 端口冲突

**症状**：启动失败，提示端口已被占用

**解决方案**：

```bash
# 检查端口占用
lsof -i :8080
lsof -i :27017

# 停止占用端口的进程或修改配置文件
```

### 问题 3: 编译错误

**症状**：Air 报告编译错误

**解决方案**：

```bash
# 查看错误详情
cat build-errors.log

# 在本地测试编译
go build .

# 修复错误后，Air 会自动重试
```

### 问题 4: 容器无法访问

**症状**：`curl http://localhost:8080/health` 失败

**解决方案**：

```bash
# 1. 检查容器状态
docker-compose -f docker-compose.yml -f docker-compose.dev.yml ps

# 2. 查看容器日志
make dev-hot-logs

# 3. 检查健康状态
docker-compose -f docker-compose.yml -f docker-compose.dev.yml ps

# 4. 重启容器
make dev-hot-restart
```

### 问题 5: 数据库连接失败

**症状**：应用无法连接 MongoDB

**解决方案**：

```bash
# 1. 检查 MongoDB 容器
docker-compose ps mongodb

# 2. 查看 MongoDB 日志
make logs-db

# 3. 验证连接字符串
# 确认 config.yaml 中的 MongoDB URI 正确
# mongodb://admin:admin123@mongodb:27017

# 4. 重启 MongoDB
docker-compose restart mongodb
```

---

## 性能优化

### 开发环境优化建议

1. **使用 SSD**：Docker 卷挂载在 SSD 上性能更好
2. **分配足够资源**：Docker Desktop 至少分配 4GB 内存
3. **排除不必要的文件**：`.dockerignore` 和 `.air.toml` 中排除大文件
4. **使用缓存**：利用 Docker 层缓存加速构建

### Air 性能优化

在 `.air.toml` 中调整：

```toml
[build]
  delay = 500  # 减少延迟到 500ms（如果系统性能好）
  exclude_unchanged = true  # 只重新编译修改的包
```

---

## 调试技巧

### 1. 使用日志调试

```go
// 在代码中添加调试日志
logger.RunLogger.Debug().
    Str("variable", value).
    Msg("Debug message")
```

### 2. 使用断点（Delve）

虽然热重载环境不支持交互式调试，但可以：

```bash
# 停止热重载环境
make dev-hot-down

# 本地运行并使用 Delve
dlv debug -- -config=config.yaml
```

### 3. 查看环境变量

```bash
# 进入容器查看环境变量
docker-compose -f docker-compose.yml -f docker-compose.dev.yml exec app env
```

---

## 最佳实践

### 1. 代码组织

- 保持函数简短易测试
- 使用接口实现依赖注入
- 遵循项目命名规范

### 2. 配置管理

- 敏感信息使用环境变量
- 不同环境使用不同配置文件
- 配置文件不提交到 Git

### 3. 日志记录

- 使用结构化日志（zerolog）
- 合理的日志级别（Debug/Info/Warn/Error）
- 包含足够的上下文信息

### 4. 数据库操作

- 开发环境使用 Docker MongoDB
- 定期备份测试数据
- 使用迁移脚本管理 schema 变更

---

## 相关文档

- [Docker 使用指南](docker.md)
- [配置说明](../config/README.md)
- [项目 README](../README.md)
- [Phase 1 测试报告](PHASE1_TEST_REPORT.md)

---

**更新时间**: 2026-01-14  
**适用版本**: Phase 2+
