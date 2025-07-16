# 技术选型建议 - sing-box-web

## 1. 系统架构概览

### 1.1 整体架构图

```
┌─────────────────────────────────────────────────────────────────┐
│                         前端用户界面                              │
│                        (React/Vue.js)                           │
└─────────────────────────────┬───────────────────────────────────┘
                              │ HTTP/WebSocket
┌─────────────────────────────▼───────────────────────────────────┐
│                      sing-box-web                               │
│                    (Go + Gin Framework)                         │
│                                                                 │
│ • 静态资源服务    • 用户认证     • 会话管理                        │
│ • API代理转发     • 操作审计     • WebSocket通信                  │
└─────────────────────────────┬───────────────────────────────────┘
                              │ gRPC over TLS
┌─────────────────────────────▼───────────────────────────────────┐
│                      sing-box-api                               │
│                   (Go + Gin + gRPC)                            │
│                                                                 │
│ • 节点管理        • 配置模板     • 部署编排                        │
│ • 监控数据        • 业务逻辑     • 数据存储                        │
└─────────────────────────────┬───────────────────────────────────┘
                              │ gRPC Client Pool
┌─────────────────────────────▼───────────────────────────────────┐
│                    sing-box-agent                               │
│                     (Go + gRPC)                                │
│                                                                 │
│ • sing-box控制    • 系统监控     • Clash API代理                  │
│ • 配置应用        • 进程管理     • 实时状态上报                     │
└─────────────────────────────┬───────────────────────────────────┘
                              │ Process Control
┌─────────────────────────────▼───────────────────────────────────┐
│                        sing-box                                │
│                   (C++ Core Binary)                            │
│                                                                 │
│ • 代理服务        • Clash API   • 配置热重载                      │
│ • 流量处理        • 状态监控     • 日志输出                        │
└─────────────────────────────────────────────────────────────────┘
```

### 1.2 技术栈分层

| 层级 | 组件 | 主要技术 | 职责 |
|------|------|----------|------|
| **前端层** | Web UI | React 18 + TypeScript + Vite | 用户界面、状态管理 |
| **网关层** | sing-box-web | Go 1.21 + Gin + WebSocket | 前端服务、认证代理 |
| **业务层** | sing-box-api | Go 1.21 + Gin + gRPC + GORM | 核心业务、数据管理 |
| **代理层** | sing-box-agent | Go 1.21 + gRPC Client | 节点控制、监控收集 |
| **存储层** | Database | PostgreSQL + Redis + TimescaleDB | 数据持久化、缓存 |
| **监控层** | Monitoring | Prometheus + Grafana + SkyWalking | 可观测性、告警 |

---

## 2. 核心技术选型

### 2.1 编程语言选择

#### Go 1.21+ (主要开发语言)

**选择理由**：
- **高性能**：原生并发支持，适合高并发场景
- **跨平台**：支持多操作系统和架构编译
- **生态丰富**：完善的Web框架、gRPC、数据库工具
- **部署简单**：单二进制文件，无运行时依赖
- **团队友好**：语法简洁，学习成本低
- **内存安全**：垃圾回收，避免内存泄漏

**版本要求**：Go 1.21+
- 支持泛型和最新性能优化
- 改进的garbage collector
- 更好的编译时间和运行性能

---

## 3. 应用框架选型

### 3.1 sing-box-web (前端服务器)

#### Web框架：Gin 1.9+

```go
// 选择理由
• 轻量高性能    - 路由性能优秀，中间件丰富
• 开发效率高    - 简洁的API设计，快速开发
• 社区活跃      - 大量中间件和插件支持
• 部署简单      - 编译为单一可执行文件
```

**核心依赖包**：
```go
// Web框架核心
github.com/gin-gonic/gin v1.9.1

// 认证相关
github.com/golang-jwt/jwt/v5 v5.0.0
github.com/gin-contrib/sessions v0.0.5

// 中间件
github.com/gin-contrib/cors v1.4.0
github.com/gin-contrib/secure v0.0.1
github.com/gin-contrib/gzip v0.0.6

// WebSocket支持
github.com/gorilla/websocket v1.5.0

// HTTP客户端
github.com/go-resty/resty/v2 v2.7.0
```

#### 架构特点
```
┌─────────────────┐
│   HTTP Router   │ ← Gin路由层
├─────────────────┤
│   Middleware    │ ← 认证、日志、CORS
├─────────────────┤
│   Handlers      │ ← 业务处理器
├─────────────────┤
│   gRPC Client   │ ← API服务调用
├─────────────────┤
│   WebSocket     │ ← 实时通信
└─────────────────┘
```

### 3.2 sing-box-api (核心API服务)

#### gRPC框架：grpc-go 1.59+

```go
// 选择理由
• 高性能通信    - Protocol Buffers序列化，HTTP/2传输
• 强类型接口    - 接口定义明确，支持多语言
• 流式处理      - 支持双向流，适合实时数据
• 负载均衡      - 内置负载均衡和服务发现
```

**核心依赖包**：
```go
// gRPC核心
google.golang.org/grpc v1.59.0
google.golang.org/protobuf v1.31.0

// HTTP服务（管理接口）
github.com/gin-gonic/gin v1.9.1

// 数据库ORM
gorm.io/gorm v1.25.4
gorm.io/driver/postgres v1.5.2

// 缓存
github.com/redis/go-redis/v9 v9.2.1

// 配置管理
github.com/spf13/viper v1.17.0

// 监控指标
github.com/prometheus/client_golang v1.17.0


```

#### 服务架构
```
┌─────────────────┐
│   gRPC Server   │ ← 对外服务接口
├─────────────────┤
│  HTTP Server    │ ← 管理和监控接口
├─────────────────┤
│ Business Logic  │ ← 业务逻辑层
├─────────────────┤
│   Repository    │ ← 数据访问层
├─────────────────┤
│   Database      │ ← PostgreSQL + Redis
└─────────────────┘
```

### 3.3 sing-box-agent (节点代理)

#### 系统控制：原生Go + 系统调用

```go
// 选择理由
• 系统集成     - 直接调用systemctl、进程管理
• 轻量部署     - 最小资源占用，单文件部署
• 可靠性高     - 断线重连，状态恢复
• 监控能力     - 实时系统指标收集
```

**核心依赖包**：
```go
// gRPC客户端
google.golang.org/grpc v1.59.0

// 系统监控
github.com/shirou/gopsutil/v3 v3.23.9

// 文件监控
github.com/fsnotify/fsnotify v1.7.0

// HTTP客户端（Clash API）
github.com/go-resty/resty/v2 v2.7.0

// 进程管理
golang.org/x/sys v0.13.0

// 配置管理
github.com/spf13/viper v1.17.0
```

#### Clash API集成架构
```
┌─────────────────┐
│  gRPC Client    │ ← 与API服务通信
├─────────────────┤
│ Clash API Proxy │ ← HTTP代理转发
├─────────────────┤
│ Process Manager │ ← sing-box进程控制
├─────────────────┤
│ System Monitor  │ ← 系统指标收集
├─────────────────┤
│ Config Manager  │ ← 配置文件管理
└─────────────────┘
```

---

## 4. 数据存储技术

### 4.1 主数据库：PostgreSQL 15+

**选择理由**：
- **ACID特性**：完整的事务支持，数据一致性保证
- **JSON支持**：原生JSONB类型，适合动态配置存储
- **扩展能力**：丰富的扩展插件（如TimescaleDB）
- **性能优秀**：查询优化器强大，并发性能好
- **运维成熟**：完善的备份、监控、调优工具

**配置建议**：
```yaml
postgres:
  version: "15.4"
  settings:
    max_connections: 200
    shared_buffers: 256MB
    effective_cache_size: 1GB
    work_mem: 4MB
    maintenance_work_mem: 64MB
    checkpoint_completion_target: 0.9
    wal_buffers: 16MB
    default_statistics_target: 100
```

### 4.2 缓存数据库：Redis 7+

**选择理由**：
- **高性能**：内存存储，微秒级响应时间
- **数据结构丰富**：String、Hash、List、Set、ZSet
- **分布式锁**：支持分布式场景下的并发控制
- **持久化选项**：RDB快照 + AOF日志
- **集群支持**：Redis Cluster模式

**应用场景**：
```yaml
redis_usage:
  session_storage:    # 用户会话存储
    key_pattern: "session:{token}"
    ttl: 3600s
  
  api_cache:          # API响应缓存
    key_pattern: "api:{method}:{path}:{params_hash}"
    ttl: 300s
  
  node_status:        # 节点状态缓存
    key_pattern: "node:status:{node_id}"
    ttl: 60s
  
  distributed_lock:   # 分布式锁
    key_pattern: "lock:{resource}"
    ttl: 30s
```

### 4.3 时序数据库：TimescaleDB

**选择理由**：
- **PostgreSQL兼容**：完全兼容PostgreSQL生态
- **自动分区**：按时间自动分区，查询性能优秀
- **数据压缩**：自动压缩历史数据，节省存储
- **保留策略**：自动清理过期数据
- **连续聚合**：预计算聚合视图，提升查询速度

**监控数据存储**：
```sql
-- 节点监控指标
CREATE TABLE node_metrics (
    time TIMESTAMPTZ NOT NULL,
    node_id UUID NOT NULL,
    metric_type VARCHAR(50) NOT NULL,
    value DOUBLE PRECISION,
    labels JSONB
);

-- 转换为超表
SELECT create_hypertable('node_metrics', 'time');

-- 数据保留策略
SELECT add_retention_policy('node_metrics', INTERVAL '90 days');
```

---

## 5. 通信协议与API

### 5.1 gRPC通信协议

#### Protocol Buffers定义

```protobuf
// api/proto/manager/v1/service.proto
syntax = "proto3";

package manager.v1;

import "google/protobuf/timestamp.proto";
import "google/protobuf/empty.proto";

// 节点管理服务
service NodeService {
  // 节点注册
  rpc RegisterNode(RegisterNodeRequest) returns (RegisterNodeResponse);
  
  // 节点心跳
  rpc Heartbeat(HeartbeatRequest) returns (HeartbeatResponse);
  
  // 配置分发
  rpc DeployConfig(DeployConfigRequest) returns (stream DeployConfigResponse);
  
  // 监控数据上报
  rpc ReportMetrics(stream MetricsReport) returns (google.protobuf.Empty);
}

// Clash API代理服务
service ClashService {
  // 代理Clash API请求
  rpc ProxyClashAPI(ClashAPIRequest) returns (ClashAPIResponse);
  
  // 获取代理状态
  rpc GetProxyStatus(GetProxyStatusRequest) returns (ProxyStatus);
  
  // 切换代理
  rpc SwitchProxy(SwitchProxyRequest) returns (SwitchProxyResponse);
}
```

#### gRPC连接配置

```yaml
grpc:
  server:
    addr: ":9090"
    network: tcp
    timeout: 30s
    keepalive:
      max_connection_idle: 15s
      max_connection_age: 30s
      max_connection_age_grace: 5s
      time: 5s
      timeout: 1s
  
  client:
    timeout: 10s
    keepalive:
      time: 30s
      timeout: 5s
      permit_without_stream: true
    retry:
      max_attempts: 3
      initial_backoff: 1s
      max_backoff: 30s
```

### 5.2 RESTful API设计

#### API版本管理
```
/api/v1/auth/login          # 用户登录
/api/v1/auth/logout         # 用户登出
/api/v1/auth/refresh        # 刷新Token

/api/v1/users               # 用户管理
/api/v1/nodes               # 节点管理（代理到API服务）
/api/v1/configs             # 配置管理（代理到API服务）
/api/v1/deployments         # 部署管理（代理到API服务）

/api/v1/ws/events           # WebSocket事件推送
```

#### HTTP状态码规范
```yaml
http_status:
  success:
    200: "OK - 请求成功"
    201: "Created - 资源创建成功"
    204: "No Content - 删除成功"
  
  client_error:
    400: "Bad Request - 请求参数错误"
    401: "Unauthorized - 未认证"
    403: "Forbidden - 权限不足"
    404: "Not Found - 资源不存在"
    409: "Conflict - 资源冲突"
    422: "Unprocessable Entity - 验证失败"
  
  server_error:
    500: "Internal Server Error - 服务器内部错误"
    502: "Bad Gateway - 网关错误"
    503: "Service Unavailable - 服务不可用"
```

---

## 6. Clash API集成方案

### 6.1 Clash API功能分析

根据 [sing-box Clash API文档](https://sing-box.sagernet.org/zh/configuration/experimental/clash-api/)，主要功能包括：

#### 核心功能
```yaml
clash_api_features:
  proxy_management:
    - 获取代理列表      # GET /proxies
    - 获取代理组        # GET /proxies/{name}
    - 选择代理          # PUT /proxies/{name}
    - 代理延迟测试      # GET /proxies/{name}/delay
  
  traffic_control:
    - 切换代理模式      # PUT /configs (Rule/Global/Direct)
    - 流量统计         # GET /traffic
    - 连接管理         # GET /connections
    - 关闭连接         # DELETE /connections/{id}
  
  configuration:
    - 获取配置         # GET /configs
    - 重载配置         # PUT /configs
    - 获取规则         # GET /rules
  
  monitoring:
    - 实时日志         # GET /logs (WebSocket)
    - 系统信息         # GET /version
```

### 6.2 集成架构设计

#### Agent端Clash API代理

```go
// internal/agent/clash/proxy.go
type ClashAPIProxy struct {
    baseURL     string
    secret      string
    client      *http.Client
    logger      *logrus.Logger
    
    // WebSocket连接池
    wsConnections map[string]*websocket.Conn
    wsLock        sync.RWMutex
}

// 代理HTTP请求
func (p *ClashAPIProxy) ProxyHTTPRequest(ctx context.Context, req *ClashAPIRequest) (*ClashAPIResponse, error) {
    // 1. 构建目标URL
    targetURL := fmt.Sprintf("%s%s", p.baseURL, req.Path)
    
    // 2. 创建HTTP请求
    httpReq, err := http.NewRequestWithContext(ctx, req.Method, targetURL, bytes.NewReader(req.Body))
    if err != nil {
        return nil, err
    }
    
    // 3. 添加认证头
    if p.secret != "" {
        httpReq.Header.Set("Authorization", "Bearer "+p.secret)
    }
    
    // 4. 复制请求头
    for k, v := range req.Headers {
        httpReq.Header.Set(k, v)
    }
    
    // 5. 发送请求
    resp, err := p.client.Do(httpReq)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    // 6. 读取响应
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, err
    }
    
    return &ClashAPIResponse{
        StatusCode: resp.StatusCode,
        Headers:    convertHeaders(resp.Header),
        Body:       body,
    }, nil
}

// 代理WebSocket连接
func (p *ClashAPIProxy) ProxyWebSocket(ctx context.Context, req *WebSocketRequest) error {
    // WebSocket升级和代理逻辑
    // ...
}
```

#### API服务端Clash管理

```go
// internal/api/service/clash/service.go
type ClashService struct {
    nodeManager  NodeManager
    agentClients map[string]AgentClient
    logger       *logrus.Logger
}

func (s *ClashService) ProxyClashAPI(ctx context.Context, nodeID string, req *ClashAPIRequest) (*ClashAPIResponse, error) {
    // 1. 获取节点Agent客户端
    client, exists := s.agentClients[nodeID]
    if !exists {
        return nil, ErrNodeNotConnected
    }
    
    // 2. 通过gRPC调用Agent
    grpcReq := &pb.ClashAPIRequest{
        Method:  req.Method,
        Path:    req.Path,
        Headers: req.Headers,
        Body:    req.Body,
    }
    
    grpcResp, err := client.ProxyClashAPI(ctx, grpcReq)
    if err != nil {
        return nil, err
    }
    
    // 3. 转换响应格式
    return &ClashAPIResponse{
        StatusCode: int(grpcResp.StatusCode),
        Headers:    grpcResp.Headers,
        Body:       grpcResp.Body,
    }, nil
}

func (s *ClashService) GetProxyStatus(ctx context.Context, nodeID string) (*ProxyStatus, error) {
    // 获取节点代理状态
    req := &ClashAPIRequest{
        Method: "GET",
        Path:   "/proxies",
    }
    
    resp, err := s.ProxyClashAPI(ctx, nodeID, req)
    if err != nil {
        return nil, err
    }
    
    // 解析代理状态
    var proxies map[string]interface{}
    if err := json.Unmarshal(resp.Body, &proxies); err != nil {
        return nil, err
    }
    
    return &ProxyStatus{
        Mode:     extractProxyMode(proxies),
        Current:  extractCurrentProxy(proxies),
        Proxies:  convertProxyList(proxies),
    }, nil
}
```

#### Web端API代理

```go
// internal/web/handlers/proxy.go
func (h *ProxyHandler) HandleClashAPI(c *gin.Context) {
    nodeID := c.Param("node_id")
    
    // 构建代理请求
    req := &ClashAPIRequest{
        Method:  c.Request.Method,
        Path:    c.Param("path"),
        Headers: convertGinHeaders(c.Request.Header),
    }
    
    // 读取请求体
    if c.Request.Body != nil {
        body, err := io.ReadAll(c.Request.Body)
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
            return
        }
        req.Body = body
    }
    
    // 调用API服务
    resp, err := h.apiClient.ProxyClashAPI(c.Request.Context(), nodeID, req)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    
    // 返回代理响应
    for k, v := range resp.Headers {
        c.Header(k, v)
    }
    c.Data(resp.StatusCode, "application/json", resp.Body)
}
```

### 6.3 Clash API路由设计

#### RESTful API路由
```go
// Web服务器路由
v1.GET("/nodes/:node_id/clash/proxies", handlers.GetProxies)
v1.PUT("/nodes/:node_id/clash/proxies/:name", handlers.SelectProxy)
v1.GET("/nodes/:node_id/clash/proxies/:name/delay", handlers.TestDelay)
v1.GET("/nodes/:node_id/clash/traffic", handlers.GetTraffic)
v1.GET("/nodes/:node_id/clash/connections", handlers.GetConnections)
v1.DELETE("/nodes/:node_id/clash/connections/:id", handlers.CloseConnection)
v1.PUT("/nodes/:node_id/clash/configs", handlers.UpdateConfig)

// WebSocket路由
v1.GET("/nodes/:node_id/clash/logs", handlers.StreamLogs)
```

---

## 7. 监控与可观测性

### 7.1 Prometheus + Grafana

#### 监控指标设计

```go
// pkg/metrics/prometheus/metrics.go
var (
    // HTTP请求指标
    httpRequestsTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "http_requests_total",
            Help: "Total number of HTTP requests",
        },
        []string{"method", "endpoint", "status_code"},
    )
    
    // gRPC请求指标
    grpcRequestsTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "grpc_requests_total", 
            Help: "Total number of gRPC requests",
        },
        []string{"service", "method", "status_code"},
    )
    
    // 节点状态指标
    nodeStatusGauge = prometheus.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "node_status",
            Help: "Node status (1=online, 0=offline)",
        },
        []string{"node_id", "node_name", "region"},
    )
    
    // Clash API指标
    clashAPIRequestsTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "clash_api_requests_total",
            Help: "Total number of Clash API requests",
        },
        []string{"node_id", "api_path", "method", "status_code"},
    )
)
```

#### Grafana Dashboard配置

```yaml
# grafana/dashboards/sing-box-overview.json
dashboard:
  title: "sing-box-web Overview"
  panels:
    - title: "Node Status"
      type: "stat"
      targets:
        - expr: "sum(node_status)"
    
    - title: "Request Rate"
      type: "graph"
      targets:
        - expr: "rate(http_requests_total[5m])"
    
    - title: "Clash API Usage"
      type: "graph"
      targets:
        - expr: "rate(clash_api_requests_total[5m])"
    
    - title: "Error Rate"
      type: "graph"
      targets:
        - expr: "rate(http_requests_total{status_code=~'5..'}[5m])"
```

### 7.2 链路追踪：SkyWalking Go Agent

#### SkyWalking Go Agent架构

SkyWalking Go Agent采用编译时自动instrument技术，通过Go的`-toolexec`参数在编译过程中自动注入追踪代码，无需显式导入依赖包。

**支持的框架和库**：
- **HTTP框架**: `net/http`, `gin-gonic/gin`, `gorilla/mux`, `labstack/echo`, `go-chi/chi`
- **数据库**: `database/sql`, `gorm.io/gorm`, `go-sql-driver/mysql`, `lib/pq`
- **gRPC**: `google.golang.org/grpc`
- **Redis**: `go-redis/redis`, `gomodule/redigo`
- **Kafka**: `Shopify/sarama`, `segmentio/kafka-go`
- **Elasticsearch**: `elastic/go-elasticsearch`

#### 安装SkyWalking Go Agent

```bash
# 下载SkyWalking Go Agent
go install github.com/apache/skywalking-go/tools/go-agent@latest

# 或者从GitHub releases下载二进制文件
wget https://github.com/apache/skywalking-go/releases/download/v0.4.0/skywalking-go-agent-0.4.0-linux-amd64.tgz
tar -xzf skywalking-go-agent-0.4.0-linux-amd64.tgz
```

#### 构建配置

```makefile
# Makefile中的构建配置
# SkyWalking Go Agent路径
SKYWALKING_AGENT_PATH := $(shell which go-agent)

# 环境变量配置
export SW_AGENT_NAME ?= sing-box-web
export SW_AGENT_COLLECTOR_BACKEND_SERVICES ?= skywalking-oap:11800
export SW_AGENT_SAMPLE_N_PER_3_SECS ?= -1

# 构建sing-box-web (自动instrument)
build-web:
	@echo "Building sing-box-web with SkyWalking agent..."
	@CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build \
		-toolexec="$(SKYWALKING_AGENT_PATH)" \
		-ldflags "$(LDFLAGS)" \
		-o $(BUILD_DIR)/sing-box-web \
		./cmd/sing-box-web

# 构建sing-box-api (自动instrument)
build-api:
	@echo "Building sing-box-api with SkyWalking agent..."
	@CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build \
		-toolexec="$(SKYWALKING_AGENT_PATH)" \
		-ldflags "$(LDFLAGS)" \
		-o $(BUILD_DIR)/sing-box-api \
		-tags skywalking \
		./cmd/sing-box-api

# 构建sing-box-agent (自动instrument)
build-agent:
	@echo "Building sing-box-agent with SkyWalking agent..."
	@CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build \
		-toolexec="$(SKYWALKING_AGENT_PATH)" \
		-ldflags "$(LDFLAGS)" \
		-o $(BUILD_DIR)/sing-box-agent \
		./cmd/sing-box-agent

# 不使用SkyWalking的构建（开发/调试用）
build-no-tracing:
	@echo "Building without SkyWalking agent..."
	@CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/sing-box-web ./cmd/sing-box-web
	@CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/sing-box-api ./cmd/sing-box-api
	@CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/sing-box-agent ./cmd/sing-box-agent
```

#### 应用代码（无需修改）

由于使用自动instrument，应用代码保持原样，无需添加追踪相关代码：

```go
// cmd/sing-box-web/main.go - 无需修改
func main() {
    // 正常的应用代码，SkyWalking会自动instrument
    r := gin.Default()
    
    // 正常注册中间件和路由
    r.Use(middleware.Logger())
    r.Use(middleware.Recovery())
    
    setupRoutes(r)
    
    // 启动服务器
    r.Run(":3000")
}

// 正常的gRPC服务器代码
func NewGRPCServer() *grpc.Server {
    return grpc.NewServer()  // SkyWalking会自动添加拦截器
}

// 正常的数据库连接代码
func NewDBConnection(dsn string) (*gorm.DB, error) {
    return gorm.Open(postgres.Open(dsn), &gorm.Config{})  // 自动instrument
}

// 正常的HTTP客户端代码
func CallAPI(url string) error {
    resp, err := http.Get(url)  // 自动instrument
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    return nil
}
```

#### 环境变量配置

```yaml
# SkyWalking Go Agent配置
environment:
  # 基础配置
  SW_AGENT_NAME: "sing-box-web"                    # 服务名称
  SW_AGENT_INSTANCE_NAME: "web-001"                # 实例名称
  SW_AGENT_COLLECTOR_BACKEND_SERVICES: "skywalking-oap:11800"  # OAP地址
  
  # 采样配置
  SW_AGENT_SAMPLE_N_PER_3_SECS: "-1"               # 采样率(-1=全采样, 0=不采样)
  
  # 认证和分组
  SW_AGENT_AUTHENTICATION: ""                      # 认证token
  SW_AGENT_NAMESPACE: "production"                 # 命名空间
  SW_AGENT_CLUSTER: "sing-box-cluster"             # 集群名称
  
  # 高级配置
  SW_AGENT_MAX_SEGMENT_SIZE: "300"                 # 最大segment大小
  SW_AGENT_DISABLE_PLUGINS: ""                     # 禁用插件列表(逗号分隔)
  SW_AGENT_LOG_LEVEL: "INFO"                       # 日志级别
  SW_AGENT_LOG_FILE_PATH: "/var/log/skywalking/agent.log"  # 日志文件路径
```

#### Docker构建配置

```dockerfile
# cmd/sing-box-web/Dockerfile
FROM golang:1.21-alpine AS builder

# 安装SkyWalking Go Agent
RUN wget -O /tmp/skywalking-go-agent.tgz \
    https://github.com/apache/skywalking-go/releases/download/v0.4.0/skywalking-go-agent-0.4.0-linux-amd64.tgz && \
    tar -xzf /tmp/skywalking-go-agent.tgz -C /usr/local/bin && \
    chmod +x /usr/local/bin/go-agent

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# 使用SkyWalking Go Agent构建
ENV SW_AGENT_NAME=sing-box-web
ENV SW_AGENT_COLLECTOR_BACKEND_SERVICES=skywalking-oap:11800
RUN CGO_ENABLED=0 GOOS=linux go build \
    -toolexec="/usr/local/bin/go-agent" \
    -a -installsuffix cgo \
    -ldflags "-X main.version=${VERSION}" \
    -o sing-box-web ./cmd/sing-box-web

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/

COPY --from=builder /app/sing-box-web .
COPY --from=builder /app/configs ./configs

EXPOSE 3000
CMD ["./sing-box-web"]
```

#### 支持情况说明

**自动支持的功能**：
- ✅ Gin HTTP路由和中间件
- ✅ gRPC服务端和客户端
- ✅ GORM数据库操作
- ✅ Redis操作（go-redis）
- ✅ 标准库http.Client

**需要手动处理的场景**：
- ❌ WebSocket连接（需要手动span）
- ❌ 自定义RPC协议
- ❌ 文件操作
- ❌ 第三方SDK调用

#### 手动Span示例（不支持的场景）

```go
// pkg/tracing/manual.go
// 为不支持自动instrument的场景提供手动span

import (
    "context"
    "github.com/SkyAPM/go2sky"  // 仅用于手动span创建
)

// WebSocket处理器手动添加span
func HandleWebSocket(w http.ResponseWriter, r *http.Request) {
    // 从context获取当前span（由自动instrument创建）
    span := go2sky.GetSpan(r.Context())
    if span != nil {
        span.SetTag("ws.upgrade", "true")
        span.Log("WebSocket connection established")
    }
    
    // WebSocket升级和处理逻辑
    // ...
}

// 文件操作手动span
func ProcessConfigFile(ctx context.Context, filename string) error {
    span, ctx := go2sky.GetTracer().CreateLocalSpan(ctx, "process-config-file")
    if span != nil {
        defer span.End()
        span.SetTag("file.name", filename)
    }
    
    // 文件处理逻辑
    // ...
    
    return nil
}

// 第三方API调用手动span
func CallThirdPartyAPI(ctx context.Context, url string) error {
    span, ctx := go2sky.GetTracer().CreateExitSpan(ctx, "third-party-api", url, 
        func(header string) error {
            // 注入trace header到HTTP请求
            return nil
        })
    if span != nil {
        defer span.End()
    }
    
    // API调用逻辑
    // ...
    
    return nil
}
```

### 7.3 日志聚合：ELK Stack

```yaml
# logging configuration
logging:
  level: info
  format: json
  outputs:
    - type: stdout
    - type: file
      filename: /var/log/sing-box-web/app.log
      max_size: 100MB
      max_backups: 10
      max_age: 30
    - type: elasticsearch
      addresses:
        - http://localhost:9200
      index: sing-box-web
```

---

## 8. 部署架构

### 8.1 容器化部署：Docker

#### Multi-stage Dockerfile

```dockerfile
# cmd/sing-box-web/Dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo \
    -ldflags "-X main.version=${VERSION} -X main.buildTime=${BUILD_TIME}" \
    -o sing-box-web ./cmd/sing-box-web

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/

COPY --from=builder /app/sing-box-web .
COPY --from=builder /app/configs ./configs
COPY --from=builder /app/ui/dist ./ui/dist

EXPOSE 3000
CMD ["./sing-box-web", "serve"]
```

#### Docker Compose配置

```yaml
# docker-compose.yml
version: '3.8'

services:
  # 数据库服务
  postgres:
    image: postgres:15-alpine
    environment:
      POSTGRES_DB: sing_box_web
      POSTGRES_USER: sing_box_user
      POSTGRES_PASSWORD: ${POSTGRES_PASSWORD}
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./scripts/init.sql:/docker-entrypoint-initdb.d/init.sql
    ports:
      - "5432:5432"

  redis:
    image: redis:7-alpine
    command: redis-server --requirepass ${REDIS_PASSWORD}
    volumes:
      - redis_data:/data
    ports:
      - "6379:6379"

  # SkyWalking服务
  elasticsearch:
    image: docker.elastic.co/elasticsearch/elasticsearch:7.17.0
    environment:
      - discovery.type=single-node
      - bootstrap.memory_lock=true
      - "ES_JAVA_OPTS=-Xms512m -Xmx512m"
    ulimits:
      memlock:
        soft: -1
        hard: -1
    volumes:
      - elasticsearch_data:/usr/share/elasticsearch/data
    ports:
      - "9200:9200"

  skywalking-oap:
    image: apache/skywalking-oap-server:9.7.0
    environment:
      SW_STORAGE: elasticsearch
      SW_STORAGE_ES_CLUSTER_NODES: elasticsearch:9200
      SW_HEALTH_CHECKER: default
      SW_TELEMETRY: prometheus
      JAVA_OPTS: "-Xms512m -Xmx512m"
    ports:
      - "11800:11800"  # gRPC
      - "12800:12800"  # HTTP
    depends_on:
      - elasticsearch

  skywalking-ui:
    image: apache/skywalking-ui:9.7.0
    environment:
      SW_OAP_ADDRESS: http://skywalking-oap:12800
    ports:
      - "8080:8080"
    depends_on:
      - skywalking-oap

  # 应用服务
  sing-box-api:
    build:
      context: .
      dockerfile: cmd/sing-box-api/Dockerfile
    environment:
      - DATABASE_URL=postgres://sing_box_user:${POSTGRES_PASSWORD}@postgres:5432/sing_box_web
      - REDIS_URL=redis://:${REDIS_PASSWORD}@redis:6379/0
      # SkyWalking配置
      - SW_AGENT_NAME=sing-box-api
      - SW_AGENT_INSTANCE_NAME=api-001
      - SW_AGENT_COLLECTOR_BACKEND_SERVICES=skywalking-oap:11800
      - SW_AGENT_SAMPLE_N_PER_3_SECS=-1
      - SW_AGENT_NAMESPACE=production
      - SW_AGENT_CLUSTER=sing-box-cluster
    ports:
      - "8081:8080"  # HTTP管理接口
      - "9090:9090"  # gRPC接口
    depends_on:
      - postgres
      - redis
      - skywalking-oap

  sing-box-web:
    build:
      context: .
      dockerfile: cmd/sing-box-web/Dockerfile
    environment:
      - API_SERVER_URL=grpc://sing-box-api:9090
      - REDIS_URL=redis://:${REDIS_PASSWORD}@redis:6379/0
      # SkyWalking配置
      - SW_AGENT_NAME=sing-box-web
      - SW_AGENT_INSTANCE_NAME=web-001
      - SW_AGENT_COLLECTOR_BACKEND_SERVICES=skywalking-oap:11800
      - SW_AGENT_SAMPLE_N_PER_3_SECS=-1
      - SW_AGENT_NAMESPACE=production
      - SW_AGENT_CLUSTER=sing-box-cluster
    ports:
      - "3000:3000"
    depends_on:
      - sing-box-api
      - skywalking-oap

  # 监控服务
  prometheus:
    image: prom/prometheus:latest
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--storage.tsdb.path=/prometheus'
      - '--web.console.libraries=/etc/prometheus/console_libraries'
      - '--web.console.templates=/etc/prometheus/consoles'
      - '--web.enable-lifecycle'
    ports:
      - "9090:9090"
    volumes:
      - ./monitoring/prometheus.yml:/etc/prometheus/prometheus.yml
      - prometheus_data:/prometheus

  grafana:
    image: grafana/grafana:latest
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=${GRAFANA_PASSWORD}
    ports:
      - "3001:3000"
    volumes:
      - grafana_data:/var/lib/grafana
      - ./monitoring/grafana/dashboards:/etc/grafana/provisioning/dashboards
      - ./monitoring/grafana/datasources:/etc/grafana/provisioning/datasources

volumes:
  postgres_data:
  redis_data:
  elasticsearch_data:
  prometheus_data:
  grafana_data:
```

### 8.2 Kubernetes部署

```yaml
# deployments/k8s/sing-box-web.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: sing-box-web
  labels:
    app: sing-box-web
spec:
  replicas: 3
  selector:
    matchLabels:
      app: sing-box-web
  template:
    metadata:
      labels:
        app: sing-box-web
    spec:
      containers:
      - name: sing-box-web
        image: sing-box-web:latest
        ports:
        - containerPort: 3000
        env:
        - name: API_SERVER_URL
          value: "grpc://sing-box-api:9090"
        resources:
          requests:
            memory: "64Mi"
            cpu: "250m"
          limits:
            memory: "128Mi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /health
            port: 3000
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /ready
            port: 3000
          initialDelaySeconds: 5
          periodSeconds: 5
```

---

## 9. 开发工具链

### 9.1 代码质量工具

#### golangci-lint配置

```yaml
# .golangci.yml
run:
  timeout: 5m
  modules-download-mode: readonly

linters:
  enable:
    - errcheck
    - gosimple
    - govet
    - ineffassign
    - staticcheck
    - typecheck
    - unused
    - varcheck
    - structcheck
    - deadcode
    - gocyclo
    - gofmt
    - goimports
    - gosec
    - unconvert
    - dupl
    - goconst
    - gocognit

linters-settings:
  gocyclo:
    min-complexity: 15
  dupl:
    threshold: 100
  goconst:
    min-len: 2
    min-occurrences: 2
```

### 9.2 构建工具：Makefile

```makefile
# 主要构建目标
.PHONY: build test lint docker-build

# 变量定义
VERSION ?= $(shell git describe --tags --always --dirty)
BUILD_TIME ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS = -X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME)

# SkyWalking Go Agent配置
SKYWALKING_AGENT_PATH := $(shell which go-agent)

# 环境变量配置
export SW_AGENT_NAME ?= sing-box-web
export SW_AGENT_COLLECTOR_BACKEND_SERVICES ?= skywalking-oap:11800
export SW_AGENT_SAMPLE_N_PER_3_SECS ?= -1

# 构建所有应用（默认使用SkyWalking）
build: build-web build-api build-agent

# 使用SkyWalking Go Agent构建
build-web:
	@echo "Building sing-box-web with SkyWalking agent..."
	@CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build \
		-toolexec="$(SKYWALKING_AGENT_PATH)" \
		-ldflags "$(LDFLAGS)" \
		-o bin/sing-box-web \
		./cmd/sing-box-web

build-api:
	@echo "Building sing-box-api with SkyWalking agent..."
	@CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build \
		-toolexec="$(SKYWALKING_AGENT_PATH)" \
		-ldflags "$(LDFLAGS)" \
		-o bin/sing-box-api \
		./cmd/sing-box-api

build-agent:
	@echo "Building sing-box-agent with SkyWalking agent..."
	@CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build \
		-toolexec="$(SKYWALKING_AGENT_PATH)" \
		-ldflags "$(LDFLAGS)" \
		-o bin/sing-box-agent \
		./cmd/sing-box-agent

# 不使用SkyWalking的构建（开发/调试用）
build-no-tracing:
	@echo "Building without SkyWalking agent..."
	@CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)" -o bin/sing-box-web ./cmd/sing-box-web
	@CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)" -o bin/sing-box-api ./cmd/sing-box-api
	@CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)" -o bin/sing-box-agent ./cmd/sing-box-agent

# 测试
test:
	go test -race -coverprofile=coverage.out ./...

test-integration:
	go test -tags=integration ./test/integration/...

# 代码检查
lint:
	golangci-lint run

# Docker构建
docker-build:
	docker build -f cmd/sing-box-web/Dockerfile -t sing-box-web:$(VERSION) .
	docker build -f cmd/sing-box-api/Dockerfile -t sing-box-api:$(VERSION) .
	docker build -f cmd/sing-box-agent/Dockerfile -t sing-box-agent:$(VERSION) .

# protobuf生成
proto-gen:
	protoc --go_out=. --go-grpc_out=. api/proto/**/*.proto

# 数据库迁移
migrate-up:
	migrate -path migrations -database "$(DATABASE_URL)" up

migrate-down:
	migrate -path migrations -database "$(DATABASE_URL)" down

# 开发环境
dev-up:
	docker-compose -f docker-compose.dev.yml up -d

dev-down:
	docker-compose -f docker-compose.dev.yml down
```

---

## 10. 安全考虑

### 10.1 认证与授权

```go
// JWT配置
jwt:
  secret: "${JWT_SECRET}"
  expire_time: 24h
  refresh_time: 7d
  issuer: "sing-box-web"

// RBAC权限模型
permissions:
  node_management:
    - "node:read"
    - "node:write"
    - "node:delete"
  
  config_management:
    - "config:read"
    - "config:write"
    - "config:deploy"
  
  clash_api:
    - "clash:proxy"
    - "clash:traffic"
    - "clash:config"
```

### 10.2 TLS/SSL配置

```yaml
tls:
  enabled: true
  cert_file: "/etc/ssl/certs/sing-box-web.crt"
  key_file: "/etc/ssl/private/sing-box-web.key"
  min_version: "1.3"
  cipher_suites:
    - "TLS_AES_128_GCM_SHA256"
    - "TLS_AES_256_GCM_SHA384"
    - "TLS_CHACHA20_POLY1305_SHA256"
```

### 10.3 API安全

```go
// API限流
rate_limit:
  enabled: true
  requests_per_minute: 100
  burst: 10

// CORS配置
cors:
  allowed_origins:
    - "https://sing-box.example.com"
  allowed_methods:
    - "GET"
    - "POST"
    - "PUT"
    - "DELETE"
  allowed_headers:
    - "Authorization"
    - "Content-Type"
  max_age: 86400
```

---

## 11. 总结

### 11.1 技术栈优势

这个技术栈为sing-box-web项目提供了：

#### 性能优势
- **Go语言**：高并发、低延迟、内存安全
- **gRPC**：高效的二进制协议，支持流式处理
- **PostgreSQL**：强ACID特性，高并发读写
- **Redis**：内存缓存，微秒级响应
- **TimescaleDB**：时序数据专业处理

#### 开发效率
- **Gin框架**：简洁API，快速开发
- **GORM**：强大的ORM工具
- **Protocol Buffers**：强类型接口定义
- **丰富生态**：Go生态系统成熟

#### 运维友好
- **单二进制部署**：无运行时依赖
- **容器化支持**：Docker + Kubernetes
- **完整监控**：Prometheus + Grafana + SkyWalking
- **自动化构建**：完整的CI/CD流程

#### 扩展性
- **微服务架构**：清晰的服务边界
- **水平扩展**：支持多实例部署
- **插件化设计**：可插拔的组件架构
- **API版本管理**：平滑的版本升级

### 11.2 Clash API集成优势

- **透明代理**：Agent端完整代理Clash API
- **统一管理**：通过API服务统一调度
- **实时监控**：WebSocket支持实时日志
- **安全控制**：gRPC传输加密，API访问控制

### 11.3 未来扩展计划

1. **多租户支持**：企业级多租户管理
2. **插件系统**：第三方插件扩展
3. **移动端支持**：React Native移动应用
4. **AI集成**：智能配置推荐和故障诊断
5. **边缘计算**：边缘节点管理和调度

这个技术栈充分考虑了性能、可维护性、扩展性和安全性，为构建企业级sing-box管理平台提供了坚实的技术基础。 