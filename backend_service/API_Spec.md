# API 接口定义文档 - sing-box-web

## 1. 文档信息

- **API 版本**: v1.0
- **协议**: HTTPS + gRPC over TLS
- **认证方式**: JWT Bearer Token
- **内容类型**: application/json
- **字符编码**: UTF-8

---

## 2. OpenAPI 3.0 规范定义

### 2.1 基础信息

```yaml
openapi: 3.0.3
info:
  title: sing-box-web API
  description: |
    sing-box-web 分布式管理平台 REST API
    
    ## 认证
    本API使用JWT Bearer Token进行认证。在请求头中包含：
    ```
    Authorization: Bearer <your-jwt-token>
    ```
    
    ## 错误处理
    API使用标准HTTP状态码，错误响应格式统一为：
    ```json
    {
      "error": {
        "code": "ERROR_CODE",
        "message": "Human readable error message",
        "details": {}
      }
    }
    ```
  version: 1.0.0
  contact:
    name: sing-box-web API Support
    email: support@example.com
  license:
    name: MIT
    url: https://opensource.org/licenses/MIT

servers:
  - url: https://api.singbox.example.com/v1
    description: 生产环境
  - url: https://staging-api.singbox.example.com/v1  
    description: 测试环境
  - url: http://localhost:8080/v1
    description: 本地开发环境

security:
  - bearerAuth: []

components:
  securitySchemes:
    bearerAuth:
      type: http
      scheme: bearer
      bearerFormat: JWT
      description: JWT认证令牌

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
      required:
        - page
        - per_page
        - total
        - total_pages

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
      required:
        - id
        - username
        - email
        - role
        - is_active
        - created_at
        - updated_at

    # 节点模型
    Node:
      type: object
      properties:
        id:
          type: string
          format: uuid
          description: 节点唯一标识
        name:
          type: string
          maxLength: 100
          description: 节点名称
        ip_address:
          type: string
          format: ipv4
          description: 节点IP地址
        port:
          type: integer
          minimum: 1
          maximum: 65535
          default: 22
          description: SSH端口
        status:
          type: string
          enum: [online, offline, error, maintenance]
          description: 节点状态
        agent_version:
          type: string
          description: Agent版本
        os_type:
          type: string
          enum: [linux, windows, darwin]
          description: 操作系统类型
        os_version:
          type: string
          description: 操作系统版本
        architecture:
          type: string
          enum: [amd64, arm64, 386]
          description: 系统架构
        cpu_cores:
          type: integer
          minimum: 1
          description: CPU核心数
        total_memory:
          type: integer
          format: int64
          description: 总内存(字节)
        total_disk:
          type: integer
          format: int64
          description: 总磁盘(字节)
        labels:
          type: object
          additionalProperties:
            type: string
          description: 自定义标签
        metadata:
          type: object
          description: 扩展元数据
        registered_at:
          type: string
          format: date-time
          description: 注册时间
        last_heartbeat_at:
          type: string
          format: date-time
          nullable: true
          description: 最后心跳时间
        last_config_sync_at:
          type: string
          format: date-time
          nullable: true
          description: 最后配置同步时间
        created_at:
          type: string
          format: date-time
          description: 创建时间
        updated_at:
          type: string
          format: date-time
          description: 更新时间
      required:
        - id
        - name
        - ip_address
        - status
        - registered_at
        - created_at
        - updated_at

    # 系统监控指标
    SystemMetrics:
      type: object
      properties:
        timestamp:
          type: string
          format: date-time
          description: 采集时间
        cpu:
          $ref: '#/components/schemas/CPUMetrics'
        memory:
          $ref: '#/components/schemas/MemoryMetrics'
        disk:
          $ref: '#/components/schemas/DiskMetrics'
        network:
          $ref: '#/components/schemas/NetworkMetrics'
        singbox:
          $ref: '#/components/schemas/SingBoxMetrics'
      required:
        - timestamp

    CPUMetrics:
      type: object
      properties:
        usage_percent:
          type: number
          format: float
          minimum: 0
          maximum: 100
          description: CPU使用率百分比
        cores:
          type: integer
          minimum: 1
          description: CPU核心数
        load_average_1m:
          type: number
          format: float
          description: 1分钟平均负载
        load_average_5m:
          type: number
          format: float
          description: 5分钟平均负载
        load_average_15m:
          type: number
          format: float
          description: 15分钟平均负载

    MemoryMetrics:
      type: object
      properties:
        total_bytes:
          type: integer
          format: int64
          description: 总内存(字节)
        used_bytes:
          type: integer
          format: int64
          description: 已用内存(字节)
        available_bytes:
          type: integer
          format: int64
          description: 可用内存(字节)
        usage_percent:
          type: number
          format: float
          minimum: 0
          maximum: 100
          description: 内存使用率百分比
        swap_total_bytes:
          type: integer
          format: int64
          description: 交换分区总大小(字节)
        swap_used_bytes:
          type: integer
          format: int64
          description: 交换分区已用(字节)

    DiskMetrics:
      type: object
      properties:
        devices:
          type: array
          items:
            $ref: '#/components/schemas/DiskDevice'
          description: 磁盘设备列表

    DiskDevice:
      type: object
      properties:
        device:
          type: string
          description: 设备名
        mountpoint:
          type: string
          description: 挂载点
        total_bytes:
          type: integer
          format: int64
          description: 总大小(字节)
        used_bytes:
          type: integer
          format: int64
          description: 已用大小(字节)
        available_bytes:
          type: integer
          format: int64
          description: 可用大小(字节)
        usage_percent:
          type: number
          format: float
          minimum: 0
          maximum: 100
          description: 使用率百分比

    NetworkMetrics:
      type: object
      properties:
        interfaces:
          type: array
          items:
            $ref: '#/components/schemas/NetworkInterface'
          description: 网络接口列表

    NetworkInterface:
      type: object
      properties:
        name:
          type: string
          description: 接口名
        bytes_sent:
          type: integer
          format: int64
          description: 发送字节数
        bytes_recv:
          type: integer
          format: int64
          description: 接收字节数
        packets_sent:
          type: integer
          format: int64
          description: 发送包数
        packets_recv:
          type: integer
          format: int64
          description: 接收包数
        bytes_sent_per_sec:
          type: integer
          format: int64
          description: 每秒发送字节数
        bytes_recv_per_sec:
          type: integer
          format: int64
          description: 每秒接收字节数

    SingBoxMetrics:
      type: object
      properties:
        status:
          type: string
          enum: [running, stopped, error]
          description: sing-box状态
        uptime_seconds:
          type: integer
          format: int64
          description: 运行时间(秒)
        connections_count:
          type: integer
          description: 连接数

    # 配置模板
    ConfigTemplate:
      type: object
      properties:
        id:
          type: string
          format: uuid
          description: 模板ID
        name:
          type: string
          maxLength: 100
          description: 模板名称
        description:
          type: string
          description: 模板描述
        content:
          type: object
          description: sing-box配置JSON
        version:
          type: integer
          minimum: 1
          description: 版本号
        is_default:
          type: boolean
          description: 是否为默认模板
        checksum:
          type: string
          description: 配置校验和
        created_by:
          type: integer
          description: 创建者用户ID
        created_at:
          type: string
          format: date-time
          description: 创建时间
        updated_at:
          type: string
          format: date-time
          description: 更新时间
      required:
        - id
        - name
        - content
        - version
        - checksum
        - created_at
        - updated_at

    # 配置部署记录
    ConfigDeployment:
      type: object
      properties:
        id:
          type: string
          format: uuid
          description: 部署记录ID
        node_id:
          type: string
          format: uuid
          description: 节点ID
        template_id:
          type: string
          format: uuid
          description: 配置模板ID
        template_version:
          type: integer
          description: 模板版本
        status:
          type: string
          enum: [pending, success, failed, timeout]
          description: 部署状态
        config_content:
          type: object
          description: 部署的配置内容
        error_message:
          type: string
          nullable: true
          description: 错误消息
        deployed_by:
          type: integer
          description: 部署者用户ID
        deployed_at:
          type: string
          format: date-time
          description: 部署时间
        completed_at:
          type: string
          format: date-time
          nullable: true
          description: 完成时间
        created_at:
          type: string
          format: date-time
          description: 创建时间
      required:
        - id
        - node_id
        - template_id
        - template_version
        - status
        - config_content
        - deployed_by
        - deployed_at
        - created_at

paths:
  # 认证相关
  /auth/login:
    post:
      summary: 用户登录
      description: 使用用户名/邮箱和密码进行登录
      tags: [认证]
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
        '401':
          description: 认证失败
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'

  /auth/logout:
    post:
      summary: 用户登出
      description: 注销当前会话
      tags: [认证]
      responses:
        '200':
          description: 登出成功
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ApiResponse'

  /auth/refresh:
    post:
      summary: 刷新令牌
      description: 刷新JWT访问令牌
      tags: [认证]
      responses:
        '200':
          description: 刷新成功
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
                            description: 新的JWT访问令牌
                          expires_at:
                            type: string
                            format: date-time
                            description: 令牌过期时间

  # 节点管理
  /nodes:
    get:
      summary: 获取节点列表
      description: 分页获取所有节点信息，支持筛选和搜索
      tags: [节点管理]
      parameters:
        - name: page
          in: query
          schema:
            type: integer
            minimum: 1
            default: 1
          description: 页码
        - name: per_page
          in: query
          schema:
            type: integer
            minimum: 1
            maximum: 100
            default: 20
          description: 每页记录数
        - name: status
          in: query
          schema:
            type: string
            enum: [online, offline, error, maintenance]
          description: 按状态筛选
        - name: search
          in: query
          schema:
            type: string
          description: 搜索关键词(节点名称、IP地址)
        - name: labels
          in: query
          schema:
            type: string
          description: 标签筛选(格式：key1=value1,key2=value2)
        - name: sort
          in: query
          schema:
            type: string
            enum: [name, ip_address, status, created_at, last_heartbeat_at]
            default: created_at
          description: 排序字段
        - name: order
          in: query
          schema:
            type: string
            enum: [asc, desc]
            default: desc
          description: 排序方向
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
                          nodes:
                            type: array
                            items:
                              $ref: '#/components/schemas/Node'
                          meta:
                            $ref: '#/components/schemas/PaginationMeta'

    post:
      summary: 手动注册节点
      description: 手动添加新节点（用于无法自动注册的场景）
      tags: [节点管理]
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              properties:
                name:
                  type: string
                  maxLength: 100
                  description: 节点名称
                ip_address:
                  type: string
                  format: ipv4
                  description: 节点IP地址
                port:
                  type: integer
                  minimum: 1
                  maximum: 65535
                  default: 22
                  description: SSH端口
                labels:
                  type: object
                  additionalProperties:
                    type: string
                  description: 自定义标签
                metadata:
                  type: object
                  description: 扩展元数据
              required:
                - name
                - ip_address
      responses:
        '201':
          description: 创建成功
          content:
            application/json:
              schema:
                allOf:
                  - $ref: '#/components/schemas/ApiResponse'
                  - type: object
                    properties:
                      data:
                        $ref: '#/components/schemas/Node'

  /nodes/{node_id}:
    get:
      summary: 获取节点详情
      description: 获取指定节点的详细信息
      tags: [节点管理]
      parameters:
        - name: node_id
          in: path
          required: true
          schema:
            type: string
            format: uuid
          description: 节点ID
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
                        $ref: '#/components/schemas/Node'
        '404':
          description: 节点不存在
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'

    put:
      summary: 更新节点信息
      description: 更新节点的基本信息和标签
      tags: [节点管理]
      parameters:
        - name: node_id
          in: path
          required: true
          schema:
            type: string
            format: uuid
          description: 节点ID
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              properties:
                name:
                  type: string
                  maxLength: 100
                  description: 节点名称
                labels:
                  type: object
                  additionalProperties:
                    type: string
                  description: 自定义标签
                metadata:
                  type: object
                  description: 扩展元数据
      responses:
        '200':
          description: 更新成功
          content:
            application/json:
              schema:
                allOf:
                  - $ref: '#/components/schemas/ApiResponse'
                  - type: object
                    properties:
                      data:
                        $ref: '#/components/schemas/Node'

    delete:
      summary: 删除节点
      description: 删除指定节点及其相关数据
      tags: [节点管理]
      parameters:
        - name: node_id
          in: path
          required: true
          schema:
            type: string
            format: uuid
          description: 节点ID
      responses:
        '200':
          description: 删除成功
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ApiResponse'

  /nodes/{node_id}/metrics:
    get:
      summary: 获取节点监控数据
      description: 获取指定节点的历史监控数据
      tags: [监控数据]
      parameters:
        - name: node_id
          in: path
          required: true
          schema:
            type: string
            format: uuid
          description: 节点ID
        - name: start_time
          in: query
          schema:
            type: string
            format: date-time
          description: 开始时间(ISO 8601格式)
        - name: end_time
          in: query
          schema:
            type: string
            format: date-time
          description: 结束时间(ISO 8601格式)
        - name: interval
          in: query
          schema:
            type: string
            enum: [1m, 5m, 15m, 1h, 6h, 24h]
            default: 5m
          description: 数据聚合间隔
        - name: metrics
          in: query
          schema:
            type: string
          description: 指定监控指标(逗号分隔)，如：cpu,memory,disk
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
                          metrics:
                            type: array
                            items:
                              $ref: '#/components/schemas/SystemMetrics'
                          summary:
                            type: object
                            properties:
                              total_samples:
                                type: integer
                                description: 样本总数
                              start_time:
                                type: string
                                format: date-time
                                description: 实际开始时间
                              end_time:
                                type: string
                                format: date-time
                                description: 实际结束时间

  # 配置管理
  /config-templates:
    get:
      summary: 获取配置模板列表
      description: 分页获取所有配置模板
      tags: [配置管理]
      parameters:
        - name: page
          in: query
          schema:
            type: integer
            minimum: 1
            default: 1
          description: 页码
        - name: per_page
          in: query
          schema:
            type: integer
            minimum: 1
            maximum: 100
            default: 20
          description: 每页记录数
        - name: search
          in: query
          schema:
            type: string
          description: 搜索关键词(模板名称、描述)
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
                          templates:
                            type: array
                            items:
                              $ref: '#/components/schemas/ConfigTemplate'
                          meta:
                            $ref: '#/components/schemas/PaginationMeta'

    post:
      summary: 创建配置模板
      description: 创建新的配置模板
      tags: [配置管理]
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              properties:
                name:
                  type: string
                  maxLength: 100
                  description: 模板名称
                description:
                  type: string
                  description: 模板描述
                content:
                  type: object
                  description: sing-box配置JSON
                is_default:
                  type: boolean
                  default: false
                  description: 是否为默认模板
              required:
                - name
                - content
      responses:
        '201':
          description: 创建成功
          content:
            application/json:
              schema:
                allOf:
                  - $ref: '#/components/schemas/ApiResponse'
                  - type: object
                    properties:
                      data:
                        $ref: '#/components/schemas/ConfigTemplate'

  /config-templates/{template_id}:
    get:
      summary: 获取配置模板详情
      description: 获取指定配置模板的详细信息
      tags: [配置管理]
      parameters:
        - name: template_id
          in: path
          required: true
          schema:
            type: string
            format: uuid
          description: 模板ID
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
                        $ref: '#/components/schemas/ConfigTemplate'

    put:
      summary: 更新配置模板
      description: 更新配置模板（会创建新版本）
      tags: [配置管理]
      parameters:
        - name: template_id
          in: path
          required: true
          schema:
            type: string
            format: uuid
          description: 模板ID
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              properties:
                name:
                  type: string
                  maxLength: 100
                  description: 模板名称
                description:
                  type: string
                  description: 模板描述
                content:
                  type: object
                  description: sing-box配置JSON
                is_default:
                  type: boolean
                  description: 是否为默认模板
              required:
                - content
      responses:
        '200':
          description: 更新成功
          content:
            application/json:
              schema:
                allOf:
                  - $ref: '#/components/schemas/ApiResponse'
                  - type: object
                    properties:
                      data:
                        $ref: '#/components/schemas/ConfigTemplate'

    delete:
      summary: 删除配置模板
      description: 删除指定配置模板
      tags: [配置管理]
      parameters:
        - name: template_id
          in: path
          required: true
          schema:
            type: string
            format: uuid
          description: 模板ID
      responses:
        '200':
          description: 删除成功
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ApiResponse'

  /config-templates/{template_id}/deploy:
    post:
      summary: 部署配置到节点
      description: 将配置模板部署到一个或多个节点
      tags: [配置管理]
      parameters:
        - name: template_id
          in: path
          required: true
          schema:
            type: string
            format: uuid
          description: 模板ID
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              properties:
                node_ids:
                  type: array
                  items:
                    type: string
                    format: uuid
                  description: 目标节点ID列表
                template_version:
                  type: integer
                  description: 模板版本（不指定则使用最新版本）
                description:
                  type: string
                  description: 部署说明
              required:
                - node_ids
      responses:
        '202':
          description: 部署任务已提交
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
                          deployment_ids:
                            type: array
                            items:
                              type: string
                              format: uuid
                            description: 部署任务ID列表

  # 配置部署记录
  /deployments:
    get:
      summary: 获取配置部署记录
      description: 分页获取配置部署历史记录
      tags: [配置管理]
      parameters:
        - name: page
          in: query
          schema:
            type: integer
            minimum: 1
            default: 1
          description: 页码
        - name: per_page
          in: query
          schema:
            type: integer
            minimum: 1
            maximum: 100
            default: 20
          description: 每页记录数
        - name: node_id
          in: query
          schema:
            type: string
            format: uuid
          description: 按节点ID筛选
        - name: template_id
          in: query
          schema:
            type: string
            format: uuid
          description: 按模板ID筛选
        - name: status
          in: query
          schema:
            type: string
            enum: [pending, success, failed, timeout]
          description: 按状态筛选
        - name: start_time
          in: query
          schema:
            type: string
            format: date-time
          description: 部署开始时间筛选
        - name: end_time
          in: query
          schema:
            type: string
            format: date-time
          description: 部署结束时间筛选
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
                          deployments:
                            type: array
                            items:
                              $ref: '#/components/schemas/ConfigDeployment'
                          meta:
                            $ref: '#/components/schemas/PaginationMeta'

  /deployments/{deployment_id}:
    get:
      summary: 获取部署记录详情
      description: 获取指定部署记录的详细信息
      tags: [配置管理]
      parameters:
        - name: deployment_id
          in: path
          required: true
          schema:
            type: string
            format: uuid
          description: 部署记录ID
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
                        $ref: '#/components/schemas/ConfigDeployment'

  # 仪表盘统计
  /dashboard/stats:
    get:
      summary: 获取仪表盘统计数据
      description: 获取系统概览统计信息
      tags: [仪表盘]
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
                          nodes:
                            type: object
                            properties:
                              total:
                                type: integer
                                description: 节点总数
                              online:
                                type: integer
                                description: 在线节点数
                              offline:
                                type: integer
                                description: 离线节点数
                              error:
                                type: integer
                                description: 异常节点数
                          deployments:
                            type: object
                            properties:
                              today:
                                type: integer
                                description: 今日部署次数
                              success_rate:
                                type: number
                                format: float
                                description: 近7天成功率
                          templates:
                            type: object
                            properties:
                              total:
                                type: integer
                                description: 配置模板总数

  # 系统设置
  /system/settings:
    get:
      summary: 获取系统设置
      description: 获取公开的系统设置
      tags: [系统管理]
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
                        additionalProperties:
                          type: string
                        description: 系统设置键值对

tags:
  - name: 认证
    description: 用户认证相关接口
  - name: 节点管理
    description: 节点管理相关接口
  - name: 配置管理
    description: 配置模板和部署相关接口
  - name: 监控数据
    description: 监控数据相关接口
  - name: 仪表盘
    description: 仪表盘统计相关接口
  - name: 系统管理
    description: 系统管理相关接口
```

---

## 3. gRPC 服务定义

### 3.1 Manager Service (管理平台服务)

```protobuf
syntax = "proto3";

package manager.v1;

import "google/protobuf/timestamp.proto";
import "google/protobuf/any.proto";
import "google/protobuf/empty.proto";

// Manager 主服务
service ManagerService {
  // Agent连接的双向流
  rpc AgentStream(stream AgentMessage) returns (stream ManagerMessage);
  
  // 节点管理
  rpc GetNodes(GetNodesRequest) returns (GetNodesResponse);
  rpc GetNode(GetNodeRequest) returns (GetNodeResponse);
  rpc UpdateNode(UpdateNodeRequest) returns (UpdateNodeResponse);
  rpc DeleteNode(DeleteNodeRequest) returns (google.protobuf.Empty);
  
  // 配置管理
  rpc GetNodeConfig(GetNodeConfigRequest) returns (GetNodeConfigResponse);
  rpc UpdateNodeConfig(UpdateNodeConfigRequest) returns (UpdateNodeConfigResponse);
  
  // 监控数据
  rpc GetNodeMetrics(GetNodeMetricsRequest) returns (GetNodeMetricsResponse);
}

// Agent发送给Manager的消息
message AgentMessage {
  oneof message {
    AgentRegisterRequest register = 1;        // 注册请求
    AgentHeartbeat heartbeat = 2;             // 心跳
    AgentMetrics metrics = 3;                 // 监控数据
    ConfigUpdateResponse config_response = 4;  // 配置更新响应
    HealthCheckResponse health_response = 5;   // 健康检查响应
  }
}

// Manager发送给Agent的消息
message ManagerMessage {
  oneof message {
    AgentRegisterResponse register_response = 1; // 注册响应
    ConfigUpdateRequest config_request = 2;      // 配置更新请求
    HealthCheckRequest health_check = 3;         // 健康检查请求
    AgentCommand command = 4;                    // 其他命令
  }
}

// Agent注册请求
message AgentRegisterRequest {
  string node_id = 1;
  string node_name = 2;
  string ip_address = 3;
  string agent_version = 4;
  string os_type = 5;
  string os_version = 6;
  string architecture = 7;
  int32 cpu_cores = 8;
  int64 total_memory = 9;
  int64 total_disk = 10;
  map<string, string> labels = 11;
  google.protobuf.Any metadata = 12;
}

// Agent注册响应
message AgentRegisterResponse {
  bool success = 1;
  string message = 2;
  string node_id = 3;
  AgentConfig config = 4;
}

// Agent配置
message AgentConfig {
  int32 heartbeat_interval_seconds = 1;
  int32 metrics_interval_seconds = 2;
  bool enable_monitoring = 3;
  repeated string enabled_metrics = 4;
}

// Agent心跳
message AgentHeartbeat {
  string node_id = 1;
  google.protobuf.Timestamp timestamp = 2;
  NodeStatus status = 3;
  SingBoxStatus singbox_status = 4;
}

// 节点状态枚举
enum NodeStatus {
  NODE_STATUS_UNKNOWN = 0;
  NODE_STATUS_ONLINE = 1;
  NODE_STATUS_OFFLINE = 2;
  NODE_STATUS_ERROR = 3;
  NODE_STATUS_MAINTENANCE = 4;
}

// sing-box状态
message SingBoxStatus {
  string status = 1;              // running, stopped, error
  int64 uptime_seconds = 2;
  int32 connections_count = 3;
  string version = 4;
  google.protobuf.Timestamp last_restart = 5;
}

// Agent监控数据
message AgentMetrics {
  string node_id = 1;
  google.protobuf.Timestamp timestamp = 2;
  SystemMetrics system = 3;
}

// 系统监控指标
message SystemMetrics {
  CPUMetrics cpu = 1;
  MemoryMetrics memory = 2;
  DiskMetrics disk = 3;
  NetworkMetrics network = 4;
}

message CPUMetrics {
  double usage_percent = 1;
  int32 cores = 2;
  double load_average_1m = 3;
  double load_average_5m = 4;
  double load_average_15m = 5;
}

message MemoryMetrics {
  int64 total_bytes = 1;
  int64 used_bytes = 2;
  int64 available_bytes = 3;
  double usage_percent = 4;
  int64 swap_total_bytes = 5;
  int64 swap_used_bytes = 6;
}

message DiskMetrics {
  repeated DiskDevice devices = 1;
}

message DiskDevice {
  string device = 1;
  string mountpoint = 2;
  int64 total_bytes = 3;
  int64 used_bytes = 4;
  int64 available_bytes = 5;
  double usage_percent = 6;
  int64 read_bytes_per_sec = 7;
  int64 write_bytes_per_sec = 8;
}

message NetworkMetrics {
  repeated NetworkInterface interfaces = 1;
}

message NetworkInterface {
  string name = 1;
  int64 bytes_sent = 2;
  int64 bytes_recv = 3;
  int64 packets_sent = 4;
  int64 packets_recv = 5;
  int64 bytes_sent_per_sec = 6;
  int64 bytes_recv_per_sec = 7;
}

// 配置更新请求
message ConfigUpdateRequest {
  string request_id = 1;
  string node_id = 2;
  google.protobuf.Any config_content = 3;
  string checksum = 4;
  bool restart_required = 5;
  string description = 6;
}

// 配置更新响应
message ConfigUpdateResponse {
  string request_id = 1;
  string node_id = 2;
  bool success = 3;
  string error_message = 4;
  google.protobuf.Timestamp applied_at = 5;
}

// 健康检查请求
message HealthCheckRequest {
  string request_id = 1;
  repeated string checks = 2;  // system, singbox, network
}

// 健康检查响应
message HealthCheckResponse {
  string request_id = 1;
  map<string, HealthStatus> results = 2;
}

message HealthStatus {
  bool healthy = 1;
  string message = 2;
  google.protobuf.Timestamp checked_at = 3;
}

// Agent命令
message AgentCommand {
  string command_id = 1;
  string command_type = 2;  // restart, update_agent, collect_logs
  google.protobuf.Any payload = 3;
}

// 其他请求/响应消息...
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
  NodeStatus status = 5;
  string agent_version = 6;
  string os_type = 7;
  string os_version = 8;
  string architecture = 9;
  int32 cpu_cores = 10;
  int64 total_memory = 11;
  int64 total_disk = 12;
  map<string, string> labels = 13;
  google.protobuf.Any metadata = 14;
  google.protobuf.Timestamp registered_at = 15;
  google.protobuf.Timestamp last_heartbeat_at = 16;
  google.protobuf.Timestamp last_config_sync_at = 17;
  google.protobuf.Timestamp created_at = 18;
  google.protobuf.Timestamp updated_at = 19;
}

message PaginationMeta {
  int32 page = 1;
  int32 per_page = 2;
  int32 total = 3;
  int32 total_pages = 4;
}

// 更多消息定义...
```

---

## 4. WebSocket API 定义

### 4.1 实时监控数据推送

```yaml
# WebSocket连接端点
ws://localhost:8080/v1/ws/monitor

# 连接参数
Authorization: Bearer <jwt-token>

# 订阅消息格式
{
  "type": "subscribe",
  "data": {
    "topics": ["nodes", "metrics", "deployments"],
    "node_ids": ["uuid1", "uuid2"],  // 可选，指定监控的节点
    "interval": "5s"                 // 可选，推送间隔
  }
}

# 推送消息格式
{
  "type": "node_status",
  "timestamp": "2024-01-01T12:00:00Z",
  "data": {
    "node_id": "uuid1",
    "status": "online",
    "metrics": {
      // SystemMetrics数据
    }
  }
}

{
  "type": "deployment_update",
  "timestamp": "2024-01-01T12:00:00Z",
  "data": {
    "deployment_id": "uuid1",
    "node_id": "uuid2",
    "status": "success",
    "message": "Configuration applied successfully"
  }
}
```

---

## 5. 错误代码定义

| 错误代码 | HTTP状态码 | 描述 |
|---------|-----------|------|
| `INVALID_REQUEST` | 400 | 请求参数无效 |
| `UNAUTHORIZED` | 401 | 未授权访问 |
| `FORBIDDEN` | 403 | 禁止访问 |
| `NOT_FOUND` | 404 | 资源不存在 |
| `CONFLICT` | 409 | 资源冲突 |
| `VALIDATION_ERROR` | 422 | 数据验证失败 |
| `INTERNAL_ERROR` | 500 | 内部服务器错误 |
| `SERVICE_UNAVAILABLE` | 503 | 服务不可用 |
| `NODE_OFFLINE` | 424 | 节点离线 |
| `CONFIG_INVALID` | 422 | 配置格式无效 |
| `DEPLOYMENT_FAILED` | 424 | 部署失败 |
| `AGENT_TIMEOUT` | 408 | Agent响应超时 |

---

## 6. API使用示例

### 6.1 完整的配置部署流程

```bash
# 1. 登录获取token
curl -X POST http://localhost:8080/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username": "admin", "password": "admin123"}'

# 2. 获取节点列表
curl -X GET http://localhost:8080/v1/nodes \
  -H "Authorization: Bearer <token>" \
  -G -d "status=online"

# 3. 创建配置模板
curl -X POST http://localhost:8080/v1/config-templates \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "测试配置",
    "description": "用于测试的sing-box配置",
    "content": {
      "log": {"level": "info"},
      "inbounds": [...],
      "outbounds": [...]
    }
  }'

# 4. 部署配置到节点
curl -X POST http://localhost:8080/v1/config-templates/{template_id}/deploy \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "node_ids": ["node-uuid-1", "node-uuid-2"],
    "description": "部署新的路由规则"
  }'

# 5. 查看部署状态
curl -X GET http://localhost:8080/v1/deployments \
  -H "Authorization: Bearer <token>" \
  -G -d "status=pending"

# 6. 获取节点监控数据
curl -X GET http://localhost:8080/v1/nodes/{node_id}/metrics \
  -H "Authorization: Bearer <token>" \
  -G -d "start_time=2024-01-01T00:00:00Z" \
  -d "end_time=2024-01-01T23:59:59Z" \
  -d "interval=1h"
```

这个API定义文档提供了完整的REST API和gRPC接口规范，涵盖了sing-box-web平台的所有核心功能，为前端开发和系统集成提供了详细的接口指导。 