# Docker 使用指南

本文档详细说明如何使用 Docker 来开发、构建、测试和部署 CentralHub 应用。

## 目录

1. [环境要求](#环境要求)
2. [快速开始](#快速开始)
3. [Docker 配置说明](#docker-配置说明)
4. [开发环境](#开发环境)
5. [构建和部署](#构建和部署)
6. [故障排查](#故障排查)

## 环境要求

- Docker 20.10 或更高版本
- Docker Compose 2.0 或更高版本
- Make 工具（可选，用于简化命令）

### 检查环境

```bash
docker --version
docker-compose --version
make --version
```

## 快速开始

### 1. 准备配置文件

```bash
# 复制开发配置
cp config/config.dev.yaml config.yaml

# 或复制示例配置并编辑
cp config.yaml.example config.yaml
# 编辑 config.yaml 配置你的设置
```

### 2. 启动服务

```bash
# 使用 Makefile（推荐）
make dev

# 或直接使用 docker-compose
docker-compose up -d
```

### 3. 验证服务

```bash
# 检查服务状态
make ps
# 或
docker-compose ps

# 检查健康状态
curl http://localhost:8080/health

# 查看日志
make logs
```

### 4. 停止服务

```bash
make stop
# 或
docker-compose down
```

## Docker 配置说明

### Dockerfile

多阶段构建 Dockerfile：

- **Builder 阶段**：使用 `golang:1.24.3-alpine` 编译应用
- **Runtime 阶段**：使用轻量级 `alpine:latest` 运行应用

特性：
- 最小化镜像大小
- 非 root 用户运行（安全）
- 包含健康检查
- 时区支持（tzdata）
- HTTPS 支持（ca-certificates）

### docker-compose.yml

服务组成：
- **app**: CentralHub 应用服务
- **mongodb**: MongoDB 7.0 数据库

网络：
- `centralhub-network`: 桥接网络，用于服务间通信

卷：
- `mongodb-data`: MongoDB 数据持久化
- `mongodb-config`: MongoDB 配置持久化
- `./logs`: 应用日志目录（挂载到主机）

## 开发环境

### 启动开发环境

```bash
make dev
```

这将启动：
- CentralHub 应用（端口 8080）
- MongoDB 数据库（端口 27017）

### 查看日志

```bash
# 所有服务日志
make logs

# 只看应用日志
make logs-app

# 只看数据库日志
make logs-db
```

### 进入容器

```bash
# 进入应用容器
make shell-app

# 进入数据库容器
make shell-db
```

### 重启服务

```bash
# 重启所有服务
make restart

# 只重启应用
docker-compose restart app

# 只重启数据库
docker-compose restart mongodb
```

### 修改代码后重新构建

当代码修改后，需要重新构建镜像：

```bash
# 停止服务
make stop

# 重新构建并启动
docker-compose up -d --build
```

## 构建和部署

### 构建 Docker 镜像

```bash
# 构建镜像
make docker-build

# 或使用 docker 命令
docker build -t centralhub:latest .

# 构建带版本标签的镜像
docker build -t centralhub:v1.0.0 .
```

### 推送镜像到仓库

```bash
# 登录 Docker Hub
docker login

# 标记镜像
docker tag centralhub:latest username/centralhub:latest
docker tag centralhub:latest username/centralhub:v1.0.0

# 推送镜像
docker push username/centralhub:latest
docker push username/centralhub:v1.0.0
```

### 生产环境部署

生产环境建议：

1. **使用环境特定的配置文件**
```bash
cp config/config.prod.yaml config.yaml
```

2. **使用环境变量覆盖敏感配置**
```bash
# 创建 .env 文件
cp .env.example .env
# 编辑 .env，填入生产环境的值
```

3. **使用生产环境 compose 文件**
```bash
docker-compose -f docker-compose.yml -f docker-compose.prod.yml up -d
```

## 故障排查

### 查看容器状态

```bash
docker-compose ps
```

### 查看容器日志

```bash
# 实时跟踪日志
docker-compose logs -f app

# 查看最近 100 行日志
docker-compose logs --tail=100 app

# 查看特定时间的日志
docker-compose logs --since 2024-01-14T10:00:00 app
```

### 容器无法启动

1. **检查端口占用**
```bash
# 检查 8080 端口
lsof -i :8080
netstat -an | grep 8080

# 检查 27017 端口
lsof -i :27017
```

2. **检查容器日志**
```bash
docker-compose logs app
docker-compose logs mongodb
```

3. **检查配置文件**
```bash
# 确保 config.yaml 存在且格式正确
cat config.yaml
```

### 数据库连接失败

1. **检查 MongoDB 是否健康**
```bash
docker-compose ps mongodb
docker-compose logs mongodb
```

2. **验证连接**
```bash
# 进入应用容器
make shell-app

# 尝试连接 MongoDB
wget -O- mongodb:27017
```

3. **检查网络**
```bash
# 查看网络
docker network ls
docker network inspect centralhub_centralhub-network
```

### 清理和重置

```bash
# 停止并删除所有容器、网络
docker-compose down

# 同时删除卷（会清空数据库数据）
docker-compose down -v

# 完全清理
make docker-clean
```

### 磁盘空间问题

```bash
# 查看 Docker 磁盘使用
docker system df

# 清理未使用的镜像、容器、网络
docker system prune -a

# 清理所有未使用的卷
docker volume prune
```

## 性能优化

### 镜像大小优化

当前 Dockerfile 已使用多阶段构建和 Alpine 基础镜像，镜像大小已经很小。

查看镜像大小：
```bash
docker images centralhub
```

### 容器资源限制

在 docker-compose.yml 中添加资源限制：

```yaml
services:
  app:
    deploy:
      resources:
        limits:
          cpus: '1.0'
          memory: 512M
        reservations:
          cpus: '0.5'
          memory: 256M
```

## 安全建议

1. **使用非 root 用户**（已在 Dockerfile 中实现）
2. **不要在镜像中包含敏感信息**
3. **使用 Docker secrets 管理敏感数据**
4. **定期更新基础镜像**
5. **扫描镜像漏洞**

```bash
# 使用 Docker scan（需要登录）
docker scan centralhub:latest
```

## 常见问题

### Q: 如何更新应用代码？
A: 修改代码后运行 `docker-compose up -d --build`

### Q: 如何备份数据库？
A: 使用 `docker-compose exec mongodb mongodump` 或挂载卷进行备份

### Q: 如何查看应用性能？
A: 使用 `docker stats` 查看实时资源使用情况

### Q: 开发时如何实现热重载？
A: 可以使用卷挂载源代码，并在容器中使用 Air 等热重载工具（见 Phase 2）

## 下一步

- 配置 CI/CD 流程
- 添加监控和日志收集
- 实现自动化测试
- 配置生产环境部署流程

更多信息请参考：
- [配置说明](../config/README.md)
- [项目 README](../README.md)
