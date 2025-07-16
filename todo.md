# sing-box-web 项目开发 Todo 清单

## 🚀 项目状态
- [x] 需求分析阶段
- [x] 架构设计阶段
- [x] 项目启动阶段
- [ ] MVP 开发阶段  
- [ ] 测试验证阶段
- [ ] 生产部署阶段

---

## 🏗️ 项目架构说明

### 应用职责分工
- **sing-box-web**: 前端服务器、用户管理、登录认证、操作日志记录、与前端通信
- **sing-box-api**: 分布式节点管理、与Agent通信、核心业务逻辑、数据存储
- **sing-box-agent**: sing-box进程管理、配置文件管理、systemctl控制、Clash API集成

### 数据流向
```
前端 ↔ sing-box-web ↔ sing-box-api ↔ sing-box-agent ↔ sing-box进程
```

---

## ✅ 已完成的设计工作

### 📋 项目规划与架构设计 
- [x] **技术选型完成** - 详见 `Tech_Stack.md`
  - [x] 核心语言选择：Go 1.21+
  - [x] Web框架选择：Gin（三个应用统一）
  - [x] RPC框架选择：gRPC（内部通信）
  - [x] 数据库选择：PostgreSQL + Redis + TimescaleDB
  - [x] 监控方案：Prometheus + Grafana + SkyWalking
  - [x] Clash API集成技术方案
- [x] **代码结构设计完成** - 详见 `Code_Structure.md`
  - [x] 三应用架构完整目录结构规划
  - [x] 模块划分和依赖关系（删除ctl应用）
  - [x] Clash API集成架构设计
  - [x] 开发规范和最佳实践
- [x] **数据库设计完成** - 详见 `DB_Schema.md`
  - [x] 应用数据职责分工（Web/API数据库分离）
  - [x] 核心数据表设计（用户认证、节点管理、监控等）
  - [x] 时序数据库设计（TimescaleDB）
  - [x] 索引优化和分区策略
  - [x] 数据安全和备份方案
- [x] **API 接口设计完成** - 详见 `API_Spec.md`
  - [x] sing-box-web REST API（用户认证、操作审计、API代理）
  - [x] sing-box-api gRPC服务（节点管理、配置管理、监控数据）
  - [x] Clash API集成设计（代理转发、状态管理）
  - [x] WebSocket实时通信接口
  - [x] OpenAPI 3.0 规范定义

---

## 📋 第一阶段：项目初始化与基础设施 (Week 1-2) ✅

### 🏗️ 项目结构搭建
- [x] 初始化 Go module (`go mod init sing-box-web`)
- [x] 创建完整的目录结构（参考 Code_Structure.md）
  - [x] `cmd/` - 应用程序入口（sing-box-web、sing-box-api、sing-box-agent）
  - [x] `internal/` - 内部应用代码
  - [x] `pkg/` - 公共库代码
  - [x] `api/` - API 定义和生成代码
  - [x] `configs/` - 配置文件
- [x] 设置 `.gitignore` 文件
- [x] 创建基础 `Makefile`
- [x] 配置开发工具
  - [x] `.golangci.yml` 配置
  - [ ] `.editorconfig` 配置
  - [ ] VSCode/GoLand 配置

### 🛠️ 开发工具安装
- [x] 安装 Go 1.21+
- [ ] 安装开发工具 (`make install-tools`)
  - [ ] `golangci-lint` - 代码检查
  - [ ] `mockgen` - Mock 生成
  - [ ] `buf` - Protocol Buffers 工具
  - [ ] `protoc-gen-go` - Go 代码生成
  - [ ] `protoc-gen-go-grpc` - gRPC 代码生成
  - [ ] `go-agent` - SkyWalking Go Agent (链路追踪)
- [ ] 验证工具安装 (`make check-skywalking`)

### 📝 基础配置
- [x] 创建配置文件结构
  - [x] `configs/api/config.yaml` - API服务器配置
  - [x] `configs/web/config.yaml` - Web服务器配置
  - [x] `configs/agent/config.yaml` - Agent配置
- [x] 实现配置管理器 (`pkg/config/`)
- [x] 设置日志系统 (`pkg/logger/`)
- [x] 创建版本信息管理 (`pkg/version/`)

### 🔧 核心应用框架
- [ ] 完成 `cmd/sing-box-web/main.go` - Web服务器入口
- [ ] 完成 `cmd/sing-box-api/main.go` - API服务器入口
- [ ] 完成 `cmd/sing-box-agent/main.go` - Agent入口

---

## 💾 第二阶段：数据库与存储 (Week 2-3)

### 🗄️ 数据库设计实现
- [ ] 设置数据库驱动 (`pkg/database/`)
  - [ ] PostgreSQL 驱动实现
  - [ ] SQLite 驱动实现（用于开发环境）
- [ ] 创建数据库迁移系统
  - [ ] 迁移框架实现 (`internal/api/repository/migration/`)
  - [ ] 初始化迁移脚本 `001_initial_schema.up.sql`
- [ ] 实现核心数据表
  - [ ] 用户表 (`users`)
  - [ ] 会话表 (`user_sessions`)
  - [ ] 节点表 (`nodes`)
  - [ ] 配置模板表 (`config_templates`)
  - [ ] 配置部署记录表 (`node_config_deployments`)
  - [ ] 监控数据表 (`node_metrics`)
  - [ ] 系统设置表 (`system_settings`)
  - [ ] 审计日志表 (`audit_logs`)

### 📊 数据访问层
- [ ] 实现仓库接口 (`internal/api/repository/interfaces.go`)
- [ ] 用户数据仓库 (`internal/api/repository/user/`)
- [ ] 节点数据仓库 (`internal/api/repository/node/`)
- [ ] 配置数据仓库 (`internal/api/repository/config/`)
- [ ] 监控数据仓库 (`internal/api/repository/metrics/`)

### 🔄 缓存系统
- [ ] Redis 客户端实现 (`pkg/cache/redis/`)
- [ ] 内存缓存实现 (`pkg/cache/memory/`)
- [ ] 缓存接口定义 (`pkg/cache/interfaces.go`)
- [ ] 会话缓存
- [ ] 监控数据缓存
- [ ] 分布式锁实现

---

## 🔐 第三阶段：认证与安全 (Week 3-4)

### 👤 用户认证系统
- [ ] JWT 认证实现 (`pkg/auth/jwt/`)
  - [ ] Token 生成和验证
  - [ ] Refresh Token 机制
- [ ] 密码安全 (`pkg/auth/bcrypt/`)
  - [ ] 密码哈希
  - [ ] 密码验证
- [ ] 中间件实现
  - [ ] JWT 认证中间件
  - [ ] 权限验证中间件
  - [ ] 请求ID中间件
  - [ ] 审计日志中间件

### 🛡️ 安全加固
- [ ] CORS 配置
- [ ] 限流中间件
- [ ] 输入验证 (`pkg/validation/`)
- [ ] 敏感数据加密 (`pkg/crypto/`)
- [ ] SQL 注入防护
- [ ] XSS 防护

---

## 🌐 第四阶段：API 接口开发 (Week 4-6)

### 📡 Protocol Buffers 定义
- [ ] 创建 Proto 文件
  - [ ] `api/proto/common/v1/common.proto`
  - [ ] `api/proto/manager/v1/service.proto`
  - [ ] `api/proto/agent/v1/service.proto`
- [ ] 生成 Go 代码
- [ ] 生成 OpenAPI 文档

### 🔌 gRPC 服务实现 (sing-box-api)
- [ ] Manager 服务实现 (`internal/api/server/grpc/`)
  - [ ] Agent 双向流连接
  - [ ] 节点管理接口
  - [ ] 配置管理接口
  - [ ] 监控数据接口
- [ ] gRPC 拦截器
  - [ ] 认证拦截器
  - [ ] 日志拦截器
  - [ ] 监控拦截器
  - [ ] 恢复拦截器

### 🌍 Web服务器 REST API (sing-box-web)
- [ ] HTTP 服务器实现 (`internal/web/server/http/`)
- [ ] 路由配置 (`internal/web/server/http/router.go`)
- [ ] 业务处理器实现
  - [ ] 用户认证处理器 (`/auth/*`)
  - [ ] 用户管理处理器 (`/users/*`)
  - [ ] 操作日志处理器 (`/audit-logs/*`)
  - [ ] API代理处理器 (代理到sing-box-api)
  - [ ] 健康检查处理器 (`/health`)

### 🔄 WebSocket 实现 (sing-box-web)
- [ ] WebSocket Hub (`internal/web/server/websocket/hub.go`)
- [ ] 客户端连接管理
- [ ] 实时数据推送
  - [ ] 节点状态更新
  - [ ] 监控数据推送
  - [ ] 部署状态推送

---

## 🧠 第五阶段：业务逻辑层 (Week 5-7)

### 🎯 Web服务层 (sing-box-web)
- [ ] 用户管理服务 (`internal/web/service/user/`)
  - [ ] 用户注册/登录/登出
  - [ ] 用户信息管理
  - [ ] 密码修改
- [ ] 认证服务 (`internal/web/service/auth/`)
  - [ ] JWT Token管理
  - [ ] 会话管理
  - [ ] 权限验证
- [ ] 审计日志服务 (`internal/web/service/audit/`)
  - [ ] 操作日志记录
  - [ ] 日志查询和导出
- [ ] API代理服务 (`internal/web/service/proxy/`)
  - [ ] 请求转发到sing-box-api
  - [ ] 响应处理和错误转换

### 🎯 API服务层 (sing-box-api)
- [ ] 节点管理服务 (`internal/api/service/node/`)
  - [ ] 节点注册
  - [ ] 节点状态更新
  - [ ] 节点查询和过滤
  - [ ] 节点删除
- [ ] 配置管理服务 (`internal/api/service/config/`)
  - [ ] 模板创建和更新
  - [ ] 配置验证
  - [ ] 配置部署
  - [ ] 部署状态跟踪
- [ ] 监控服务 (`internal/api/service/monitoring/`)
  - [ ] 监控数据收集
  - [ ] 数据聚合
  - [ ] 告警处理

### 📊 通知服务 (sing-box-api)
- [ ] 通知服务实现 (`internal/api/service/notification/`)
  - [ ] 邮件通知
  - [ ] Webhook 通知
  - [ ] 系统消息

---

## 🤖 第六阶段：Agent 开发 (Week 6-8)

### 📱 Agent 核心功能
- [ ] Agent 应用实现 (`cmd/sing-box-agent/`)
- [ ] gRPC 客户端 (`internal/agent/client/`)
  - [ ] 连接管理
  - [ ] 重试逻辑
  - [ ] 心跳机制
- [ ] 监控数据收集 (`internal/agent/monitor/`)
  - [ ] 系统监控（CPU、内存、磁盘、网络）
  - [ ] sing-box 状态监控
  - [ ] 日志监控

### ⚙️ sing-box 进程管理
- [ ] systemctl 控制器 (`internal/agent/systemctl/`)
  - [ ] sing-box 服务启动/停止/重启
  - [ ] 服务状态查询
  - [ ] 日志获取
- [ ] 配置管理器 (`internal/agent/config/`)
  - [ ] 配置文件接收和验证
  - [ ] 配置文件写入和备份
  - [ ] 配置热重载
- [ ] sing-box 控制器 (`internal/agent/singbox/`)
  - [ ] 进程健康检查
  - [ ] 性能监控
  - [ ] 错误处理

### 🌐 Clash API 集成
- [ ] Clash API 客户端 (`internal/agent/clash/`)
  - [ ] API代理实现
  - [ ] 状态查询接口
  - [ ] 代理选择器管理
  - [ ] 模式切换接口
- [ ] Web界面代理 (`internal/agent/webui/`)
  - [ ] 静态文件服务
  - [ ] API请求转发
  - [ ] 访问控制

---

## 🌐 第七阶段：Web 前端服务 (Week 7-8)

### 🖥️ Web 服务器 (sing-box-web)
- [ ] 静态资源服务 (`internal/web/assets/`)
  - [ ] 前端文件嵌入 (`embed` 包)
  - [ ] 资源压缩和缓存
- [ ] SPA 路由处理 (`internal/web/spa/`)
  - [ ] 单页应用路由
  - [ ] 历史模式支持
- [ ] API 代理中间件
  - [ ] 请求代理到sing-box-api
  - [ ] 错误处理和转换

### 🔗 前端集成
- [ ] 前端构建集成
  - [ ] 自动化构建流程
  - [ ] 版本管理
- [ ] 开发环境配置
  - [ ] 热重载支持
  - [ ] 代理配置

---

## 🧪 第八阶段：测试 (Week 8-10)

### 🔬 单元测试
- [ ] 测试工具配置 (`pkg/testing/`)
- [ ] 数据库测试工具
- [ ] Mock 生成和使用
- [ ] Web服务层测试
  - [ ] 用户管理测试
  - [ ] 认证服务测试
  - [ ] API代理测试
- [ ] API服务层测试
  - [ ] 节点管理测试
  - [ ] 配置管理测试
  - [ ] 监控服务测试
- [ ] Agent测试
  - [ ] systemctl控制测试
  - [ ] 配置管理测试
  - [ ] Clash API集成测试
- [ ] 仓库层测试
- [ ] 工具函数测试

### 🧪 集成测试
- [ ] Web API 接口测试
- [ ] gRPC 服务测试
- [ ] Agent通信测试
- [ ] 数据库集成测试
- [ ] Redis 集成测试
- [ ] Clash API集成测试

### 📊 性能测试
- [ ] 压力测试脚本
- [ ] 并发连接测试
- [ ] API 性能测试
- [ ] 数据库性能测试
- [ ] Agent响应性能测试

---

## 📦 第九阶段：部署与运维 (Week 9-11)

### 🐳 容器化
- [ ] Dockerfile 编写
  - [ ] sing-box-web Dockerfile
  - [ ] sing-box-api Dockerfile
  - [ ] sing-box-agent Dockerfile
- [ ] 多阶段构建优化
- [ ] 镜像安全配置

### ☸️ Kubernetes 部署
- [ ] K8s 部署清单
  - [ ] Web服务器 Deployment
  - [ ] API服务器 Deployment
  - [ ] Agent DaemonSet
  - [ ] Service 配置
  - [ ] ConfigMap 和 Secret
  - [ ] Ingress 配置
- [ ] Helm Chart（可选）

### 📈 监控与可观测性
- [ ] Prometheus 指标实现
  - [ ] Web服务器指标
  - [ ] API服务器指标
  - [ ] Agent指标
  - [ ] 业务指标
- [ ] Grafana 仪表盘
  - [ ] 系统监控面板
  - [ ] 业务监控面板
  - [ ] Agent状态面板
- [ ] 日志收集配置
- [ ] 告警规则配置

### 🔄 CI/CD 流水线
- [ ] GitHub Actions 配置
  - [ ] 代码检查流水线
  - [ ] 测试流水线
  - [ ] 构建流水线
  - [ ] 部署流水线
- [ ] 自动化测试集成
- [ ] 容器镜像推送

---

## 📚 第十阶段：文档与完善 (Week 10-12)

### 📖 文档编写
- [ ] README.md 完善
- [ ] API 文档生成
- [ ] 部署文档
- [ ] 开发者文档
- [ ] 用户手册
- [ ] Clash API集成说明

### 🛠️ 工具与脚本
- [ ] 数据库迁移脚本
- [ ] 部署脚本
- [ ] 备份恢复脚本
- [ ] 监控脚本
- [ ] Agent安装脚本

### 🔧 优化与完善
- [ ] 性能优化
- [ ] 安全审计
- [ ] 代码重构
- [ ] 错误处理完善
- [ ] Clash API功能增强

---

## 🎯 额外任务（根据需要）

### 🚀 扩展功能
- [ ] 多租户支持
- [ ] 配置模板市场
- [ ] 自动化测试套件
- [ ] 国际化支持
- [ ] Clash API Dashboard集成

### 🔒 安全增强
- [ ] OAuth2 集成
- [ ] RBAC 权限系统
- [ ] API 密钥管理
- [ ] 安全扫描集成

### 📊 高级监控
- [ ] 分布式链路追踪
- [ ] 高级告警规则
- [ ] 性能分析工具
- [ ] 自动化故障恢复

---

## ✅ 里程碑检查点

- [ ] **Week 4**: 基础架构完成，三个应用框架搭建完毕
- [ ] **Week 6**: 核心功能开发完成，基本 API 可用
- [ ] **Week 8**: Agent 开发完成，端到端功能打通，Clash API集成完成
- [ ] **Week 10**: 测试完成，产品基本可用
- [ ] **Week 12**: 部署配置完成，正式发布

---

## 📝 注意事项

1. **架构清晰**: 三个应用职责分明，避免功能重叠
2. **测试驱动**: 每个功能开发完成后立即编写测试
3. **文档同步**: 代码开发的同时更新相关文档
4. **安全优先**: 在开发过程中始终考虑安全问题
5. **性能考虑**: 关键路径需要进行性能测试和优化
6. **Clash API**: 重点关注Agent中Clash API的稳定性和性能
7. **链路追踪**: 使用SkyWalking Go Agent自动instrument，注意不支持的场景需手动添加span

## 🔍 SkyWalking 集成说明

### 自动支持的功能
- ✅ Gin HTTP框架 (sing-box-web)
- ✅ gRPC服务端和客户端 (sing-box-api ↔ sing-box-agent)
- ✅ GORM数据库操作 (PostgreSQL)
- ✅ Redis操作 (go-redis)
- ✅ 标准库http.Client

### 需要手动处理的场景
- ❌ WebSocket连接 (需要手动span)
- ❌ Clash API代理 (需要手动span)
- ❌ 文件操作
- ❌ 第三方SDK调用

### 构建和部署
- 使用 `make build` 构建带SkyWalking的版本
- 使用 `make build-no-tracing` 构建不带追踪的版本
- 环境变量配置详见 `Tech_Stack.md`

---

*最后更新时间: 2024-01-XX*
