# centralHub
a manage platform


*TODO* 
- architecture
    - dir/files
- middleware
    - x
    - log
- error


dir 说明

hubServer
    主服务目录
    handle_xxxx， 对应服务提供的对外接口

models
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
    