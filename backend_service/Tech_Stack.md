# 技术选型建议 - sing-box-web

## 1. 技术选型概述

基于 sing-box-web 项目的功能需求、性能要求和团队技术背景，我们推荐以下核心技术栈：

| 技术类别 | 推荐选择 | 版本要求 | 选择理由 |
|---------|---------|---------|---------|
| **核心语言** | Go | 1.21+ | 高性能、并发优秀、生态丰富 |
| **Web框架** | Gin | v1.9+ | 轻量高效、中间件丰富、社区活跃 |
| **RPC框架** | gRPC | v1.58+ | 高性能、类型安全、双向流支持 |
| **数据库** | PostgreSQL | 15+ | 功能完整、JSON支持、高并发 |
| **缓存** | Redis | 7.0+ | 高性能、数据结构丰富 |
| **监控** | Prometheus + Grafana | Latest | 开源标准、功能强大 |
| **日志** | Logrus + ELK | Latest | 结构化日志、集中收集 |

---

## 2. 核心技术详细分析

### 2.1 编程语言：Go

#### 选择理由
- **并发优势**: Goroutine 轻量级线程，天然适合处理大量 Agent 连接
- **网络性能**: 优秀的网络库，适合实时通信和 gRPC 服务
- **部署便利**: 单二进制文件，容器化部署简单
- **生态丰富**: gRPC、数据库驱动、监控工具等库成熟
- **团队匹配**: 与 sing-box 项目技术栈一致，降低学习成本

#### 版本建议
- **最低版本**: Go 1.21（支持最新语言特性和性能优化）
- **推荐版本**: Go 1.22+（更好的性能和工具链）

#### 替代方案对比
| 语言 | 优势 | 劣势 | 适用场景 |
|------|------|------|---------|
| **Rust** | 极致性能、内存安全 | 学习曲线陡峭、开发效率低 | 对性能要求极高的场景 |
| **Java** | 生态成熟、企业级支持 | 资源占用大、启动慢 | 大型企业系统 |
| **Node.js** | 开发效率高、JSON原生 | 单线程瓶颈、CPU密集型弱 | 轻量级API服务 |

### 2.2 Web框架：Gin

#### 选择理由
- **性能优秀**: 基于 httprouter，路由性能极佳
- **中间件丰富**: 认证、CORS、日志、限流等中间件完善
- **开发效率**: API简洁，学习成本低
- **社区活跃**: 大量第三方插件和示例

#### 核心特性
```go
// 路由定义示例
r := gin.Default()

// 中间件
r.Use(middleware.CORS())
r.Use(middleware.JWT())
r.Use(middleware.RequestID())

// 路由组
v1 := r.Group("/v1")
{
    nodes := v1.Group("/nodes")
    {
        nodes.GET("", nodeHandler.List)
        nodes.POST("", nodeHandler.Create)
        nodes.GET("/:id", nodeHandler.Get)
        nodes.PUT("/:id", nodeHandler.Update)
        nodes.DELETE("/:id", nodeHandler.Delete)
    }
}
```

#### 替代方案
| 框架 | 优势 | 劣势 | 选择建议 |
|------|------|------|---------|
| **Echo** | 性能相近、功能丰富 | 社区相对较小 | 可选替代 |
| **Fiber** | 性能最优、Express风格 | 较新、生态待完善 | 性能优先场景 |
| **标准库** | 零依赖、完全控制 | 开发效率低、样板代码多 | 极简场景 |

### 2.3 RPC框架：gRPC

#### 选择理由
- **高性能**: HTTP/2 协议，二进制传输，压缩效率高
- **类型安全**: Protobuf 强类型定义，避免接口不一致
- **双向流**: 支持 Agent-Manager 实时通信需求
- **跨语言**: 便于未来多语言客户端开发
- **流量控制**: 内置背压、超时、重试机制

#### 服务定义示例
```protobuf
service ManagerService {
  // 双向流连接
  rpc AgentStream(stream AgentMessage) returns (stream ManagerMessage);
  
  // 标准RPC调用
  rpc GetNodes(GetNodesRequest) returns (GetNodesResponse);
  rpc UpdateNodeConfig(UpdateNodeConfigRequest) returns (UpdateNodeConfigResponse);
}
```

#### 性能特点
- **吞吐量**: 相比 REST API 提升 2-3 倍
- **延迟**: 二进制序列化，延迟降低 30-50%
- **连接复用**: HTTP/2 多路复用，减少连接开销

#### 替代方案
| 协议 | 优势 | 劣势 | 适用场景 |
|------|------|------|---------|
| **WebSocket** | 实时性好、浏览器原生支持 | 协议简单、需要自定义消息格式 | 浏览器实时通信 |
| **TCP Socket** | 性能最优、完全控制 | 开发复杂、需要处理粘包等问题 | 极致性能要求 |
| **HTTP Long Polling** | 实现简单、兼容性好 | 资源效率低、延迟高 | 简单实时通知 |

### 2.4 数据库：PostgreSQL

#### 选择理由
- **JSON支持**: 原生 JSONB 类型，适合存储节点元数据和配置
- **并发性能**: MVCC 机制，读写并发性能优秀
- **功能完整**: 支持分区、索引、存储过程、窗口函数
- **数据一致性**: ACID 事务，保证数据安全
- **扩展性**: 支持插件、自定义类型和函数

#### 数据类型优势
```sql
-- JSONB 高效存储和查询
CREATE TABLE nodes (
    id UUID PRIMARY KEY,
    name VARCHAR(100),
    labels JSONB,
    metadata JSONB
);

-- GIN 索引支持 JSON 查询
CREATE INDEX idx_nodes_labels ON nodes USING GIN(labels);

-- 高效 JSON 查询
SELECT * FROM nodes WHERE labels @> '{"environment": "production"}';
```

#### 性能配置
```postgresql
# postgresql.conf 优化建议
shared_buffers = 256MB          # 25% of RAM
effective_cache_size = 1GB      # 75% of RAM
random_page_cost = 1.1          # SSD 优化
checkpoint_completion_target = 0.9
wal_buffers = 16MB
max_connections = 200
```

#### 替代方案
| 数据库 | 优势 | 劣势 | 适用场景 |
|-------|------|------|---------|
| **SQLite** | 零配置、单文件 | 并发写入限制、功能有限 | 单机小规模部署 |
| **MySQL** | 生态成熟、运维简单 | JSON 支持较弱、锁机制限制 | 传统 Web 应用 |
| **MongoDB** | 文档存储、水平扩展 | 事务支持弱、内存占用大 | 文档型数据 |

### 2.5 缓存：Redis

#### 选择理由
- **数据结构**: 支持 String、Hash、List、Set、ZSet，适合多种缓存场景
- **高性能**: 内存存储，单线程模型避免锁竞争
- **持久化**: RDB + AOF 双重保障
- **集群支持**: Redis Cluster 水平扩展

#### 使用场景
```go
// 会话缓存
redis.Set("session:"+tokenHash, userID, 24*time.Hour)

// 监控数据缓存（最近 1 小时）
redis.ZAdd("metrics:"+nodeID, redis.Z{
    Score:  float64(time.Now().Unix()),
    Member: metricsJSON,
})

// 分布式锁
redis.SetNX("lock:config:"+nodeID, "locked", 30*time.Second)
```

#### 配置建议
```redis
# redis.conf 优化
maxmemory 512mb
maxmemory-policy allkeys-lru
save 900 1
save 300 10
save 60 10000
```

#### 替代方案
| 缓存方案 | 优势 | 劣势 | 适用场景 |
|---------|------|------|---------|
| **Memcached** | 性能极优、内存效率高 | 功能单一、无持久化 | 纯缓存场景 |
| **内存缓存** | 零延迟、零网络开销 | 无法共享、重启丢失 | 单机缓存 |
| **Etcd** | 强一致性、分布式 | 性能较低、使用复杂 | 配置存储 |

---

## 3. 监控与可观测性

### 3.1 监控体系：Prometheus + Grafana

#### 架构设计
```yaml
# 监控架构
监控数据流:
  sing-box-api → Prometheus 指标 → Prometheus Server → Grafana 展示
  
关键指标:
  - 系统指标: CPU、内存、磁盘、网络
  - 应用指标: API响应时间、错误率、并发数
  - 业务指标: 节点在线率、配置成功率、Agent连接数
```

#### Prometheus 指标定义
```go
// 业务指标示例
var (
    // 节点状态统计
    nodeStatusGauge = prometheus.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "singbox_node_status",
            Help: "Node status (1=online, 0=offline)",
        },
        []string{"node_id", "node_name", "status"},
    )
    
    // 配置部署耗时
    configDeployDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "singbox_config_deploy_duration_seconds",
            Help: "Configuration deployment duration",
            Buckets: prometheus.DefBuckets,
        },
        []string{"node_id", "status"},
    )
    
    // API 请求计数
    apiRequestsTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "singbox_api_requests_total",
            Help: "Total number of API requests",
        },
        []string{"method", "endpoint", "status_code"},
    )
)
```

#### Grafana 仪表盘
- **系统概览**: 节点状态分布、在线率趋势、资源使用概况
- **性能监控**: API 响应时间、错误率、吞吐量
- **业务监控**: 配置部署成功率、Agent 连接状态
- **告警面板**: 异常节点、资源超限、API 错误

### 3.2 日志系统：Logrus + ELK

#### 日志框架配置
```go
// 日志配置示例
log := logrus.New()
log.SetFormatter(&logrus.JSONFormatter{
    TimestampFormat: "2006-01-02T15:04:05.000Z",
    FieldMap: logrus.FieldMap{
        logrus.FieldKeyTime:  "timestamp",
        logrus.FieldKeyLevel: "level",
        logrus.FieldKeyMsg:   "message",
    },
})

// 结构化日志
log.WithFields(logrus.Fields{
    "node_id":       nodeID,
    "deployment_id": deploymentID,
    "duration":      duration,
    "status":        "success",
}).Info("Configuration deployed successfully")
```

#### 日志等级规范
- **ERROR**: 系统错误、配置部署失败、Agent 连接异常
- **WARN**: 节点离线、配置验证警告、性能告警
- **INFO**: 关键业务操作、用户登录、配置变更
- **DEBUG**: 详细的调试信息、gRPC 消息追踪

---

## 4. 开发工具与环境

### 4.1 开发工具链

#### 代码质量工具
```yaml
# .golangci.yml
linters:
  enable:
    - errcheck      # 错误检查
    - gofmt         # 代码格式化
    - goimports     # 导入排序
    - govet         # Go 静态分析
    - ineffassign   # 无效赋值
    - misspell      # 拼写检查
    - unconvert     # 不必要的类型转换
    - unparam       # 未使用的参数
  
linters-settings:
  gofmt:
    simplify: true
  goimports:
    local-prefixes: github.com/your-org/sing-box-web
```

#### 测试工具
```go
// 测试框架组合
func TestNodeService(t *testing.T) {
    // 使用 testify 断言
    assert := assert.New(t)
    
    // 数据库测试（使用 dockertest）
    db, cleanup := setupTestDB(t)
    defer cleanup()
    
    // 服务测试
    service := NewNodeService(db)
    node, err := service.CreateNode(context.Background(), createReq)
    
    assert.NoError(err)
    assert.Equal("test-node", node.Name)
}
```

#### 代码生成工具
```makefile
# Makefile 工具链
generate:
	@echo "Generating protobuf code..."
	@buf generate
	@echo "Generating mocks..."
	@go generate ./...
	@echo "Generating OpenAPI docs..."
	@swag init -g cmd/api/main.go

tools:
	@go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	@go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	@go install github.com/golang/mock/mockgen@latest
	@go install github.com/swaggo/swag/cmd/swag@latest
```

### 4.2 CI/CD 工具选择

#### GitHub Actions 工作流
```yaml
# .github/workflows/ci.yml
name: CI
on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:15
        env:
          POSTGRES_PASSWORD: postgres
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
    
    steps:
    - uses: actions/checkout@v3
    - uses: actions/setup-go@v3
      with:
        go-version: 1.21
    
    - name: Run tests
      run: |
        go mod download
        make test
        make lint
        
    - name: Build
      run: make build
```

#### 替代CI/CD方案
| 方案 | 优势 | 劣势 | 适用场景 |
|------|------|------|---------|
| **GitLab CI** | 功能完整、私有部署 | 资源占用大 | 企业内部 |
| **Jenkins** | 高度可定制、插件丰富 | 配置复杂、维护成本高 | 复杂流水线 |
| **Drone** | 轻量级、容器原生 | 生态相对较小 | 简单 CI/CD |

---

## 5. 部署与容器化

### 5.1 容器化方案：Docker

#### 多阶段构建
```dockerfile
# Dockerfile.api
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o sing-box-api ./cmd/api

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/sing-box-api .
COPY --from=builder /app/configs ./configs

CMD ["./sing-box-api", "serve", "--config-file", "configs/api.yaml"]
```

#### 镜像优化策略
- **多阶段构建**: 减少最终镜像大小（从 ~1GB 降至 ~20MB）
- **Alpine 基础镜像**: 安全性高、体积小
- **非 root 用户**: 提升容器安全性
- **健康检查**: 容器状态监控

### 5.2 编排方案：Kubernetes

#### 部署清单示例
```yaml
# k8s/api-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: sing-box-api
spec:
  replicas: 3
  selector:
    matchLabels:
      app: sing-box-api
  template:
    metadata:
      labels:
        app: sing-box-api
    spec:
      containers:
      - name: api
        image: sing-box-api:latest
        ports:
        - containerPort: 8080
        - containerPort: 9090
        env:
        - name: DB_HOST
          value: postgres-service
        - name: REDIS_HOST
          value: redis-service
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /ready
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
```

#### 替代编排方案
| 方案 | 优势 | 劣势 | 适用场景 |
|------|------|------|---------|
| **Docker Compose** | 简单易用、开发友好 | 单机限制、功能有限 | 开发环境、小规模部署 |
| **Docker Swarm** | 原生集群、学习成本低 | 功能相对简单 | 中小规模集群 |
| **Nomad** | 轻量级、多工作负载 | 生态相对较小 | 混合工作负载 |

---

## 6. 安全考虑

### 6.1 认证与授权

#### JWT 实现
```go
// JWT 配置
type JWTConfig struct {
    SecretKey       string        `yaml:"secret_key"`
    ExpirationTime  time.Duration `yaml:"expiration_time"`
    RefreshTime     time.Duration `yaml:"refresh_time"`
    Issuer          string        `yaml:"issuer"`
}

// 生成 JWT Token
func (j *JWTService) GenerateToken(userID int, username string) (string, error) {
    claims := &JWTClaims{
        UserID:   userID,
        Username: username,
        StandardClaims: jwt.StandardClaims{
            ExpiresAt: time.Now().Add(j.config.ExpirationTime).Unix(),
            Issuer:    j.config.Issuer,
            IssuedAt:  time.Now().Unix(),
        },
    }
    
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString([]byte(j.config.SecretKey))
}
```

#### gRPC 安全
```go
// TLS 配置
func NewTLSCredentials(certFile, keyFile string) (credentials.TransportCredentials, error) {
    cert, err := tls.LoadX509KeyPair(certFile, keyFile)
    if err != nil {
        return nil, err
    }
    
    config := &tls.Config{
        Certificates: []tls.Certificate{cert},
        ClientAuth:   tls.RequireAndVerifyClientCert,
    }
    
    return credentials.NewTLS(config), nil
}
```

### 6.2 数据安全

#### 敏感信息加密
```go
// 密码加密
func HashPassword(password string) (string, error) {
    bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    return string(bytes), err
}

// 配置加密（AES-256-GCM）
func EncryptConfig(plaintext, key []byte) ([]byte, error) {
    block, err := aes.NewCipher(key)
    if err != nil {
        return nil, err
    }
    
    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return nil, err
    }
    
    nonce := make([]byte, gcm.NonceSize())
    if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
        return nil, err
    }
    
    ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
    return ciphertext, nil
}
```

---

## 7. 性能优化建议

### 7.1 数据库优化

#### 连接池配置
```go
// 数据库连接池
config := &gorm.Config{
    Logger: logger.Default.LogMode(logger.Info),
}

db, err := gorm.Open(postgres.Open(dsn), config)
if err != nil {
    return nil, err
}

sqlDB, err := db.DB()
if err != nil {
    return nil, err
}

// 连接池设置
sqlDB.SetMaxOpenConns(25)           // 最大连接数
sqlDB.SetMaxIdleConns(5)            // 最大空闲连接
sqlDB.SetConnMaxLifetime(time.Hour) // 连接最大生命周期
```

#### 查询优化
```sql
-- 索引优化
CREATE INDEX CONCURRENTLY idx_nodes_status_heartbeat 
ON nodes(status, last_heartbeat_at DESC);

-- 分区表（监控数据）
CREATE TABLE node_metrics_y2024m01 PARTITION OF node_metrics
FOR VALUES FROM ('2024-01-01') TO ('2024-02-01');

-- 查询优化
EXPLAIN ANALYZE SELECT * FROM nodes 
WHERE status = 'online' 
AND last_heartbeat_at > NOW() - INTERVAL '5 minutes';
```

### 7.2 应用层优化

#### 缓存策略
```go
// 多级缓存
type CacheService struct {
    local  *bigcache.BigCache    // 本地缓存
    redis  *redis.Client         // 分布式缓存
}

func (c *CacheService) Get(key string) ([]byte, error) {
    // 1. 尝试本地缓存
    if data, err := c.local.Get(key); err == nil {
        return data, nil
    }
    
    // 2. 尝试 Redis 缓存
    if data, err := c.redis.Get(key).Bytes(); err == nil {
        c.local.Set(key, data) // 回写本地缓存
        return data, nil
    }
    
    return nil, ErrNotFound
}
```

#### 并发控制
```go
// 限流器
var limiter = rate.NewLimiter(100, 200) // 100 QPS，突发 200

func RateLimitMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        if !limiter.Allow() {
            c.JSON(429, gin.H{"error": "Rate limit exceeded"})
            c.Abort()
            return
        }
        c.Next()
    }
}

// 工作池
type WorkerPool struct {
    workers int
    jobs    chan Job
}

func (wp *WorkerPool) Start() {
    for i := 0; i < wp.workers; i++ {
        go wp.worker()
    }
}
```

---

## 8. 总结与建议

### 8.1 技术选型总结

| 组件 | 推荐方案 | 关键优势 | 风险评估 |
|------|---------|---------|---------|
| **语言** | Go 1.21+ | 高性能、并发、生态 | ⭐⭐⭐⭐⭐ |
| **Web框架** | Gin | 轻量、高效、易用 | ⭐⭐⭐⭐⭐ |
| **RPC** | gRPC | 类型安全、高性能 | ⭐⭐⭐⭐⭐ |
| **数据库** | PostgreSQL | 功能完整、JSON支持 | ⭐⭐⭐⭐ |
| **缓存** | Redis | 高性能、数据结构丰富 | ⭐⭐⭐⭐ |

### 8.2 实施建议

#### 开发阶段
1. **MVP 阶段**: 使用 SQLite + 单机部署，快速验证功能
2. **测试阶段**: 引入 PostgreSQL + Redis，完善监控
3. **生产阶段**: 容器化部署，添加集群支持

#### 团队建议
- **后端开发**: 2 名 Go 工程师，1 名 DevOps 工程师
- **技能要求**: Go、gRPC、PostgreSQL、Docker、K8s
- **学习路径**: Go 基础 → Web 框架 → gRPC → 数据库优化

#### 风险控制
- **技术风险**: 选择成熟稳定的技术，避免过新的版本
- **性能风险**: 提前进行压力测试，制定扩容方案
- **安全风险**: 实施多层安全防护，定期安全审计

这个技术选型方案既保证了系统的高性能和可扩展性，又兼顾了开发效率和维护成本，为 sing-box-web 项目的成功实施提供了坚实的技术基础。 