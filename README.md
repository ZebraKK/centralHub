# CentralHub

一个基于 Go 和 Gin 框架的管理平台。

## 功能特性

- 域名管理
- DNS 解析
- ICP 备案
- 与 Volcengine CDN 集成
- RESTful API 接口
- 审计日志
- Docker 支持

## 快速开始

### 方式一：使用 Docker（推荐）

#### 前置要求
- Docker 20.10+
- Docker Compose 2.0+

#### 启动开发环境

1. **复制配置文件**
```bash
cp config/config.dev.yaml config.yaml
```

2. **启动所有服务**
```bash
make dev
# 或者
docker-compose up -d
```

3. **查看日志**
```bash
make logs
# 或者
docker-compose logs -f
```

4. **访问应用**
- 应用地址: http://localhost:8080
- 健康检查: http://localhost:8080/health
- MongoDB: localhost:27017

#### 常用命令

```bash
# 查看所有可用命令
make help

# 启动开发环境
make dev

# 停止所有服务
make stop

# 重启服务
make restart

# 查看日志
make logs

# 查看应用日志
make logs-app

# 查看数据库日志
make logs-db

# 进入应用容器
make shell-app

# 进入数据库容器
make shell-db

# 清理所有容器和卷
make docker-clean
```

### 方式二：本地运行

#### 前置要求
- Go 1.24.3+
- MongoDB 7.0+

#### 安装依赖

```bash
go mod download
```

#### 配置文件

```bash
cp config.yaml.example config.yaml
# 编辑 config.yaml，配置数据库连接等信息
```

#### 运行应用

```bash
# 直接运行
go run main.go

# 或构建后运行
make build
./centralhub

# 指定配置文件
./centralhub -config=/path/to/config.yaml
```

## 配置说明

配置文件支持 YAML 格式，包含以下部分：

- **server**: 服务器配置（端口、模式、超时）
- **database**: 数据库配置（MongoDB 连接）
- **logger**: 日志配置（级别、输出、文件设置）
- **external**: 外部服务配置（Volcengine 凭证）

详细配置说明见 [config/README.md](config/README.md)

## API 接口

### 创建域名
```bash
POST /create
```

### 查询域名
```bash
GET /query
```

### 健康检查
```bash
GET /health
```

## 开发指南

### 目录结构


*TODO* 
- helloworld工程运行起来
    - docker
    - 文档
- architecture
    - dir/files
- middleware
    - x
    - log
- 接口需要
    - error定义
    - http resp结构
    - gin request/respone
    - create/ownership

- domain struct
    - 主要单元
    - 外围（账户/公司）

- mongo interface

- 服务运行起来，日志文件等。
    - to check

- 逻辑伪代码

- workflow 替代

- config 加载(AI)
- http 基础请求封装(AI)


- review&修正AI 代码
    - 啰嗦
    - 验证

- 域名归属验证()
    - 调研流程
    - 实现 验证接口
        - 接口、实现逻辑，流程图



dir 说明

hubServer
    主服务目录
    handle_xxxx， 对应服务提供的对外接口
    mvp版本为单体服务,即hubServer,提供接口,以及接口的逻辑实现,包括和外部服务的交互。
    保持service 的提炼和封装性，后续根据负载情况拆分

model
    结构定义
    域名（整体需要管理、存储的 & domain cdn业务功能块的。参考volc）
    请求接口（request， resp）


store
    主要涉及DB接口
    or 缓存层？

client
    访问外部的client，SDK

service
    包含了多个client + 逻辑，或者sdk+逻辑的
    后期可以按微服务独立出来 

client 和service 紧密联系

middleware
    字面意思
        audit.go 审计日志

workflow
    任务执行的引擎，把管理流程抽象出多个task 和action组合。
    最最初版本, 作为主要逻辑目录。
    完成功能逻辑后，整理成 task和action, 再引入workflow引擎


命名规则遵循简洁、统一、易读的原则 
1. 包名（package）
    - 一般为小写单词，不使用下划线或大写字母。
    - 包名应与目录名一致，且能表达包的主要功能。
    - 例如：package workflow，目录为workflow/
2. 文件名
    - 文件名全部小写，单词间用下划线分隔（如有需要），但推荐直接用功能名。
    - 例如：workflow.go、workflow_service.go。
3. 类型、函数、变量名
    - 使用驼峰命名法（CamelCase），首字母大写表示导出（public），小写为包内可见（private）。
    - 例如：type Workflow struct、func NewWorkflow()。
4. 目录结构
    - 目录名与包名一致，全部小写，无下划线。



// 服务购买
dns 解析
icp
