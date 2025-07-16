# 后端代码结构方案 - sing-box-web

## 1. 项目整体目录结构

基于 Go 项目最佳实践和《Standard Go Project Layout》，结合 sing-box-web 项目的特点，我们采用以下目录结构：

```
sing-box-web/
├── cmd/                            # 应用程序入口点
│   ├── sing-box-api/               # API 服务器
│   ├── sing-box-web/               # Web 前端服务器
│   ├── sing-box-agent/             # 节点代理
│   └── sing-box-ctl/               # 管理工具
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

## 2. 核心应用代码结构

### 2.1 API 服务器 (cmd/sing-box-api)

```
cmd/sing-box-api/
├── main.go                         # 应用入口点
├── app/                            # 应用逻辑层
│   ├── options/                    # 命令行选项
│   │   ├── options.go              # 主要选项定义
│   │   ├── validation.go           # 选项验证
│   │   └── completion.go           # 命令行补全
│   ├── server/                     # 服务器逻辑
│   │   ├── server.go               # 服务器主逻辑
│   │   ├── grpc.go                 # gRPC 服务器
│   │   ├── http.go                 # HTTP 服务器
│   │   ├── metrics.go              # 监控指标服务器
│   │   └── websocket.go            # WebSocket 服务器
│   ├── config/                     # 配置管理
│   │   ├── config.go               # 配置结构定义
│   │   ├── loader.go               # 配置加载器
│   │   └── validator.go            # 配置验证器
│   └── version/                    # 版本信息
│       ├── version.go              # 版本定义
│       └── build_info.go           # 构建信息
└── Dockerfile                      # API 服务器容器构建
```

#### main.go 入口文件示例

```go
// cmd/sing-box-api/main.go
package main

import (
    "context"
    "flag"
    "fmt"
    "os"
    "os/signal"
    "syscall"

    "github.com/spf13/cobra"
    "github.com/your-org/sing-box-web/cmd/sing-box-api/app"
    "github.com/your-org/sing-box-web/cmd/sing-box-api/app/options"
    "github.com/your-org/sing-box-web/pkg/version"
)

func main() {
    cmd := newAPIServerCommand()
    if err := cmd.Execute(); err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
}

func newAPIServerCommand() *cobra.Command {
    opts := options.NewAPIServerOptions()

    cmd := &cobra.Command{
        Use:   "sing-box-api",
        Short: "sing-box web management API server",
        Long:  `API server for sing-box-web distributed management platform`,
        RunE: func(cmd *cobra.Command, args []string) error {
            if err := opts.Validate(); err != nil {
                return err
            }
            return runAPIServer(opts)
        },
    }

    // 添加子命令
    cmd.AddCommand(newServeCommand(opts))
    cmd.AddCommand(newVersionCommand())
    cmd.AddCommand(newConfigCommand())

    return cmd
}

func newServeCommand(opts *options.APIServerOptions) *cobra.Command {
    cmd := &cobra.Command{
        Use:   "serve",
        Short: "Start the API server",
        RunE: func(cmd *cobra.Command, args []string) error {
            return runAPIServer(opts)
        },
    }

    opts.AddFlags(cmd.Flags())
    return cmd
}

func runAPIServer(opts *options.APIServerOptions) error {
    ctx, cancel := signal.NotifyContext(context.Background(), 
        syscall.SIGINT, syscall.SIGTERM)
    defer cancel()

    server, err := app.NewAPIServer(opts)
    if err != nil {
        return fmt.Errorf("failed to create API server: %w", err)
    }

    return server.Run(ctx)
}
```

### 2.2 Web 服务器 (cmd/sing-box-web)

```
cmd/sing-box-web/
├── main.go                         # 应用入口点
├── app/                            # 应用逻辑层
│   ├── options/                    # 命令行选项
│   │   ├── options.go              # Web 服务器选项
│   │   └── validation.go           # 选项验证
│   ├── server/                     # 服务器逻辑
│   │   ├── server.go               # Web 服务器主逻辑
│   │   ├── handlers/               # HTTP 处理器
│   │   │   ├── spa.go              # 单页应用处理器
│   │   │   ├── api_proxy.go        # API 代理处理器
│   │   │   ├── health.go           # 健康检查处理器
│   │   │   └── assets.go           # 静态资源处理器
│   │   ├── middleware/             # 中间件
│   │   │   ├── cors.go             # CORS 中间件
│   │   │   ├── security.go         # 安全中间件
│   │   │   ├── logging.go          # 日志中间件
│   │   │   └── metrics.go          # 监控中间件
│   │   └── routes.go               # 路由定义
│   ├── config/                     # 配置管理
│   └── assets/                     # 静态资源管理
│       ├── embed.go                # 嵌入式资源
│       └── handler.go              # 资源处理器
└── Dockerfile                      # Web 服务器容器构建
```

### 2.3 Agent 应用 (cmd/sing-box-agent)

```
cmd/sing-box-agent/
├── main.go                         # 应用入口点
├── app/                            # 应用逻辑层
│   ├── options/                    # 命令行选项
│   │   ├── options.go              # Agent 选项
│   │   └── validation.go           # 选项验证
│   ├── agent/                      # Agent 主逻辑
│   │   ├── agent.go                # Agent 主体
│   │   ├── connection.go           # 连接管理
│   │   ├── heartbeat.go            # 心跳管理
│   │   └── registry.go             # 注册逻辑
│   ├── monitor/                    # 监控模块
│   │   ├── collector.go            # 监控数据收集器
│   │   ├── system.go               # 系统监控
│   │   ├── singbox.go              # sing-box 监控
│   │   └── reporter.go             # 监控数据上报
│   ├── config/                     # 配置管理
│   │   ├── manager.go              # 配置管理器
│   │   ├── watcher.go              # 配置文件监控
│   │   └── validator.go            # 配置验证器
│   └── singbox/                    # sing-box 控制
│       ├── controller.go           # sing-box 控制器
│       ├── process.go              # 进程管理
│       └── config.go               # 配置操作
└── Dockerfile                      # Agent 容器构建
```

---

## 3. 内部应用代码 (internal/)

```
internal/
├── api/                            # API 服务器内部代码
│   ├── server/                     # 服务器实现
│   │   ├── grpc/                   # gRPC 服务实现
│   │   │   ├── server.go           # gRPC 服务器
│   │   │   ├── manager.go          # Manager 服务实现
│   │   │   ├── interceptors/       # gRPC 拦截器
│   │   │   │   ├── auth.go         # 认证拦截器
│   │   │   │   ├── logging.go      # 日志拦截器
│   │   │   │   ├── metrics.go      # 监控拦截器
│   │   │   │   └── recovery.go     # 恢复拦截器
│   │   │   └── handlers/           # gRPC 处理器
│   │   │       ├── agent.go        # Agent 相关处理
│   │   │       ├── node.go         # 节点管理处理
│   │   │       └── config.go       # 配置管理处理
│   │   ├── http/                   # HTTP/REST API 实现
│   │   │   ├── server.go           # HTTP 服务器
│   │   │   ├── router.go           # 路由配置
│   │   │   ├── middleware/         # HTTP 中间件
│   │   │   │   ├── auth.go         # JWT 认证中间件
│   │   │   │   ├── cors.go         # CORS 中间件
│   │   │   │   ├── ratelimit.go    # 限流中间件
│   │   │   │   ├── requestid.go    # 请求ID中间件
│   │   │   │   └── recovery.go     # 恢复中间件
│   │   │   └── handlers/           # HTTP 处理器
│   │   │       ├── auth.go         # 认证处理器
│   │   │       ├── nodes.go        # 节点管理
│   │   │       ├── configs.go      # 配置管理
│   │   │       ├── deployments.go  # 部署管理
│   │   │       ├── metrics.go      # 监控数据
│   │   │       └── dashboard.go    # 仪表盘
│   │   └── websocket/              # WebSocket 服务
│   │       ├── hub.go              # 连接中心
│   │       ├── client.go           # 客户端连接
│   │       └── message.go          # 消息处理
│   ├── service/                    # 业务服务层
│   │   ├── interfaces.go           # 服务接口定义
│   │   ├── auth/                   # 认证服务
│   │   │   ├── service.go          # 认证服务实现
│   │   │   ├── jwt.go              # JWT 处理
│   │   │   └── password.go         # 密码处理
│   │   ├── node/                   # 节点管理服务
│   │   │   ├── service.go          # 节点服务实现
│   │   │   ├── manager.go          # 节点管理器
│   │   │   └── registry.go         # 节点注册
│   │   ├── config/                 # 配置管理服务
│   │   │   ├── service.go          # 配置服务实现
│   │   │   ├── template.go         # 模板管理
│   │   │   ├── deployment.go       # 部署管理
│   │   │   └── validator.go        # 配置验证
│   │   ├── monitoring/             # 监控服务
│   │   │   ├── service.go          # 监控服务实现
│   │   │   ├── collector.go        # 数据收集
│   │   │   ├── aggregator.go       # 数据聚合
│   │   │   └── alerting.go         # 告警处理
│   │   └── notification/           # 通知服务
│   │       ├── service.go          # 通知服务实现
│   │       ├── email.go            # 邮件通知
│   │       └── webhook.go          # Webhook 通知
│   ├── repository/                 # 数据访问层
│   │   ├── interfaces.go           # 仓库接口定义
│   │   ├── user/                   # 用户数据仓库
│   │   │   ├── repository.go       # 用户仓库实现
│   │   │   └── model.go            # 用户数据模型
│   │   ├── node/                   # 节点数据仓库
│   │   │   ├── repository.go       # 节点仓库实现
│   │   │   └── model.go            # 节点数据模型
│   │   ├── config/                 # 配置数据仓库
│   │   │   ├── repository.go       # 配置仓库实现
│   │   │   └── model.go            # 配置数据模型
│   │   ├── metrics/                # 监控数据仓库
│   │   │   ├── repository.go       # 监控仓库实现
│   │   │   └── model.go            # 监控数据模型
│   │   └── migration/              # 数据库迁移
│   │       ├── migrator.go         # 迁移器
│   │       └── versions/           # 迁移版本
│   │           ├── 001_initial.go  # 初始化迁移
│   │           └── 002_add_labels.go # 添加标签迁移
│   └── cache/                      # 缓存层
│       ├── interfaces.go           # 缓存接口定义
│       ├── redis/                  # Redis 缓存实现
│       │   ├── client.go           # Redis 客户端
│       │   ├── session.go          # 会话缓存
│       │   ├── metrics.go          # 监控数据缓存
│       │   └── lock.go             # 分布式锁
│       └── memory/                 # 内存缓存实现
│           ├── cache.go            # 内存缓存
│           └── lru.go              # LRU 缓存
├── agent/                          # Agent 内部代码
│   ├── client/                     # gRPC 客户端
│   │   ├── client.go               # 客户端实现
│   │   ├── connection.go           # 连接管理
│   │   └── retry.go                # 重试逻辑
│   ├── monitor/                    # 系统监控
│   │   ├── system/                 # 系统监控实现
│   │   │   ├── cpu.go              # CPU 监控
│   │   │   ├── memory.go           # 内存监控
│   │   │   ├── disk.go             # 磁盘监控
│   │   │   └── network.go          # 网络监控
│   │   ├── singbox/                # sing-box 监控
│   │   │   ├── status.go           # 状态监控
│   │   │   ├── metrics.go          # 指标采集
│   │   │   └── logs.go             # 日志监控
│   │   └── collector.go            # 监控数据收集器
│   ├── config/                     # 配置管理
│   │   ├── manager.go              # 配置管理器
│   │   ├── applier.go              # 配置应用器
│   │   └── validator.go            # 配置验证器
│   └── singbox/                    # sing-box 控制
│       ├── controller.go           # sing-box 控制器
│       ├── process.go              # 进程管理
│       ├── config.go               # 配置操作
│       └── health.go               # 健康检查
└── shared/                         # 内部共享代码
    ├── errors/                     # 错误定义
    │   ├── codes.go                # 错误代码
    │   ├── errors.go               # 错误类型
    │   └── handler.go              # 错误处理器
    ├── constants/                  # 常量定义
    │   ├── status.go               # 状态常量
    │   ├── events.go               # 事件常量
    │   └── metrics.go              # 监控常量
    ├── utils/                      # 工具函数
    │   ├── crypto.go               # 加密工具
    │   ├── json.go                 # JSON 工具
    │   ├── time.go                 # 时间工具
    │   └── validation.go           # 验证工具
    └── types/                      # 共享类型
        ├── pagination.go           # 分页类型
        ├── filter.go               # 过滤器类型
        └── response.go             # 响应类型
```

---

## 4. 公共库代码 (pkg/)

```
pkg/
├── version/                        # 版本信息
│   ├── version.go                  # 版本定义
│   └── build.go                    # 构建信息
├── logger/                         # 统一日志
│   ├── logger.go                   # 日志接口
│   ├── logrus.go                   # Logrus 实现
│   ├── config.go                   # 日志配置
│   └── formatter.go                # 日志格式化器
├── config/                         # 配置管理
│   ├── config.go                   # 配置接口
│   ├── loader.go                   # 配置加载器
│   ├── watcher.go                  # 配置监控器
│   └── validator.go                # 配置验证器
├── database/                       # 数据库
│   ├── postgres/                   # PostgreSQL 驱动
│   │   ├── client.go               # 客户端
│   │   ├── config.go               # 配置
│   │   └── migration.go            # 迁移工具
│   ├── sqlite/                     # SQLite 驱动
│   │   ├── client.go               # 客户端
│   │   └── config.go               # 配置
│   └── interfaces.go               # 数据库接口
├── cache/                          # 缓存
│   ├── redis/                      # Redis 客户端
│   │   ├── client.go               # 客户端实现
│   │   ├── config.go               # 配置
│   │   └── cluster.go              # 集群支持
│   ├── memory/                     # 内存缓存
│   │   └── cache.go                # 内存缓存实现
│   └── interfaces.go               # 缓存接口
├── metrics/                        # 监控指标
│   ├── prometheus/                 # Prometheus 指标
│   │   ├── metrics.go              # 指标定义
│   │   ├── collector.go            # 指标收集器
│   │   └── middleware.go           # 中间件
│   ├── system/                     # 系统指标收集
│   │   ├── cpu.go                  # CPU 指标
│   │   ├── memory.go               # 内存指标
│   │   ├── disk.go                 # 磁盘指标
│   │   └── network.go              # 网络指标
│   └── interfaces.go               # 监控接口
├── grpc/                           # gRPC 通用组件
│   ├── server/                     # gRPC 服务器
│   │   ├── server.go               # 服务器实现
│   │   ├── config.go               # 服务器配置
│   │   └── interceptors.go         # 拦截器
│   ├── client/                     # gRPC 客户端
│   │   ├── client.go               # 客户端实现
│   │   ├── config.go               # 客户端配置
│   │   ├── pool.go                 # 连接池
│   │   └── retry.go                # 重试逻辑
│   └── middleware/                 # 中间件
│       ├── auth.go                 # 认证中间件
│       ├── logging.go              # 日志中间件
│       ├── metrics.go              # 监控中间件
│       └── tracing.go              # 链路追踪中间件
├── http/                           # HTTP 通用组件
│   ├── server/                     # HTTP 服务器
│   │   ├── server.go               # 服务器实现
│   │   ├── config.go               # 服务器配置
│   │   └── graceful.go             # 优雅关闭
│   ├── client/                     # HTTP 客户端
│   │   ├── client.go               # 客户端实现
│   │   ├── config.go               # 客户端配置
│   │   └── retry.go                # 重试逻辑
│   ├── middleware/                 # HTTP 中间件
│   │   ├── cors.go                 # CORS 中间件
│   │   ├── auth.go                 # 认证中间件
│   │   ├── ratelimit.go            # 限流中间件
│   │   ├── logging.go              # 日志中间件
│   │   └── recovery.go             # 恢复中间件
│   └── response/                   # 响应处理
│       ├── response.go             # 响应结构
│       ├── error.go                # 错误响应
│       └── pagination.go           # 分页响应
├── auth/                           # 认证组件
│   ├── jwt/                        # JWT 认证
│   │   ├── jwt.go                  # JWT 实现
│   │   ├── config.go               # JWT 配置
│   │   └── middleware.go           # JWT 中间件
│   ├── bcrypt/                     # 密码加密
│   │   └── password.go             # 密码处理
│   └── interfaces.go               # 认证接口
├── crypto/                         # 加密工具
│   ├── aes.go                      # AES 加密
│   ├── rsa.go                      # RSA 加密
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
│   └── conversion.go               # 类型转换
└── testing/                        # 测试工具
    ├── mock/                       # Mock 工具
    │   ├── database.go             # 数据库 Mock
    │   ├── grpc.go                 # gRPC Mock
    │   └── http.go                 # HTTP Mock
    ├── fixtures/                   # 测试数据
    │   ├── nodes.go                # 节点测试数据
    │   ├── configs.go              # 配置测试数据
    │   └── users.go                # 用户测试数据
    └── testutil/                   # 测试工具
        ├── database.go             # 数据库测试工具
        ├── server.go               # 服务器测试工具
        └── assert.go               # 断言工具
```

---

## 5. API 定义和生成代码 (api/)

```
api/
├── proto/                          # Protocol Buffers 定义
│   ├── common/                     # 通用消息定义
│   │   └── v1/
│   │       ├── common.proto        # 通用类型
│   │       ├── config.proto        # 配置消息
│   │       └── metrics.proto       # 监控指标
│   ├── manager/                    # Manager 服务 API
│   │   └── v1/
│   │       ├── service.proto       # 服务定义
│   │       ├── node.proto          # 节点相关消息
│   │       ├── config.proto        # 配置相关消息
│   │       └── metrics.proto       # 监控相关消息
│   └── agent/                      # Agent 服务 API
│       └── v1/
│           ├── service.proto       # 服务定义
│           └── types.proto         # 类型定义
├── generated/                      # 自动生成的代码
│   ├── go/                         # Go 生成代码
│   │   ├── common/v1/              # 通用类型生成代码
│   │   ├── manager/v1/             # Manager 服务生成代码
│   │   └── agent/v1/               # Agent 服务生成代码
│   └── openapi/                    # OpenAPI 文档
│       ├── manager-v1.yaml         # Manager API 文档
│       └── agent-v1.yaml           # Agent API 文档
├── client/                         # 客户端 SDK
│   ├── go/                         # Go 客户端
│   │   ├── manager/                # Manager 客户端
│   │   │   ├── client.go           # 客户端实现
│   │   │   ├── config.go           # 客户端配置
│   │   │   └── options.go          # 客户端选项
│   │   └── agent/                  # Agent 客户端
│   │       ├── client.go           # 客户端实现
│   │       └── config.go           # 客户端配置
│   └── typescript/                 # TypeScript 客户端
│       ├── src/                    # 源代码
│       ├── package.json            # 包定义
│       └── tsconfig.json           # TypeScript 配置
└── docs/                           # API 文档
    ├── manager/                    # Manager API 文档
    │   ├── overview.md             # API 概览
    │   ├── authentication.md       # 认证说明
    │   ├── nodes.md                # 节点管理 API
    │   ├── configs.md              # 配置管理 API
    │   └── metrics.md              # 监控数据 API
    └── agent/                      # Agent API 文档
        ├── overview.md             # API 概览
        └── protocol.md             # 协议说明
```

---

## 6. 配置文件结构 (configs/)

```
configs/
├── api/                            # API 服务器配置
│   ├── config.yaml                 # 主配置文件
│   ├── config-dev.yaml             # 开发环境配置
│   ├── config-prod.yaml            # 生产环境配置
│   └── config-test.yaml            # 测试环境配置
├── web/                            # Web 服务器配置
│   ├── config.yaml                 # 主配置文件
│   ├── config-dev.yaml             # 开发环境配置
│   └── config-prod.yaml            # 生产环境配置
├── agent/                          # Agent 配置
│   ├── config.yaml                 # 主配置文件
│   ├── config-dev.yaml             # 开发环境配置
│   └── config-prod.yaml            # 生产环境配置
├── examples/                       # 配置示例
│   ├── api-complete.yaml           # 完整 API 配置示例
│   ├── web-complete.yaml           # 完整 Web 配置示例
│   └── agent-complete.yaml         # 完整 Agent 配置示例
└── schemas/                        # 配置模式定义
    ├── api-schema.json             # API 配置模式
    ├── web-schema.json             # Web 配置模式
    └── agent-schema.json           # Agent 配置模式
```

---

## 7. 开发规范和最佳实践

### 7.1 代码组织原则

1. **依赖方向**: 外层依赖内层，内层不依赖外层
   ```
   cmd → internal → pkg
   api → internal → pkg
   ```

2. **接口隔离**: 每个层次定义清晰的接口，降低耦合度
   ```go
   // internal/api/service/interfaces.go
   type NodeService interface {
       CreateNode(ctx context.Context, req *CreateNodeRequest) (*Node, error)
       GetNode(ctx context.Context, id string) (*Node, error)
       UpdateNode(ctx context.Context, id string, req *UpdateNodeRequest) (*Node, error)
       DeleteNode(ctx context.Context, id string) error
       ListNodes(ctx context.Context, filter *NodeFilter) (*NodeList, error)
   }
   ```

3. **错误处理**: 统一错误处理机制
   ```go
   // internal/shared/errors/errors.go
   var (
       ErrNodeNotFound = errors.New("node not found")
       ErrInvalidConfig = errors.New("invalid configuration")
       ErrAgentOffline = errors.New("agent is offline")
   )
   ```

### 7.2 命名规范

#### 包命名
- 使用小写字母
- 包名简短且具有描述性
- 避免使用 `util`、`common` 等泛化名称

#### 文件命名
- 使用小写字母和下划线
- 文件名反映其主要功能
- 测试文件以 `_test.go` 结尾

#### 变量和函数命名
- 使用驼峰命名法
- 公开成员首字母大写
- 私有成员首字母小写
- 接口名以 `er` 结尾（如 `Reader`、`Writer`）

### 7.3 注释规范

```go
// Package node provides node management functionality.
//
// This package includes services for node registration, monitoring,
// and configuration management.
package node

// NodeService defines the interface for node management operations.
//
// All methods should be implemented in a thread-safe manner and
// should handle context cancellation appropriately.
type NodeService interface {
    // CreateNode creates a new node with the given configuration.
    //
    // Parameters:
    //   - ctx: The context for the operation
    //   - req: The node creation request
    //
    // Returns:
    //   - *Node: The created node
    //   - error: Any error that occurred during creation
    CreateNode(ctx context.Context, req *CreateNodeRequest) (*Node, error)
}
```

### 7.4 测试规范

```go
// internal/api/service/node/service_test.go
package node

import (
    "context"
    "testing"
    "time"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    "github.com/your-org/sing-box-web/internal/api/repository/node"
    "github.com/your-org/sing-box-web/pkg/testing/fixtures"
)

func TestNodeService_CreateNode(t *testing.T) {
    tests := []struct {
        name    string
        req     *CreateNodeRequest
        setup   func(*node.MockRepository)
        want    *Node
        wantErr bool
    }{
        {
            name: "success",
            req:  fixtures.NewCreateNodeRequest(),
            setup: func(repo *node.MockRepository) {
                repo.On("Create", mock.Anything, mock.Anything).
                    Return(fixtures.NewNode(), nil)
            },
            want:    fixtures.NewNode(),
            wantErr: false,
        },
        {
            name: "duplicate name",
            req:  fixtures.NewCreateNodeRequest(),
            setup: func(repo *node.MockRepository) {
                repo.On("Create", mock.Anything, mock.Anything).
                    Return(nil, node.ErrDuplicateName)
            },
            want:    nil,
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            repo := &node.MockRepository{}
            tt.setup(repo)

            service := NewNodeService(repo)
            got, err := service.CreateNode(context.Background(), tt.req)

            if tt.wantErr {
                assert.Error(t, err)
                assert.Nil(t, got)
            } else {
                assert.NoError(t, err)
                assert.Equal(t, tt.want, got)
            }

            repo.AssertExpectations(t)
        })
    }
}
```

---

## 8. 构建和部署配置

### 8.1 Makefile

```makefile
# Makefile
.PHONY: all build test clean generate lint install-tools docker-build

# 构建变量
VERSION ?= $(shell git describe --tags --dirty --always)
BUILD_TIME ?= $(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT ?= $(shell git rev-parse HEAD)
GIT_BRANCH ?= $(shell git rev-parse --abbrev-ref HEAD)

# Go 变量
GO_VERSION := 1.21
GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)

# 输出目录
BUILD_DIR := build
DIST_DIR := dist

# 编译标志
LDFLAGS := -X 'github.com/your-org/sing-box-web/pkg/version.Version=$(VERSION)' \
          -X 'github.com/your-org/sing-box-web/pkg/version.BuildTime=$(BUILD_TIME)' \
          -X 'github.com/your-org/sing-box-web/pkg/version.GitCommit=$(GIT_COMMIT)' \
          -X 'github.com/your-org/sing-box-web/pkg/version.GitBranch=$(GIT_BRANCH)'

# 默认目标
all: generate build

# 生成代码
generate:
	@echo "Generating code..."
	@go generate ./...
	@$(MAKE) generate-proto
	@$(MAKE) generate-mocks

generate-proto:
	@echo "Generating protobuf code..."
	@buf generate

generate-mocks:
	@echo "Generating mocks..."
	@mockgen -source=internal/api/service/interfaces.go -destination=internal/api/service/mocks/service.go
	@mockgen -source=internal/api/repository/interfaces.go -destination=internal/api/repository/mocks/repository.go

# 构建应用
build: build-api build-web build-agent build-ctl

build-api:
	@echo "Building sing-box-api..."
	@CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build \
		-ldflags "$(LDFLAGS)" \
		-o $(BUILD_DIR)/sing-box-api \
		./cmd/sing-box-api

build-web:
	@echo "Building sing-box-web..."
	@CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build \
		-ldflags "$(LDFLAGS)" \
		-o $(BUILD_DIR)/sing-box-web \
		./cmd/sing-box-web

build-agent:
	@echo "Building sing-box-agent..."
	@CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build \
		-ldflags "$(LDFLAGS)" \
		-o $(BUILD_DIR)/sing-box-agent \
		./cmd/sing-box-agent

build-ctl:
	@echo "Building sing-box-ctl..."
	@CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build \
		-ldflags "$(LDFLAGS)" \
		-o $(BUILD_DIR)/sing-box-ctl \
		./cmd/sing-box-ctl

# 运行测试
test:
	@echo "Running tests..."
	@go test -race -coverprofile=coverage.out ./...

test-integration:
	@echo "Running integration tests..."
	@go test -tags=integration -race ./test/...

# 代码检查
lint:
	@echo "Running linters..."
	@golangci-lint run

fmt:
	@echo "Formatting code..."
	@go fmt ./...
	@goimports -w .

# 清理
clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@rm -rf $(DIST_DIR)
	@rm -f coverage.out

# 安装开发工具
install-tools:
	@echo "Installing development tools..."
	@go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	@go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	@go install github.com/golang/mock/mockgen@latest
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install golang.org/x/tools/cmd/goimports@latest
	@go install github.com/bufbuild/buf/cmd/buf@latest

# Docker 构建
docker-build: docker-build-api docker-build-web docker-build-agent

docker-build-api:
	@echo "Building sing-box-api Docker image..."
	@docker build -f deployments/docker/api.Dockerfile -t sing-box-api:$(VERSION) .

docker-build-web:
	@echo "Building sing-box-web Docker image..."
	@docker build -f deployments/docker/web.Dockerfile -t sing-box-web:$(VERSION) .

docker-build-agent:
	@echo "Building sing-box-agent Docker image..."
	@docker build -f deployments/docker/agent.Dockerfile -t sing-box-agent:$(VERSION) .

# 发布包
release: clean
	@echo "Building release packages..."
	@mkdir -p $(DIST_DIR)
	@for os in linux darwin windows; do \
		for arch in amd64 arm64; do \
			echo "Building $$os/$$arch..."; \
			GOOS=$$os GOARCH=$$arch $(MAKE) build; \
			if [ "$$os" = "windows" ]; then \
				mv $(BUILD_DIR)/sing-box-api $(BUILD_DIR)/sing-box-api.exe; \
				mv $(BUILD_DIR)/sing-box-web $(BUILD_DIR)/sing-box-web.exe; \
				mv $(BUILD_DIR)/sing-box-agent $(BUILD_DIR)/sing-box-agent.exe; \
				mv $(BUILD_DIR)/sing-box-ctl $(BUILD_DIR)/sing-box-ctl.exe; \
			fi; \
			tar -czf $(DIST_DIR)/sing-box-web-$(VERSION)-$$os-$$arch.tar.gz -C $(BUILD_DIR) .; \
			$(MAKE) clean; \
		done; \
	done

# 运行开发环境
dev-up:
	@echo "Starting development environment..."
	@docker-compose -f deployments/docker-compose.dev.yml up -d

dev-down:
	@echo "Stopping development environment..."
	@docker-compose -f deployments/docker-compose.dev.yml down

# 数据库迁移
migrate-up:
	@echo "Running database migrations..."
	@./$(BUILD_DIR)/sing-box-api migrate up

migrate-down:
	@echo "Rolling back database migrations..."
	@./$(BUILD_DIR)/sing-box-api migrate down

# 验证生成的代码是否最新
verify-generate: generate
	@echo "Verifying generated code is up-to-date..."
	@git diff --exit-code api/generated/ || (echo "Generated code is out of date. Please run 'make generate'" && exit 1)

# CI 目标
ci: install-tools generate verify-generate lint test

# 帮助信息
help:
	@echo "Available targets:"
	@echo "  all              - Generate code and build all applications"
	@echo "  generate         - Generate all code (protobuf, mocks, etc.)"
	@echo "  build            - Build all applications"
	@echo "  test             - Run all tests"
	@echo "  lint             - Run code linters"
	@echo "  clean            - Clean build artifacts"
	@echo "  install-tools    - Install development tools"
	@echo "  docker-build     - Build Docker images"
	@echo "  release          - Build release packages"
	@echo "  dev-up           - Start development environment"
	@echo "  dev-down         - Stop development environment"
	@echo "  migrate-up       - Run database migrations"
	@echo "  migrate-down     - Rollback database migrations"
	@echo "  ci               - Run CI checks"
	@echo "  help             - Show this help message"
```

### 8.2 Go Module 配置

```go
// go.mod
module github.com/your-org/sing-box-web

go 1.21

require (
    github.com/gin-gonic/gin v1.9.1
    github.com/spf13/cobra v1.7.0
    github.com/spf13/viper v1.16.0
    google.golang.org/grpc v1.58.0
    google.golang.org/protobuf v1.31.0
    gorm.io/gorm v1.25.4
    gorm.io/driver/postgres v1.5.2
    gorm.io/driver/sqlite v1.5.3
    github.com/redis/go-redis/v9 v9.1.0
    github.com/golang-jwt/jwt/v5 v5.0.0
    github.com/prometheus/client_golang v1.16.0
    github.com/sirupsen/logrus v1.9.3
    github.com/stretchr/testify v1.8.4
    golang.org/x/crypto v0.13.0
)

require (
    // 间接依赖...
)
```

---

## 9. 总结

这个代码结构方案为 sing-box-web 项目提供了：

1. **清晰的层次结构**: 遵循 Go 项目最佳实践，代码组织清晰
2. **良好的可维护性**: 模块化设计，便于功能扩展和维护
3. **强大的可测试性**: 接口驱动设计，便于单元测试和集成测试
4. **高效的开发体验**: 完善的工具链和自动化构建流程
5. **标准化的开发规范**: 统一的命名、注释和测试规范

该结构既满足了当前的功能需求，又为未来的扩展和优化提供了坚实的基础。 