# centralHub
a manage platform


*TODO* 
- architecture
    - dir/files
- middleware
    - x
    - log
- error

- gin request/respone

- domain struct

- mongo interface


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
