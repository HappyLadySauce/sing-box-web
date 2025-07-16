# API 接口定义文档 - sing-box-web

## 1. 文档信息

- **API 版本**: v1.0
- **协议**: HTTPS (Web) + gRPC over TLS (Internal)
- **认证方式**: JWT Bearer Token
- **内容类型**: application/json
- **字符编码**: UTF-8

---

## 2. 系统架构说明

### 2.1 应用分层

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   sing-box-web  │    │  sing-box-api   │    │ sing-box-agent  │
│  (前端服务器)    │◄──►│  (核心API服务)   │◄──►│   (节点代理)     │
│                 │    │                 │    │                 │
│ • 用户管理       │    │ • 节点管理       │    │ • sing-box控制   │
│ • 身份认证       │    │ • 配置管理       │    │ • 系统监控       │
│ • 操作日志       │    │ • 监控数据       │    │ • Clash API集成  │
│ • API代理        │    │ • 业务逻辑       │    │ • 配置应用       │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

### 2.2 API 分类

| API 类型 | 服务提供者 | 使用者 | 协议 | 说明 |
|---------|-----------|--------|------|------|
| **Web API** | sing-box-web | 前端页面 | HTTP/HTTPS | 用户界面相关功能 |
| **Internal gRPC** | sing-box-api | sing-box-web | gRPC | 内部服务通信 |u
| **Clash API** | sing-box-agent | 管理界面 | HTTP | sing-box原生管理接口 |

---

## 3. Web API 规范 (sing-box-web)

### 3.1 OpenAPI 3.0 定义

```yaml
openapi: 3.0.3
info:
  title: sing-box-web Frontend API
  description: |
    sing-box-web 前端服务器 REST API，主要负责用户界面相关功能
    
    ## 认证
    使用JWT Bearer Token进行认证：
    ```
    Authorization: Bearer <your-jwt-token>
    ```
    
    ## 数据流向
    前端 → sing-box-web (认证/日志) → sing-box-api (业务逻辑) → sing-box-agent (执行)
  version: 1.0.0
  contact:
    name: sing-box-web Support
    email: support@example.com

servers:
  - url: https://web.singbox.example.com/api/v1
    description: 生产环境
  - url: https://staging-web.singbox.example.com/api/v1
    description: 测试环境
  - url: http://localhost:3000/api/v1
    description: 本地开发环境

security:
  - bearerAuth: []

components:
  securitySchemes:
    bearerAuth:
      type: http
      scheme: bearer
      bearerFormat: JWT

  schemas:
    # 通用响应模型
    ApiResponse:
      type: object
      properties:
        success:
          type: boolean
          description: 请求是否成功
        message:
          type: string
          description: 响应消息
        data:
          type: object
          description: 响应数据
        timestamp:
          type: string
          format: date-time
          description: 响应时间戳
      required:
        - success
        - timestamp

    ErrorResponse:
      type: object
      properties:
        error:
          type: object
          properties:
            code:
              type: string
              description: 错误代码
            message:
              type: string
              description: 错误消息
            details:
              type: object
              description: 错误详情
          required:
            - code
            - message

    # 分页模型
    PaginationMeta:
      type: object
      properties:
        page:
          type: integer
          minimum: 1
          description: 当前页码
        per_page:
          type: integer
          minimum: 1
          maximum: 100
          description: 每页记录数
        total:
          type: integer
          minimum: 0
          description: 总记录数
        total_pages:
          type: integer
          minimum: 0
          description: 总页数

    # 用户模型
    User:
      type: object
      properties:
        id:
          type: integer
          description: 用户ID
        username:
          type: string
          minLength: 3
          maxLength: 50
          description: 用户名
        email:
          type: string
          format: email
          description: 邮箱地址
        full_name:
          type: string
          maxLength: 100
          description: 全名
        role:
          type: string
          enum: [admin, viewer]
          description: 用户角色
        is_active:
          type: boolean
          description: 是否激活
        created_at:
          type: string
          format: date-time
          description: 创建时间
        updated_at:
          type: string
          format: date-time
          description: 更新时间
        last_login_at:
          type: string
          format: date-time
          nullable: true
          description: 最后登录时间

    # 操作日志模型
    AuditLog:
      type: object
      properties:
        id:
          type: integer
          description: 日志ID
        user_id:
          type: integer
          description: 操作用户ID
        username:
          type: string
          description: 操作用户名
        action:
          type: string
          description: 操作类型
        resource_type:
          type: string
          description: 资源类型
        resource_id:
          type: string
          description: 资源ID
        details:
          type: object
          description: 操作详情
        ip_address:
          type: string
          description: 客户端IP
        user_agent:
          type: string
          description: 用户代理
        success:
          type: boolean
          description: 操作是否成功
        error_message:
          type: string
          nullable: true
          description: 错误消息
        created_at:
          type: string
          format: date-time
          description: 操作时间

paths:
  # 身份认证相关
  /auth/login:
    post:
      summary: 用户登录
      description: 使用用户名/邮箱和密码进行登录
      tags: [身份认证]
      security: []
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              properties:
                username:
                  type: string
                  description: 用户名或邮箱
                password:
                  type: string
                  description: 密码
              required:
                - username
                - password
      responses:
        '200':
          description: 登录成功
          content:
            application/json:
              schema:
                allOf:
                  - $ref: '#/components/schemas/ApiResponse'
                  - type: object
                    properties:
                      data:
                        type: object
                        properties:
                          token:
                            type: string
                            description: JWT访问令牌
                          expires_at:
                            type: string
                            format: date-time
                            description: 令牌过期时间
                          user:
                            $ref: '#/components/schemas/User'

  /auth/logout:
    post:
      summary: 用户登出
      description: 注销当前会话
      tags: [身份认证]
      responses:
        '200':
          description: 登出成功

  /auth/refresh:
    post:
      summary: 刷新令牌
      description: 刷新JWT访问令牌
      tags: [身份认证]
      responses:
        '200':
          description: 刷新成功

  /auth/profile:
    get:
      summary: 获取当前用户信息
      description: 获取当前登录用户的详细信息
      tags: [身份认证]
      responses:
        '200':
          description: 获取成功
          content:
            application/json:
              schema:
                allOf:
                  - $ref: '#/components/schemas/ApiResponse'
                  - type: object
                    properties:
                      data:
                        $ref: '#/components/schemas/User'

  # 用户管理
  /users:
    get:
      summary: 获取用户列表
      description: 分页获取所有用户信息（仅管理员）
      tags: [用户管理]
      parameters:
        - name: page
          in: query
          schema:
            type: integer
            minimum: 1
            default: 1
        - name: per_page
          in: query
          schema:
            type: integer
            minimum: 1
            maximum: 100
            default: 20
        - name: search
          in: query
          schema:
            type: string
          description: 搜索关键词
      responses:
        '200':
          description: 获取成功
          content:
            application/json:
              schema:
                allOf:
                  - $ref: '#/components/schemas/ApiResponse'
                  - type: object
                    properties:
                      data:
                        type: object
                        properties:
                          users:
                            type: array
                            items:
                              $ref: '#/components/schemas/User'
                          meta:
                            $ref: '#/components/schemas/PaginationMeta'

    post:
      summary: 创建用户
      description: 创建新用户（仅管理员）
      tags: [用户管理]
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              properties:
                username:
                  type: string
                  minLength: 3
                  maxLength: 50
                email:
                  type: string
                  format: email
                password:
                  type: string
                  minLength: 8
                full_name:
                  type: string
                  maxLength: 100
                role:
                  type: string
                  enum: [admin, viewer]
                  default: viewer
              required:
                - username
                - email
                - password
      responses:
        '201':
          description: 创建成功

  /users/{user_id}:
    get:
      summary: 获取用户详情
      description: 获取指定用户的详细信息
      tags: [用户管理]
      parameters:
        - name: user_id
          in: path
          required: true
          schema:
            type: integer
      responses:
        '200':
          description: 获取成功

    put:
      summary: 更新用户信息
      description: 更新用户基本信息
      tags: [用户管理]
      parameters:
        - name: user_id
          in: path
          required: true
          schema:
            type: integer
      responses:
        '200':
          description: 更新成功

    delete:
      summary: 删除用户
      description: 删除指定用户
      tags: [用户管理]
      parameters:
        - name: user_id
          in: path
          required: true
          schema:
            type: integer
      responses:
        '200':
          description: 删除成功

  # 操作日志
  /audit-logs:
    get:
      summary: 获取操作日志
      description: 分页获取系统操作日志
      tags: [操作日志]
      parameters:
        - name: page
          in: query
          schema:
            type: integer
            minimum: 1
            default: 1
        - name: per_page
          in: query
          schema:
            type: integer
            minimum: 1
            maximum: 100
            default: 20
        - name: user_id
          in: query
          schema:
            type: integer
          description: 按用户筛选
        - name: action
          in: query
          schema:
            type: string
          description: 按操作类型筛选
        - name: start_time
          in: query
          schema:
            type: string
            format: date-time
          description: 开始时间
        - name: end_time
          in: query
          schema:
            type: string
            format: date-time
          description: 结束时间
      responses:
        '200':
          description: 获取成功
          content:
            application/json:
              schema:
                allOf:
                  - $ref: '#/components/schemas/ApiResponse'
                  - type: object
                    properties:
                      data:
                        type: object
                        properties:
                          logs:
                            type: array
                            items:
                              $ref: '#/components/schemas/AuditLog'
                          meta:
                            $ref: '#/components/schemas/PaginationMeta'

  # API代理到sing-box-api (节点管理等)
  /api/v1/nodes:
    get:
      summary: 获取节点列表
      description: 代理请求到sing-box-api获取节点信息
      tags: [API代理]
      parameters:
        - name: page
          in: query
          schema:
            type: integer
        - name: per_page
          in: query
          schema:
            type: integer
        - name: status
          in: query
          schema:
            type: string
      responses:
        '200':
          description: 代理成功
        '502':
          description: 后端服务不可用

    post:
      summary: 创建节点
      description: 代理请求到sing-box-api创建节点
      tags: [API代理]
      responses:
        '201':
          description: 代理成功

  /api/v1/nodes/{node_id}:
    get:
      summary: 获取节点详情
      description: 代理请求到sing-box-api获取节点详情
      tags: [API代理]
      parameters:
        - name: node_id
          in: path
          required: true
          schema:
            type: string
            format: uuid
      responses:
        '200':
          description: 代理成功

  /api/v1/config-templates:
    get:
      summary: 获取配置模板列表
      description: 代理请求到sing-box-api获取配置模板
      tags: [API代理]
      responses:
        '200':
          description: 代理成功

  # 健康检查
  /health:
    get:
      summary: 健康检查
      description: 检查Web服务器健康状态
      tags: [系统]
      security: []
      responses:
        '200':
          description: 服务正常
          content:
            application/json:
              schema:
                type: object
                properties:
                  status:
                    type: string
                    enum: [healthy]
                  timestamp:
                    type: string
                    format: date-time
                  version:
                    type: string
                  uptime:
                    type: string

tags:
  - name: 身份认证
    description: 用户登录认证相关接口
  - name: 用户管理
    description: 用户账户管理相关接口
  - name: 操作日志
    description: 系统操作审计日志接口
  - name: API代理
    description: 代理到sing-box-api的接口
  - name: 系统
    description: 系统状态和健康检查接口
```

---

## 4. Internal gRPC API (sing-box-api ↔ sing-box-web)

### 4.1 Web服务接口定义

```protobuf
syntax = "proto3";

package web.v1;

import "google/protobuf/timestamp.proto";
import "google/protobuf/empty.proto";

// Web服务接口 - 供sing-box-web调用
service WebService {
  // 节点管理
  rpc GetNodes(GetNodesRequest) returns (GetNodesResponse);
  rpc GetNode(GetNodeRequest) returns (GetNodeResponse);
  rpc CreateNode(CreateNodeRequest) returns (CreateNodeResponse);
  rpc UpdateNode(UpdateNodeRequest) returns (UpdateNodeResponse);
  rpc DeleteNode(DeleteNodeRequest) returns (google.protobuf.Empty);
  
  // 配置管理
  rpc GetConfigTemplates(GetConfigTemplatesRequest) returns (GetConfigTemplatesResponse);
  rpc GetConfigTemplate(GetConfigTemplateRequest) returns (GetConfigTemplateResponse);
  rpc CreateConfigTemplate(CreateConfigTemplateRequest) returns (CreateConfigTemplateResponse);
  rpc UpdateConfigTemplate(UpdateConfigTemplateRequest) returns (UpdateConfigTemplateResponse);
  rpc DeleteConfigTemplate(DeleteConfigTemplateRequest) returns (google.protobuf.Empty);
  rpc DeployConfig(DeployConfigRequest) returns (DeployConfigResponse);
  
  // 监控数据
  rpc GetNodeMetrics(GetNodeMetricsRequest) returns (GetNodeMetricsResponse);
  rpc GetDashboardStats(GetDashboardStatsRequest) returns (GetDashboardStatsResponse);
  
  // 部署管理
  rpc GetDeployments(GetDeploymentsRequest) returns (GetDeploymentsResponse);
  rpc GetDeployment(GetDeploymentRequest) returns (GetDeploymentResponse);
}

// 请求/响应消息定义...
message GetNodesRequest {
  int32 page = 1;
  int32 per_page = 2;
  string status_filter = 3;
  string search = 4;
  map<string, string> label_filters = 5;
}

message GetNodesResponse {
  repeated Node nodes = 1;
  PaginationMeta meta = 2;
}

message Node {
  string id = 1;
  string name = 2;
  string ip_address = 3;
  int32 port = 4;
  string status = 5;
  string agent_version = 6;
  string os_type = 7;
  string os_version = 8;
  string architecture = 9;
  int32 cpu_cores = 10;
  int64 total_memory = 11;
  int64 total_disk = 12;
  map<string, string> labels = 13;
  google.protobuf.Timestamp registered_at = 14;
  google.protobuf.Timestamp last_heartbeat_at = 15;
  google.protobuf.Timestamp created_at = 16;
  google.protobuf.Timestamp updated_at = 17;
}
```

---

## 5. Agent gRPC API (sing-box-api ↔ sing-box-agent)

### 5.1 Agent管理接口

```protobuf
syntax = "proto3";

package agent.v1;

import "google/protobuf/timestamp.proto";
import "google/protobuf/any.proto";

// Agent管理服务
service AgentService {
  // Agent连接的双向流
  rpc AgentStream(stream AgentMessage) returns (stream ManagerMessage);
  
  // Clash API代理
  rpc ProxyClashAPI(ClashAPIRequest) returns (ClashAPIResponse);
  
  // 直接控制命令
  rpc ExecuteCommand(ExecuteCommandRequest) returns (ExecuteCommandResponse);
}

// Agent消息类型
message AgentMessage {
  oneof message {
    AgentRegisterRequest register = 1;
    AgentHeartbeat heartbeat = 2;
    AgentMetrics metrics = 3;
    ConfigUpdateResponse config_response = 4;
    CommandResponse command_response = 5;
    ClashAPIResponse clash_api_response = 6;
  }
}

// Manager消息类型
message ManagerMessage {
  oneof message {
    AgentRegisterResponse register_response = 1;
    ConfigUpdateRequest config_request = 2;
    CommandRequest command_request = 3;
    ClashAPIRequest clash_api_request = 4;
  }
}

// Clash API代理请求
message ClashAPIRequest {
  string request_id = 1;
  string node_id = 2;
  string method = 3;           // GET, POST, PUT, DELETE
  string path = 4;             // API路径，如 /proxies, /configs
  map<string, string> headers = 5;
  bytes body = 6;
}

// Clash API代理响应
message ClashAPIResponse {
  string request_id = 1;
  string node_id = 2;
  int32 status_code = 3;
  map<string, string> headers = 4;
  bytes body = 5;
  string error_message = 6;
}

// 命令执行请求
message ExecuteCommandRequest {
  string request_id = 1;
  string node_id = 2;
  string command_type = 3;     // systemctl_start, systemctl_stop, systemctl_restart, get_status
  map<string, string> parameters = 4;
}

// 命令执行响应
message ExecuteCommandResponse {
  string request_id = 1;
  string node_id = 2;
  bool success = 3;
  string output = 4;
  string error_message = 5;
  int32 exit_code = 6;
}
```

---

## 6. Clash API 集成方案

### 6.1 集成架构

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│     前端页面     │    │  sing-box-web   │    │  sing-box-api   │
│                 │    │                 │    │                 │
│ Clash Dashboard │◄──►│   API Proxy     │◄──►│  gRPC Gateway   │
│                 │    │                 │    │                 │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                                                      │
                                                      ▼
                                              ┌─────────────────┐
                                              │ sing-box-agent  │
                                              │                 │
                                              │ Clash API Proxy │
                                              │       │         │
                                              │       ▼         │
                                              │   sing-box      │
                                              │  :9090/api/*    │
                                              └─────────────────┘
```

### 6.2 Clash API 功能映射

| 功能分类 | Clash API 路径 | 说明 | 集成方式 |
|---------|---------------|------|---------|
| **状态监控** | `/traffic`, `/memory`, `/connections` | 实时流量、内存、连接数 | Agent代理转发 |
| **代理管理** | `/proxies`, `/proxies/{name}` | 代理列表、选择器切换 | Agent代理转发 |
| **模式切换** | `/configs`, `/configs/switch` | Rule/Global/Direct模式 | Agent代理转发 |
| **规则管理** | `/rules` | 路由规则查询 | Agent代理转发 |
| **日志查看** | `/logs` | 实时日志流 | Agent WebSocket代理 |
| **配置重载** | `/configs/reload` | 热重载配置 | Agent代理转发 |

### 6.3 API代理实现

```go
// Agent端Clash API代理实现
type ClashAPIProxy struct {
    clashBaseURL string
    client       *http.Client
}

func (p *ClashAPIProxy) ProxyRequest(ctx context.Context, req *ClashAPIRequest) (*ClashAPIResponse, error) {
    // 构建目标URL
    targetURL := fmt.Sprintf("%s%s", p.clashBaseURL, req.Path)
    
    // 创建HTTP请求
    httpReq, err := http.NewRequestWithContext(ctx, req.Method, targetURL, bytes.NewReader(req.Body))
    if err != nil {
        return nil, err
    }
    
    // 设置请求头
    for key, value := range req.Headers {
        httpReq.Header.Set(key, value)
    }
    
    // 发送请求
    resp, err := p.client.Do(httpReq)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    // 读取响应
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, err
    }
    
    // 构建响应
    response := &ClashAPIResponse{
        RequestId:  req.RequestId,
        NodeId:     req.NodeId,
        StatusCode: int32(resp.StatusCode),
        Headers:    make(map[string]string),
        Body:       body,
    }
    
    // 复制响应头
    for key, values := range resp.Header {
        if len(values) > 0 {
            response.Headers[key] = values[0]
        }
    }
    
    return response, nil
}
```

### 6.4 常用Clash API接口

```yaml
# 代理选择器状态
GET /proxies
{
  "GLOBAL": {
    "type": "Selector",
    "now": "DIRECT",
    "all": ["DIRECT", "Proxy1", "Proxy2"]
  }
}

# 切换代理
PUT /proxies/GLOBAL
{
  "name": "Proxy1"
}

# 获取实时统计
GET /traffic
{
  "up": 1024,
  "down": 2048
}

# 获取连接信息
GET /connections
{
  "downloadTotal": 1048576,
  "uploadTotal": 524288,
  "connections": [...]
}

# 模式切换
PATCH /configs
{
  "mode": "rule"  // rule, global, direct
}
```

---

## 7. WebSocket API 定义

### 7.1 实时数据推送 (sing-box-web)

```yaml
# WebSocket连接端点
ws://localhost:3000/api/v1/ws

# 认证参数
Authorization: Bearer <jwt-token>

# 订阅消息
{
  "type": "subscribe",
  "data": {
    "topics": ["audit_logs", "node_status", "deployment_status"],
    "filters": {
      "node_ids": ["uuid1", "uuid2"]
    }
  }
}

# 推送消息格式
{
  "type": "audit_log",
  "timestamp": "2024-01-01T12:00:00Z",
  "data": {
    "user_id": 1,
    "action": "create_node",
    "resource_id": "node-uuid",
    "success": true
  }
}

{
  "type": "node_status",
  "timestamp": "2024-01-01T12:00:00Z",
  "data": {
    "node_id": "uuid1",
    "status": "online",
    "last_heartbeat": "2024-01-01T12:00:00Z"
  }
}
```

---

## 8. 错误代码定义

| 错误代码 | HTTP状态码 | 描述 | 适用服务 |
|---------|-----------|------|---------|
| `INVALID_REQUEST` | 400 | 请求参数无效 | All |
| `UNAUTHORIZED` | 401 | 未授权访问 | sing-box-web |
| `FORBIDDEN` | 403 | 禁止访问 | sing-box-web |
| `NOT_FOUND` | 404 | 资源不存在 | All |
| `CONFLICT` | 409 | 资源冲突 | All |
| `VALIDATION_ERROR` | 422 | 数据验证失败 | All |
| `INTERNAL_ERROR` | 500 | 内部服务器错误 | All |
| `SERVICE_UNAVAILABLE` | 503 | 服务不可用 | All |
| `BACKEND_ERROR` | 502 | 后端服务错误 | sing-box-web |
| `NODE_OFFLINE` | 424 | 节点离线 | sing-box-api |
| `CONFIG_INVALID` | 422 | 配置格式无效 | sing-box-api |
| `DEPLOYMENT_FAILED` | 424 | 部署失败 | sing-box-api |
| `AGENT_TIMEOUT` | 408 | Agent响应超时 | sing-box-api |
| `CLASH_API_ERROR` | 502 | Clash API调用失败 | sing-box-agent |

---

## 9. API使用示例

### 9.1 完整的用户操作流程

```bash
# 1. 用户登录 (sing-box-web)
curl -X POST http://localhost:3000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username": "admin", "password": "admin123"}'

# 2. 获取节点列表 (通过Web API代理)
curl -X GET http://localhost:3000/api/v1/api/v1/nodes \
  -H "Authorization: Bearer <token>"

# 3. 通过Clash API查看代理状态
curl -X GET http://localhost:3000/api/v1/api/v1/nodes/{node_id}/clash/proxies \
  -H "Authorization: Bearer <token>"

# 4. 切换代理 (通过Clash API)
curl -X PUT http://localhost:3000/api/v1/api/v1/nodes/{node_id}/clash/proxies/GLOBAL \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"name": "Proxy1"}'

# 5. 查看操作日志
curl -X GET http://localhost:3000/api/v1/audit-logs \
  -H "Authorization: Bearer <token>"
```

### 9.2 系统内部通信示例

```bash
# sing-box-web → sing-box-api (gRPC)
grpcurl -plaintext -d '{
  "page": 1,
  "per_page": 20,
  "status_filter": "online"
}' localhost:8080 web.v1.WebService/GetNodes

# sing-box-api → sing-box-agent (gRPC)
grpcurl -plaintext -d '{
  "request_id": "req-123",
  "node_id": "node-uuid",
  "method": "GET",
  "path": "/proxies"
}' localhost:9090 agent.v1.AgentService/ProxyClashAPI
```

这个重新设计的API规范明确了三个应用的职责分工，提供了完整的Web API和内部gRPC通信接口，并详细说明了Clash API的集成方案。 