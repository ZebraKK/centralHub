## Docker化项目需要完成的工作清单

### 1. __Docker基础配置__

- [ ] 创建 `Dockerfile` - 多阶段构建，优化镜像大小
- [ ] 创建 `.dockerignore` - 排除不必要的文件（.git, logs, *.log等）
- [ ] 创建 `docker-compose.yml` - 本地开发环境编排（应用 + MongoDB）
- [ ] 创建 `docker-compose.prod.yml` - 生产环境配置

### 2. __开发环境支持__

- [ ] 配置开发环境的卷挂载（支持代码热重载）
- [ ] 配置开发环境网络（应用、数据库互联）
- [ ] 创建开发环境配置文件 `config/config.dev.yaml`
- [ ] Air工具集成（Go代码热重载工具）

### 3. __构建和打包__

- [ ] 创建 `Makefile` - 简化常用命令（build, run, test, clean等）

- [ ] 多阶段构建优化

  - Builder阶段：编译Go应用
  - Runtime阶段：最小化运行时镜像（使用alpine）

- [ ] 版本管理和镜像标签策略

### 4. __测试环境__

- [ ] 创建 `docker-compose.test.yml` - 测试环境配置
- [ ] 配置测试数据库容器
- [ ] 编写测试脚本 `scripts/test.sh`
- [ ] 集成测试环境配置 `config/config.test.yaml`

### 5. __生产环境优化__

- [ ] 生产环境配置 `config/config.prod.yaml`
- [ ] 健康检查端点实现（/health, /ready）
- [ ] 优雅关闭支持（处理SIGTERM信号）
- [ ] 日志配置（JSON格式，适合日志收集）
- [ ] 资源限制配置（CPU、内存）

### 6. __环境变量和密钥管理__

- [ ] 创建 `.env.example` - 环境变量模板
- [ ] Docker secrets支持（生产环境敏感信息）
- [ ] 配置文件与环境变量结合策略

### 7. __CI/CD流程__

- [ ] 创建 `.github/workflows/docker-build.yml` - 自动构建
- [ ] 创建 `.github/workflows/docker-test.yml` - 自动测试
- [ ] 创建 `.github/workflows/docker-publish.yml` - 发布到镜像仓库
- [ ] 配置镜像仓库（Docker Hub / 阿里云 / Harbor）

### 8. __运维脚本__

- [ ] 创建 `scripts/` 目录

  - [ ] `scripts/build.sh` - 构建脚本
  - [ ] `scripts/deploy.sh` - 部署脚本
  - [ ] `scripts/init-db.sh` - 数据库初始化
  - [ ] `scripts/backup.sh` - 备份脚本

### 9. __监控和日志__

- [ ] 集成Prometheus metrics端点（可选）
- [ ] 日志卷挂载配置
- [ ] 日志轮转配置

### 10. __文档更新__

- [ ] 更新 `README.md` - 添加Docker使用说明
- [ ] 创建 `docs/docker.md` - Docker详细文档
- [ ] 创建 `docs/deployment.md` - 部署指南
- [ ] 创建 `CONTRIBUTING.md` - 开发者指南

### 11. __依赖服务配置__

- [ ] MongoDB容器配置（持久化卷、初始化脚本）
- [ ] 网络配置（应用间通信）
- [ ] 数据持久化策略

### 12. __安全加固__

- [ ] 使用非root用户运行容器
- [ ] 镜像扫描（安全漏洞检测）
- [ ] 敏感信息处理
- [ ] 网络安全配置

## 推荐实施顺序

__Phase 1: 基础Docker化（快速启动）__

1. Dockerfile + .dockerignore
2. docker-compose.yml（开发环境）
3. Makefile（简化命令）
4. 更新README

__Phase 2: 完善开发体验__ 5. 热重载支持 6. 测试环境配置 7. 环境变量管理

__Phase 3: 生产就绪__ 8. 生产环境优化 9. 健康检查 10. 日志和监控

__Phase 4: 自动化__ 11. CI/CD流程 12. 部署脚本 13. 运维工具
