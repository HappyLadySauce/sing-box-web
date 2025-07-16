# 后端代码结构方案 - sing-box-web

## 1. 项目整体目录结构

基于 Go 项目最佳实践和《Standard Go Project Layout》，结合 sing-box-web 项目的三层架构特点，我们采用以下目录结构：

```
sing-box-web/
├── cmd/                            # 应用程序入口点
│   ├── sing-box-web/               # Web前端服务器
│   ├── sing-box-api/               # API核心服务器
│   └── sing-box-agent/             # 节点代理程序
├── pkg/                            # 公共库代码
├── internal/                       # 内部应用代码
├── api/                            # API 定义和生成代码
├── configs/                        # 配置文件
├── deployments/                    # 部署相关文件
├── docs/                           # 文档
├── scripts/                        # 脚本文件
├── test/                           # 额外的测试数据
├── ui/                             # 前端代码（已存在）
├── tools/                          # 工具和工具配置
├── hack/                           # 构建和开发工具
├── build/                          # 构建输出目录
├── .github/                        # GitHub 工作流
├── go.mod                          # Go 模块定义
├── go.sum                          # Go 模块校验和
├── Makefile                        # 构建系统
├── Dockerfile                      # 容器构建文件
├── docker-compose.yml              # 本地开发环境
└── README.md                       # 项目说明
```

---

## 2. 应用架构与职责划分

### 2.1 架构概览

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   sing-box-web  │    │  sing-box-api   │    │ sing-box-agent  │
│  (前端服务器)    │◄──►│  (核心API服务)   │◄──►│   (节点代理)     │
│                 │    │                 │    │                 │
│ • 用户界面       │    │ • 节点管理       │    │ • sing-box控制   │
│ • 身份认证       │    │ • 配置管理       │    │ • 系统监控       │
│ • 操作审计       │    │ • 监控数据       │    │ • Clash API代理  │
│ • API代理        │    │ • 业务逻辑       │    │ • 配置应用       │
│ • WebSocket     │    │ • gRPC服务       │    │ • 进程管理       │
└─────────────────┘    └─────────────────┘    └─────────────────┘
       │                         │                         │
       ▼                         ▼                         ▼
    Port 3000               Port 8080                 动态端口
   HTTP/WebSocket            gRPC/HTTP              gRPC Client
```

### 2.2 职责分工

| 应用 | 主要职责 | 技术栈 | 端口 |
|------|---------|--------|------|
| **sing-box-web** | 前端服务、用户认证、操作日志、API代理 | Gin + WebSocket | 3000 |
| **sing-box-api** | 节点管理、配置管理、数据存储、gRPC服务 | Gin + gRPC + GORM | 8080/9090 |
| **sing-box-agent** | sing-box控制、监控收集、Clash API | gRPC Client + HTTP Proxy | 动态 |

---

## 3. 核心应用代码结构

### 3.1 Web 服务器 (cmd/sing-box-web)

```
cmd/sing-box-web/
├── main.go                         # 应用入口点
├── app/                            # 应用逻辑层
│   ├── options/                    # 命令行选项
│   │   ├── options.go              # Web服务器选项
│   │   └── validation.go           # 选项验证
│   ├── server/                     # 服务器逻辑
│   │   ├── server.go               # Web服务器主逻辑
│   │   ├── handlers/               # HTTP处理器
│   │   │   ├── auth.go             # 认证处理器
│   │   │   ├── users.go            # 用户管理处理器
│   │   │   ├── audit.go            # 操作日志处理器
│   │   │   ├── proxy.go            # API代理处理器
│   │   │   ├── health.go           # 健康检查处理器
│   │   │   └── websocket.go        # WebSocket处理器
│   │   ├── middleware/             # 中间件
│   │   │   ├── auth.go             # JWT认证中间件
│   │   │   ├── audit.go            # 审计日志中间件
│   │   │   ├── cors.go             # CORS中间件
│   │   │   ├── security.go         # 安全中间件
│   │   │   ├── logging.go          # 日志中间件
│   │   │   └── proxy.go            # 代理中间件
│   │   └── routes.go               # 路由定义
│   ├── config/                     # 配置管理
│   │   ├── config.go               # Web配置结构
│   │   └── loader.go               # 配置加载器
│   └── assets/                     # 静态资源管理
│       ├── embed.go                # 嵌入式前端资源
│       └── handler.go              # 静态资源处理器
└── Dockerfile                      # Web服务器容器构建
```

#### main.go 入口文件示例

```go
// cmd/sing-box-web/main.go
package main

import (
    "context"
    "fmt"
    "os"
    "os/signal"
    "syscall"

    "github.com/spf13/cobra"
    "sing-box-web/cmd/sing-box-web/app"
    "sing-box-web/cmd/sing-box-web/app/options"
    "sing-box-web/pkg/version"
)

func main() {
    cmd := newWebServerCommand()
    if err := cmd.Execute(); err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
}

func newWebServerCommand() *cobra.Command {
    opts := options.NewWebServerOptions()

    cmd := &cobra.Command{
        Use:   "sing-box-web",
        Short: "sing-box web frontend server",
        Long:  `Frontend server for sing-box-web management platform`,
        RunE: func(cmd *cobra.Command, args []string) error {
            if err := opts.Validate(); err != nil {
                return err
            }
            return runWebServer(opts)
        },
    }

    // 添加子命令
    cmd.AddCommand(newServeCommand(opts))
    cmd.AddCommand(newVersionCommand())

    return cmd
}

func runWebServer(opts *options.WebServerOptions) error {
    ctx, cancel := signal.NotifyContext(context.Background(), 
        syscall.SIGINT, syscall.SIGTERM)
    defer cancel()

    server, err := app.NewWebServer(opts)
    if err != nil {
        return fmt.Errorf("failed to create web server: %w", err)
    }

    return server.Run(ctx)
}
```

### 3.2 API 服务器 (cmd/sing-box-api)

```
cmd/sing-box-api/
├── main.go                         # 应用入口点
├── app/                            # 应用逻辑层
│   ├── options/                    # 命令行选项
│   │   ├── options.go              # API服务器选项
│   │   └── validation.go           # 选项验证
│   ├── server/                     # 服务器逻辑
│   │   ├── server.go               # API服务器主逻辑
│   │   ├── grpc.go                 # gRPC服务器
│   │   ├── http.go                 # HTTP REST服务器
│   │   ├── metrics.go              # 监控指标服务器
│   │   └── manager.go              # Agent连接管理器
│   ├── config/                     # 配置管理
│   │   ├── config.go               # API配置结构
│   │   ├── loader.go               # 配置加载器
│   │   └── validator.go            # 配置验证器
│   └── version/                    # 版本信息
│       ├── version.go              # 版本定义
│       └── build_info.go           # 构建信息
└── Dockerfile                      # API服务器容器构建
```

### 3.3 Agent 应用 (cmd/sing-box-agent)

```
cmd/sing-box-agent/
├── main.go                         # 应用入口点
├── app/                            # 应用逻辑层
│   ├── options/                    # 命令行选项
│   │   ├── options.go              # Agent选项
│   │   └── validation.go           # 选项验证
│   ├── agent/                      # Agent主逻辑
│   │   ├── agent.go                # Agent主体
│   │   ├── connection.go           # gRPC连接管理
│   │   ├── heartbeat.go            # 心跳管理
│   │   └── registry.go             # 节点注册逻辑
│   ├── monitor/                    # 监控模块
│   │   ├── collector.go            # 监控数据收集器
│   │   ├── system.go               # 系统监控
│   │   ├── singbox.go              # sing-box监控
│   │   └── reporter.go             # 监控数据上报
│   ├── config/                     # 配置管理
│   │   ├── manager.go              # 配置管理器
│   │   ├── watcher.go              # 配置文件监控
│   │   └── validator.go            # 配置验证器
│   ├── singbox/                    # sing-box控制
│   │   ├── controller.go           # sing-box控制器
│   │   ├── systemctl.go            # systemctl操作
│   │   ├── process.go              # 进程管理
│   │   └── config.go               # 配置文件操作
│   └── clash/                      # Clash API集成
│       ├── proxy.go                # Clash API代理
│       ├── client.go               # HTTP客户端
│       └── handlers.go             # API处理器
└── Dockerfile                      # Agent容器构建
```

---

## 4. 内部应用代码 (internal/)

```
internal/
├── web/                            # Web服务器内部代码
│   ├── service/                    # 业务服务层
│   │   ├── interfaces.go           # 服务接口定义
│   │   ├── auth/                   # 认证服务
│   │   │   ├── service.go          # 认证服务实现
│   │   │   ├── jwt.go              # JWT处理
│   │   │   └── session.go          # 会话管理
│   │   ├── user/                   # 用户管理服务
│   │   │   ├── service.go          # 用户服务实现
│   │   │   ├── manager.go          # 用户管理器
│   │   │   └── validator.go        # 用户验证器
│   │   ├── audit/                  # 审计日志服务
│   │   │   ├── service.go          # 审计服务实现
│   │   │   ├── logger.go           # 日志记录器
│   │   │   └── exporter.go         # 日志导出器
│   │   └── proxy/                  # API代理服务
│   │       ├── service.go          # 代理服务实现
│   │       ├── client.go           # gRPC客户端
│   │       └── converter.go        # 数据转换器
│   ├── repository/                 # 数据访问层
│   │   ├── interfaces.go           # 仓库接口定义
│   │   ├── user/                   # 用户数据仓库
│   │   │   ├── repository.go       # 用户仓库实现
│   │   │   └── model.go            # 用户数据模型
│   │   └── audit/                  # 审计日志仓库
│   │       ├── repository.go       # 审计仓库实现
│   │       └── model.go            # 审计数据模型
│   ├── websocket/                  # WebSocket服务
│   │   ├── hub.go                  # 连接中心
│   │   ├── client.go               # 客户端连接
│   │   ├── message.go              # 消息处理
│   │   └── broadcast.go            # 广播服务
│   └── cache/                      # 缓存层
│       ├── interfaces.go           # 缓存接口定义
│       ├── session.go              # 会话缓存
│       └── user.go                 # 用户缓存
├── api/                            # API服务器内部代码
│   ├── server/                     # 服务器实现
│   │   ├── grpc/                   # gRPC服务实现
│   │   │   ├── server.go           # gRPC服务器
│   │   │   ├── web_service.go      # Web服务接口实现
│   │   │   ├── agent_service.go    # Agent服务接口实现
│   │   │   ├── interceptors/       # gRPC拦截器
│   │   │   │   ├── auth.go         # 认证拦截器
│   │   │   │   ├── logging.go      # 日志拦截器
│   │   │   │   ├── metrics.go      # 监控拦截器
│   │   │   │   └── recovery.go     # 恢复拦截器
│   │   │   └── handlers/           # gRPC处理器
│   │   │       ├── node.go         # 节点管理处理
│   │   │       ├── config.go       # 配置管理处理
│   │   │       ├── deployment.go   # 部署管理处理
│   │   │       └── metrics.go      # 监控数据处理
│   │   ├── http/                   # HTTP/REST API实现（内部管理）
│   │   │   ├── server.go           # HTTP服务器
│   │   │   ├── router.go           # 路由配置
│   │   │   ├── middleware/         # HTTP中间件
│   │   │   │   ├── cors.go         # CORS中间件
│   │   │   │   ├── logging.go      # 日志中间件
│   │   │   │   └── recovery.go     # 恢复中间件
│   │   │   └── handlers/           # HTTP处理器
│   │   │       ├── health.go       # 健康检查
│   │   │       └── metrics.go      # 监控指标
│   │   └── manager/                # Agent连接管理
│   │       ├── manager.go          # 连接管理器
│   │       ├── registry.go         # 节点注册表
│   │       ├── session.go          # Agent会话
│   │       └── pool.go             # 连接池
│   ├── service/                    # 业务服务层
│   │   ├── interfaces.go           # 服务接口定义
│   │   ├── node/                   # 节点管理服务
│   │   │   ├── service.go          # 节点服务实现
│   │   │   ├── manager.go          # 节点管理器
│   │   │   ├── registry.go         # 节点注册
│   │   │   └── validator.go        # 节点验证器
│   │   ├── config/                 # 配置管理服务
│   │   │   ├── service.go          # 配置服务实现
│   │   │   ├── template.go         # 模板管理
│   │   │   ├── deployment.go       # 部署管理
│   │   │   ├── validator.go        # 配置验证
│   │   │   └── generator.go        # 配置生成器
│   │   ├── monitoring/             # 监控服务
│   │   │   ├── service.go          # 监控服务实现
│   │   │   ├── collector.go        # 数据收集
│   │   │   ├── aggregator.go       # 数据聚合
│   │   │   ├── alerting.go         # 告警处理
│   │   │   └── dashboard.go        # 仪表盘数据
│   │   ├── notification/           # 通知服务
│   │   │   ├── service.go          # 通知服务实现
│   │   │   ├── email.go            # 邮件通知
│   │   │   ├── webhook.go          # Webhook通知
│   │   │   └── channels.go         # 通知渠道
│   │   └── clash/                  # Clash API服务
│   │       ├── service.go          # Clash API服务
│   │       ├── proxy.go            # API代理
│   │       └── manager.go          # Clash管理器
│   ├── repository/                 # 数据访问层
│   │   ├── interfaces.go           # 仓库接口定义
│   │   ├── node/                   # 节点数据仓库
│   │   │   ├── repository.go       # 节点仓库实现
│   │   │   └── model.go            # 节点数据模型
│   │   ├── config/                 # 配置数据仓库
│   │   │   ├── repository.go       # 配置仓库实现
│   │   │   ├── template.go         # 模板仓库
│   │   │   ├── deployment.go       # 部署仓库
│   │   │   └── model.go            # 配置数据模型
│   │   ├── metrics/                # 监控数据仓库
│   │   │   ├── repository.go       # 监控仓库实现
│   │   │   ├── time_series.go      # 时序数据
│   │   │   └── model.go            # 监控数据模型
│   │   └── migration/              # 数据库迁移
│   │       ├── migrator.go         # 迁移器
│   │       └── versions/           # 迁移版本
│   │           ├── 001_initial.go  # 初始化迁移
│   │           └── 002_labels.go   # 添加标签迁移
│   └── cache/                      # 缓存层
│       ├── interfaces.go           # 缓存接口定义
│       ├── redis/                  # Redis缓存实现
│       │   ├── client.go           # Redis客户端
│       │   ├── session.go          # 会话缓存
│       │   ├── metrics.go          # 监控数据缓存
│       │   ├── config.go           # 配置缓存
│       │   └── lock.go             # 分布式锁
│       └── memory/                 # 内存缓存实现
│           ├── cache.go            # 内存缓存
│           └── lru.go              # LRU缓存
├── agent/                          # Agent内部代码
│   ├── client/                     # gRPC客户端
│   │   ├── client.go               # 客户端实现
│   │   ├── connection.go           # 连接管理
│   │   ├── retry.go                # 重试逻辑
│   │   └── stream.go               # 流处理
│   ├── monitor/                    # 系统监控
│   │   ├── system/                 # 系统监控实现
│   │   │   ├── cpu.go              # CPU监控
│   │   │   ├── memory.go           # 内存监控
│   │   │   ├── disk.go             # 磁盘监控
│   │   │   └── network.go          # 网络监控
│   │   ├── singbox/                # sing-box监控
│   │   │   ├── status.go           # 状态监控
│   │   │   ├── metrics.go          # 指标采集
│   │   │   ├── logs.go             # 日志监控
│   │   │   └── health.go           # 健康检查
│   │   └── collector.go            # 监控数据收集器
│   ├── config/                     # 配置管理
│   │   ├── manager.go              # 配置管理器
│   │   ├── applier.go              # 配置应用器
│   │   ├── validator.go            # 配置验证器
│   │   ├── watcher.go              # 文件监控
│   │   └── backup.go               # 配置备份
│   ├── singbox/                    # sing-box控制
│   │   ├── controller.go           # sing-box控制器
│   │   ├── systemctl.go            # systemctl操作
│   │   ├── process.go              # 进程管理
│   │   ├── config.go               # 配置操作
│   │   ├── health.go               # 健康检查
│   │   └── installer.go            # 安装管理
│   └── clash/                      # Clash API集成
│       ├── proxy.go                # API代理服务
│       ├── client.go               # HTTP客户端
│       ├── handlers/               # API处理器
│       │   ├── proxies.go          # 代理管理
│       │   ├── configs.go          # 配置管理
│       │   ├── connections.go      # 连接管理
│       │   ├── logs.go             # 日志处理
│       │   └── traffic.go          # 流量统计
│       ├── websocket.go            # WebSocket代理
│       └── middleware.go           # 中间件
└── shared/                         # 内部共享代码
    ├── errors/                     # 错误定义
    │   ├── codes.go                # 错误代码
    │   ├── errors.go               # 错误类型
    │   └── handler.go              # 错误处理器
    ├── constants/                  # 常量定义
    │   ├── status.go               # 状态常量
    │   ├── events.go               # 事件常量
    │   ├── clash.go                # Clash API常量
    │   └── metrics.go              # 监控常量
    ├── utils/                      # 工具函数
    │   ├── crypto.go               # 加密工具
    │   ├── json.go                 # JSON工具
    │   ├── time.go                 # 时间工具
    │   ├── network.go              # 网络工具
    │   └── validation.go           # 验证工具
    └── types/                      # 共享类型
        ├── pagination.go           # 分页类型
        ├── filter.go               # 过滤器类型
        ├── response.go             # 响应类型
        └── clash.go                # Clash API类型
```

---

## 5. 公共库代码 (pkg/)

```
pkg/
├── version/                        # 版本信息
│   ├── version.go                  # 版本定义
│   └── build.go                    # 构建信息
├── logger/                         # 统一日志
│   ├── logger.go                   # 日志接口
│   ├── logrus.go                   # Logrus实现
│   ├── config.go                   # 日志配置
│   └── formatter.go                # 日志格式化器
├── config/                         # 配置管理
│   ├── config.go                   # 配置接口
│   ├── loader.go                   # 配置加载器
│   ├── watcher.go                  # 配置监控器
│   └── validator.go                # 配置验证器
├── database/                       # 数据库
│   ├── postgres/                   # PostgreSQL驱动
│   │   ├── client.go               # 客户端
│   │   ├── config.go               # 配置
│   │   └── migration.go            # 迁移工具
│   ├── sqlite/                     # SQLite驱动
│   │   ├── client.go               # 客户端
│   │   └── config.go               # 配置
│   └── interfaces.go               # 数据库接口
├── cache/                          # 缓存
│   ├── redis/                      # Redis客户端
│   │   ├── client.go               # 客户端实现
│   │   ├── config.go               # 配置
│   │   └── cluster.go              # 集群支持
│   ├── memory/                     # 内存缓存
│   │   └── cache.go                # 内存缓存实现
│   └── interfaces.go               # 缓存接口
├── metrics/                        # 监控指标
│   ├── prometheus/                 # Prometheus指标
│   │   ├── metrics.go              # 指标定义
│   │   ├── collector.go            # 指标收集器
│   │   └── middleware.go           # 中间件
│   ├── system/                     # 系统指标收集
│   │   ├── cpu.go                  # CPU指标
│   │   ├── memory.go               # 内存指标
│   │   ├── disk.go                 # 磁盘指标
│   │   └── network.go              # 网络指标
│   └── interfaces.go               # 监控接口
├── grpc/                           # gRPC通用组件
│   ├── server/                     # gRPC服务器
│   │   ├── server.go               # 服务器实现
│   │   ├── config.go               # 服务器配置
│   │   └── interceptors.go         # 拦截器
│   ├── client/                     # gRPC客户端
│   │   ├── client.go               # 客户端实现
│   │   ├── config.go               # 客户端配置
│   │   ├── pool.go                 # 连接池
│   │   └── retry.go                # 重试逻辑
│   └── middleware/                 # 中间件
│       ├── auth.go                 # 认证中间件
│       ├── logging.go              # 日志中间件
│       └── metrics.go              # 监控中间件
├── http/                           # HTTP通用组件
│   ├── server/                     # HTTP服务器
│   │   ├── server.go               # 服务器实现
│   │   ├── config.go               # 服务器配置
│   │   └── graceful.go             # 优雅关闭
│   ├── client/                     # HTTP客户端
│   │   ├── client.go               # 客户端实现
│   │   ├── config.go               # 客户端配置
│   │   ├── retry.go                # 重试逻辑
│   │   └── proxy.go                # 代理支持
│   ├── middleware/                 # HTTP中间件
│   │   ├── cors.go                 # CORS中间件
│   │   ├── auth.go                 # 认证中间件
│   │   ├── ratelimit.go            # 限流中间件
│   │   ├── logging.go              # 日志中间件
│   │   ├── proxy.go                # 代理中间件
│   │   └── recovery.go             # 恢复中间件
│   └── response/                   # 响应处理
│       ├── response.go             # 响应结构
│       ├── error.go                # 错误响应
│       └── pagination.go           # 分页响应
├── auth/                           # 认证组件
│   ├── jwt/                        # JWT认证
│   │   ├── jwt.go                  # JWT实现
│   │   ├── config.go               # JWT配置
│   │   └── middleware.go           # JWT中间件
│   ├── bcrypt/                     # 密码加密
│   │   └── password.go             # 密码处理
│   └── interfaces.go               # 认证接口
├── crypto/                         # 加密工具
│   ├── aes.go                      # AES加密
│   ├── rsa.go                      # RSA加密
│   ├── hash.go                     # 哈希函数
│   └── random.go                   # 随机数生成
├── validation/                     # 数据验证
│   ├── validator.go                # 验证器
│   ├── rules.go                    # 验证规则
│   └── errors.go                   # 验证错误
├── utils/                          # 工具函数
│   ├── strings.go                  # 字符串工具
│   ├── time.go                     # 时间工具
│   ├── files.go                    # 文件工具
│   ├── network.go                  # 网络工具
│   ├── systemctl.go                # systemctl工具
│   └── conversion.go               # 类型转换
├── clash/                          # Clash API工具
│   ├── client.go                   # Clash API客户端
│   ├── types.go                    # Clash API类型
│   ├── proxy.go                    # 代理处理
│   └── config.go                   # 配置处理
└── testing/                        # 测试工具
    ├── mock/                       # Mock工具
    │   ├── database.go             # 数据库Mock
    │   ├── grpc.go                 # gRPCMock
    │   ├── http.go                 # HTTPMock
    │   └── clash.go                # ClashAPIKMock
    ├── fixtures/                   # 测试数据
    │   ├── nodes.go                # 节点测试数据
    │   ├── configs.go              # 配置测试数据
    │   ├── users.go                # 用户测试数据
    │   └── clash.go                # Clash测试数据
    └── testutil/                   # 测试工具
        ├── database.go             # 数据库测试工具
        ├── server.go               # 服务器测试工具
        ├── grpc.go                 # gRPC测试工具
        └── assert.go               # 断言工具
```

---

## 6. API 定义和生成代码 (api/)

```
api/
├── proto/                          # Protocol Buffers定义
│   ├── common/                     # 通用消息定义
│   │   └── v1/
│   │       ├── common.proto        # 通用类型
│   │       ├── config.proto        # 配置消息
│   │       ├── metrics.proto       # 监控指标
│   │       └── clash.proto         # Clash API类型
│   ├── web/                        # Web服务API
│   │   └── v1/
│   │       ├── service.proto       # Web服务接口
│   │       ├── auth.proto          # 认证相关
│   │       ├── user.proto          # 用户相关
│   │       └── proxy.proto         # 代理相关
│   ├── manager/                    # Manager服务API
│   │   └── v1/
│   │       ├── service.proto       # 服务定义
│   │       ├── node.proto          # 节点相关消息
│   │       ├── config.proto        # 配置相关消息
│   │       └── deployment.proto    # 部署相关消息
│   └── agent/                      # Agent服务API
│       └── v1/
│           ├── service.proto       # 服务定义
│           ├── types.proto         # 类型定义
│           ├── monitor.proto       # 监控相关
│           └── clash.proto         # Clash API相关
├── generated/                      # 自动生成的代码
│   ├── go/                         # Go生成代码
│   │   ├── common/v1/              # 通用类型生成代码
│   │   ├── web/v1/                 # Web服务生成代码
│   │   ├── manager/v1/             # Manager服务生成代码
│   │   └── agent/v1/               # Agent服务生成代码
│   └── openapi/                    # OpenAPI文档
│       ├── web-v1.yaml             # Web API文档
│       ├── manager-v1.yaml         # Manager API文档
│       └── agent-v1.yaml           # Agent API文档
├── client/                         # 客户端SDK
│   ├── go/                         # Go客户端
│   │   ├── web/                    # Web客户端
│   │   │   ├── client.go           # 客户端实现
│   │   │   └── config.go           # 客户端配置
│   │   ├── manager/                # Manager客户端
│   │   │   ├── client.go           # 客户端实现
│   │   │   ├── config.go           # 客户端配置
│   │   │   └── options.go          # 客户端选项
│   │   └── agent/                  # Agent客户端
│   │       ├── client.go           # 客户端实现
│   │       └── config.go           # 客户端配置
│   └── typescript/                 # TypeScript客户端
│       ├── src/                    # 源代码
│       ├── package.json            # 包定义
│       └── tsconfig.json           # TypeScript配置
└── docs/                           # API文档
    ├── web/                        # Web API文档
    │   ├── overview.md             # API概览
    │   ├── authentication.md       # 认证说明
    │   └── endpoints.md            # 接口说明
    ├── manager/                    # Manager API文档
    │   ├── overview.md             # API概览
    │   ├── authentication.md       # 认证说明
    │   ├── nodes.md                # 节点管理API
    │   ├── configs.md              # 配置管理API
    │   └── metrics.md              # 监控数据API
    ├── agent/                      # Agent API文档
    │   ├── overview.md             # API概览
    │   ├── protocol.md             # 协议说明
    │   └── clash.md                # Clash API说明
    └── clash/                      # Clash API集成文档
        ├── overview.md             # 集成概览
        ├── proxy.md                # 代理功能
        └── examples.md             # 使用示例
```

---

## 7. 配置文件结构 (configs/)

```
configs/
├── web/                            # Web服务器配置
│   ├── config.yaml                 # 主配置文件
│   ├── config-dev.yaml             # 开发环境配置
│   └── config-prod.yaml            # 生产环境配置
├── api/                            # API服务器配置
│   ├── config.yaml                 # 主配置文件
│   ├── config-dev.yaml             # 开发环境配置
│   ├── config-prod.yaml            # 生产环境配置
│   └── config-test.yaml            # 测试环境配置
├── agent/                          # Agent配置
│   ├── config.yaml                 # 主配置文件
│   ├── config-dev.yaml             # 开发环境配置
│   └── config-prod.yaml            # 生产环境配置
├── examples/                       # 配置示例
│   ├── web-complete.yaml           # 完整Web配置示例
│   ├── api-complete.yaml           # 完整API配置示例
│   ├── agent-complete.yaml         # 完整Agent配置示例
│   └── clash-integration.yaml      # Clash集成配置示例
└── schemas/                        # 配置模式定义
    ├── web-schema.json             # Web配置模式
    ├── api-schema.json             # API配置模式
    ├── agent-schema.json           # Agent配置模式
    └── clash-schema.json           # Clash配置模式
```

---

## 8. 开发规范和最佳实践

### 8.1 代码组织原则

1. **清晰的分层架构**
   ```
   前端界面 → sing-box-web → sing-box-api → sing-box-agent → sing-box
   ```

2. **依赖方向原则**
   ```
   cmd → internal → pkg
   web → api → agent
   ```

3. **接口隔离原则**
   ```go
   // internal/api/service/interfaces.go
   type NodeService interface {
       CreateNode(ctx context.Context, req *CreateNodeRequest) (*Node, error)
       GetNode(ctx context.Context, id string) (*Node, error)
       UpdateNode(ctx context.Context, id string, req *UpdateNodeRequest) (*Node, error)
       DeleteNode(ctx context.Context, id string) error
       ListNodes(ctx context.Context, filter *NodeFilter) (*NodeList, error)
   }
   
   type ClashService interface {
       ProxyClashAPI(ctx context.Context, nodeID string, req *ClashAPIRequest) (*ClashAPIResponse, error)
       GetProxyStatus(ctx context.Context, nodeID string) (*ProxyStatus, error)
       SwitchProxy(ctx context.Context, nodeID string, proxyName string) error
   }
   ```

### 8.2 Clash API集成规范

#### Agent端实现规范
```go
// internal/agent/clash/proxy.go
type ClashAPIProxy struct {
    baseURL    string
    client     *http.Client
    secret     string
    logger     *logrus.Logger
}

func (p *ClashAPIProxy) HandleRequest(req *ClashAPIRequest) (*ClashAPIResponse, error) {
    // 1. 请求验证
    if err := p.validateRequest(req); err != nil {
        return nil, err
    }
    
    // 2. 构建HTTP请求
    httpReq, err := p.buildHTTPRequest(req)
    if err != nil {
        return nil, err
    }
    
    // 3. 发送请求
    resp, err := p.client.Do(httpReq)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    // 4. 构建响应
    return p.buildResponse(req, resp)
}
```

#### 错误处理规范
```go
// internal/shared/errors/clash.go
var (
    ErrClashAPIUnavailable = errors.New("clash API is unavailable")
    ErrClashAPITimeout     = errors.New("clash API request timeout")
    ErrClashAPIAuth        = errors.New("clash API authentication failed")
    ErrClashAPINotFound    = errors.New("clash API endpoint not found")
)
```

### 8.3 测试规范

#### 单元测试示例
```go
// internal/agent/clash/proxy_test.go
func TestClashAPIProxy_HandleRequest(t *testing.T) {
    tests := []struct {
        name       string
        req        *ClashAPIRequest
        mockSetup  func(*httptest.Server)
        want       *ClashAPIResponse
        wantErr    bool
    }{
        {
            name: "get proxies success",
            req: &ClashAPIRequest{
                Method: "GET",
                Path:   "/proxies",
            },
            mockSetup: func(server *httptest.Server) {
                // Setup mock Clash API response
            },
            want: &ClashAPIResponse{
                StatusCode: 200,
                Body:       []byte(`{"GLOBAL":{"type":"Selector"}}`),
            },
            wantErr: false,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

---

## 9. 构建和部署配置更新

### 9.1 更新的Makefile

```makefile
# 构建目标
build: build-web build-api build-agent

build-web:
	@echo "Building sing-box-web..."
	@CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build \
		-ldflags "$(LDFLAGS)" \
		-o $(BUILD_DIR)/sing-box-web \
		./cmd/sing-box-web

build-api:
	@echo "Building sing-box-api..."
	@CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build \
		-ldflags "$(LDFLAGS)" \
		-o $(BUILD_DIR)/sing-box-api \
		./cmd/sing-box-api

build-agent:
	@echo "Building sing-box-agent..."
	@CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build \
		-ldflags "$(LDFLAGS)" \
		-o $(BUILD_DIR)/sing-box-agent \
		./cmd/sing-box-agent

# Docker构建
docker-build: docker-build-web docker-build-api docker-build-agent

docker-build-web:
	@echo "Building sing-box-web Docker image..."
	@docker build -f cmd/sing-box-web/Dockerfile -t sing-box-web:$(VERSION) .

docker-build-api:
	@echo "Building sing-box-api Docker image..."
	@docker build -f cmd/sing-box-api/Dockerfile -t sing-box-api:$(VERSION) .

docker-build-agent:
	@echo "Building sing-box-agent Docker image..."
	@docker build -f cmd/sing-box-agent/Dockerfile -t sing-box-agent:$(VERSION) .

# 测试目标
test-web:
	@echo "Testing sing-box-web..."
	@go test -race -coverprofile=coverage-web.out ./internal/web/... ./cmd/sing-box-web/...

test-api:
	@echo "Testing sing-box-api..."
	@go test -race -coverprofile=coverage-api.out ./internal/api/... ./cmd/sing-box-api/...

test-agent:
	@echo "Testing sing-box-agent..."
	@go test -race -coverprofile=coverage-agent.out ./internal/agent/... ./cmd/sing-box-agent/...

test-clash:
	@echo "Testing Clash API integration..."
	@go test -race -coverprofile=coverage-clash.out ./pkg/clash/... ./internal/agent/clash/...
```

---

## 10. 总结

这个更新的代码结构方案为sing-box-web项目提供了：

### 10.1 清晰的架构分层
- **sing-box-web**: 专注前端服务和用户交互
- **sing-box-api**: 专注业务逻辑和数据管理
- **sing-box-agent**: 专注节点控制和Clash API集成

### 10.2 完整的Clash API集成
- Agent端完整的API代理实现
- 支持所有Clash API功能（代理管理、状态监控、流量统计）
- 统一的错误处理和日志记录
- 高性能的HTTP代理和WebSocket支持

### 10.3 模块化设计
- 清晰的模块边界和接口定义
- 可插拔的组件设计（数据库、缓存、监控）
- 完整的测试覆盖和Mock支持
- 灵活的配置管理

### 10.4 生产就绪特性
- 完整的监控和告警系统
- 优雅的服务启停和错误恢复
- 全面的安全防护和认证授权
- 高可用和负载均衡支持

这个架构充分考虑了sing-box的分布式管理需求，通过Clash API集成提供了强大的节点控制能力，为构建稳定可靠的sing-box管理平台奠定了坚实基础。