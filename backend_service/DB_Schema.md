# 数据库设计说明 - sing-box-web

## 1. 数据库技术选型

### 1.1 主数据库：PostgreSQL 15+

**选择理由**：
- 完善的ACID支持，确保数据一致性
- 强大的JSON支持，适合存储动态配置
- 优秀的并发性能和扩展性
- 丰富的数据类型和索引支持
- 成熟的生态系统和运维工具

### 1.2 缓存数据库：Redis 7+

**使用场景**：
- 用户会话存储
- 接口响应缓存
- 分布式锁
- 消息队列
- 实时监控数据缓存

### 1.3 时序数据库：PostgreSQL + TimescaleDB

**使用场景**：
- 节点监控指标存储
- 流量统计数据
- 系统性能数据
- Clash API监控数据

---

## 2. 系统架构与数据分布

### 2.1 应用数据职责分工

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   sing-box-web  │    │  sing-box-api   │    │ sing-box-agent  │
│                 │    │                 │    │                 │
│ • 用户数据       │    │ • 节点数据       │    │ • 本地配置       │
│ • 会话管理       │    │ • 配置模板       │    │ • 临时状态       │
│ • 操作日志       │    │ • 部署记录       │    │ • 监控缓存       │
│ • 权限控制       │    │ • 监控数据       │    │                 │
│                 │    │ • 统计分析       │    │                 │
└─────────────────┘    └─────────────────┘    └─────────────────┘
        │                         │                         │
        ▼                         ▼                         ▼
   PostgreSQL              PostgreSQL                   本地文件
   用户数据库              业务数据库                   + 内存缓存
```

### 2.2 数据库分离策略

| 数据库 | 主要应用 | 存储内容 | 备份策略 |
|--------|----------|----------|----------|
| **sing_box_web** | sing-box-web | 用户、认证、审计 | 每日全备 + 实时日志 |
| **sing_box_api** | sing-box-api | 节点、配置、监控 | 每日全备 + 实时日志 |
| **sing_box_cache** | Redis | 会话、缓存、锁 | 定期快照 |

---

## 3. 核心数据表设计

### 3.1 用户认证模块 (sing_box_web 数据库)

#### 3.1.1 用户表 (users)

```sql
CREATE TABLE users (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username        VARCHAR(50) UNIQUE NOT NULL,
    email           VARCHAR(255) UNIQUE NOT NULL,
    password_hash   VARCHAR(255) NOT NULL,
    salt           VARCHAR(32) NOT NULL,
    full_name      VARCHAR(100),
    avatar_url     VARCHAR(500),
    status         INTEGER NOT NULL DEFAULT 1, -- 1: 启用, 0: 禁用, -1: 删除
    role_id        UUID NOT NULL REFERENCES roles(id),
    last_login_at  TIMESTAMP WITH TIME ZONE,
    last_login_ip  INET,
    login_attempts INTEGER DEFAULT 0,
    locked_until   TIMESTAMP WITH TIME ZONE,
    created_at     TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at     TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT chk_username_length CHECK (char_length(username) >= 3),
    CONSTRAINT chk_email_format CHECK (email ~* '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$')
);

-- 索引
CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_status ON users(status);
CREATE INDEX idx_users_role_id ON users(role_id);
CREATE INDEX idx_users_last_login_at ON users(last_login_at);
```

#### 3.1.2 角色表 (roles)

```sql
CREATE TABLE roles (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        VARCHAR(50) UNIQUE NOT NULL,
    display_name VARCHAR(100) NOT NULL,
    description TEXT,
    permissions JSONB NOT NULL DEFAULT '[]',
    is_system   BOOLEAN NOT NULL DEFAULT FALSE,
    created_at  TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 索引
CREATE INDEX idx_roles_name ON roles(name);
CREATE INDEX idx_roles_is_system ON roles(is_system);
CREATE INDEX gin_roles_permissions ON roles USING GIN(permissions);
```

#### 3.1.3 用户会话表 (user_sessions)

```sql
CREATE TABLE user_sessions (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(255) NOT NULL UNIQUE,
    device_info JSONB,
    ip_address INET NOT NULL,
    user_agent TEXT,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    is_active  BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    last_used_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 索引
CREATE INDEX idx_sessions_user_id ON user_sessions(user_id);
CREATE INDEX idx_sessions_token_hash ON user_sessions(token_hash);
CREATE INDEX idx_sessions_expires_at ON user_sessions(expires_at);
CREATE INDEX idx_sessions_is_active ON user_sessions(is_active);
```

#### 3.1.4 操作审计表 (audit_logs)

```sql
CREATE TABLE audit_logs (
    id          BIGSERIAL PRIMARY KEY,
    user_id     UUID REFERENCES users(id),
    username    VARCHAR(50),
    action      VARCHAR(100) NOT NULL,
    resource_type VARCHAR(50) NOT NULL,
    resource_id VARCHAR(100),
    details     JSONB,
    ip_address  INET,
    user_agent  TEXT,
    status      INTEGER NOT NULL, -- 200: 成功, 4xx: 客户端错误, 5xx: 服务器错误
    error_message TEXT,
    created_at  TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 索引
CREATE INDEX idx_audit_user_id ON audit_logs(user_id);
CREATE INDEX idx_audit_action ON audit_logs(action);
CREATE INDEX idx_audit_resource_type ON audit_logs(resource_type);
CREATE INDEX idx_audit_resource_id ON audit_logs(resource_id);
CREATE INDEX idx_audit_created_at ON audit_logs(created_at);
CREATE INDEX idx_audit_status ON audit_logs(status);
CREATE INDEX gin_audit_details ON audit_logs USING GIN(details);

-- 分区表（按月分区）
CREATE TABLE audit_logs_y2024m01 PARTITION OF audit_logs
    FOR VALUES FROM ('2024-01-01') TO ('2024-02-01');
```

### 3.2 节点管理模块 (sing_box_api 数据库)

#### 3.2.1 节点表 (nodes)

```sql
CREATE TABLE nodes (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        VARCHAR(100) NOT NULL,
    host        VARCHAR(255) NOT NULL,
    port        INTEGER NOT NULL DEFAULT 9090,
    region      VARCHAR(50),
    provider    VARCHAR(50),
    tags        JSONB DEFAULT '[]',
    metadata    JSONB DEFAULT '{}',
    
    -- 连接配置
    auth_token  VARCHAR(255),
    tls_config  JSONB,
    
    -- 状态信息
    status      INTEGER NOT NULL DEFAULT 0, -- 0: 未连接, 1: 在线, 2: 离线, 3: 错误
    version     VARCHAR(50),
    last_heartbeat TIMESTAMP WITH TIME ZONE,
    
    -- Clash API配置
    clash_api_enabled BOOLEAN DEFAULT FALSE,
    clash_api_port    INTEGER,
    clash_api_secret  VARCHAR(255),
    
    -- 统计信息
    cpu_usage   DECIMAL(5,2),
    memory_usage DECIMAL(5,2),
    disk_usage  DECIMAL(5,2),
    network_in  BIGINT DEFAULT 0,
    network_out BIGINT DEFAULT 0,
    
    created_at  TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT chk_port_range CHECK (port > 0 AND port <= 65535),
    CONSTRAINT chk_usage_range CHECK (
        cpu_usage >= 0 AND cpu_usage <= 100 AND
        memory_usage >= 0 AND memory_usage <= 100 AND
        disk_usage >= 0 AND disk_usage <= 100
    )
);

-- 索引
CREATE UNIQUE INDEX idx_nodes_host_port ON nodes(host, port);
CREATE INDEX idx_nodes_name ON nodes(name);
CREATE INDEX idx_nodes_status ON nodes(status);
CREATE INDEX idx_nodes_region ON nodes(region);
CREATE INDEX idx_nodes_provider ON nodes(provider);
CREATE INDEX idx_nodes_last_heartbeat ON nodes(last_heartbeat);
CREATE INDEX gin_nodes_tags ON nodes USING GIN(tags);
CREATE INDEX gin_nodes_metadata ON nodes USING GIN(metadata);
```

#### 3.2.2 配置模板表 (config_templates)

```sql
CREATE TABLE config_templates (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        VARCHAR(100) NOT NULL,
    description TEXT,
    version     VARCHAR(20) NOT NULL DEFAULT '1.0.0',
    category    VARCHAR(50) NOT NULL,
    tags        JSONB DEFAULT '[]',
    
    -- 模板内容
    config_schema JSONB NOT NULL,
    default_values JSONB DEFAULT '{}',
    variables   JSONB DEFAULT '[]',
    
    -- 验证规则
    validation_rules JSONB DEFAULT '{}',
    
    -- 状态
    is_active   BOOLEAN NOT NULL DEFAULT TRUE,
    is_default  BOOLEAN NOT NULL DEFAULT FALSE,
    
    -- 作者信息
    created_by  UUID,
    created_at  TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 索引
CREATE INDEX idx_templates_name ON config_templates(name);
CREATE INDEX idx_templates_category ON config_templates(category);
CREATE INDEX idx_templates_is_active ON config_templates(is_active);
CREATE INDEX idx_templates_is_default ON config_templates(is_default);
CREATE INDEX gin_templates_tags ON config_templates USING GIN(tags);
```

#### 3.2.3 节点配置表 (node_configs)

```sql
CREATE TABLE node_configs (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    node_id     UUID NOT NULL REFERENCES nodes(id) ON DELETE CASCADE,
    template_id UUID REFERENCES config_templates(id),
    name        VARCHAR(100) NOT NULL,
    description TEXT,
    
    -- 配置内容
    config_data JSONB NOT NULL,
    variables   JSONB DEFAULT '{}',
    
    -- 部署状态
    status      INTEGER NOT NULL DEFAULT 0, -- 0: 草稿, 1: 已部署, 2: 部署中, 3: 部署失败
    checksum    VARCHAR(64),
    
    -- 部署信息
    deployed_at TIMESTAMP WITH TIME ZONE,
    deployed_by UUID,
    
    created_at  TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT uq_node_config_name UNIQUE(node_id, name)
);

-- 索引
CREATE INDEX idx_node_configs_node_id ON node_configs(node_id);
CREATE INDEX idx_node_configs_template_id ON node_configs(template_id);
CREATE INDEX idx_node_configs_status ON node_configs(status);
CREATE INDEX idx_node_configs_checksum ON node_configs(checksum);
```

#### 3.2.4 部署历史表 (deployment_history)

```sql
CREATE TABLE deployment_history (
    id          BIGSERIAL PRIMARY KEY,
    node_id     UUID NOT NULL REFERENCES nodes(id),
    config_id   UUID NOT NULL REFERENCES node_configs(id),
    action      VARCHAR(20) NOT NULL, -- deploy, update, rollback, delete
    
    -- 部署前状态
    old_config  JSONB,
    old_checksum VARCHAR(64),
    
    -- 部署后状态
    new_config  JSONB,
    new_checksum VARCHAR(64),
    
    -- 部署结果
    status      INTEGER NOT NULL, -- 1: 成功, 2: 进行中, 3: 失败
    error_message TEXT,
    
    -- 操作信息
    deployed_by UUID,
    started_at  TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP WITH TIME ZONE,
    
    -- 元数据
    metadata    JSONB DEFAULT '{}'
);

-- 索引
CREATE INDEX idx_deployment_node_id ON deployment_history(node_id);
CREATE INDEX idx_deployment_config_id ON deployment_history(config_id);
CREATE INDEX idx_deployment_action ON deployment_history(action);
CREATE INDEX idx_deployment_status ON deployment_history(status);
CREATE INDEX idx_deployment_started_at ON deployment_history(started_at);
```

### 3.3 监控数据模块 (TimescaleDB)

#### 3.3.1 节点监控指标表 (node_metrics)

```sql
-- 创建超表
CREATE TABLE node_metrics (
    time        TIMESTAMPTZ NOT NULL,
    node_id     UUID NOT NULL,
    metric_type VARCHAR(50) NOT NULL,
    
    -- 数值指标
    value       DOUBLE PRECISION,
    
    -- 标签和元数据
    labels      JSONB,
    metadata    JSONB
);

-- 转换为TimescaleDB超表
SELECT create_hypertable('node_metrics', 'time');

-- 索引
CREATE INDEX idx_node_metrics_node_id_time ON node_metrics (node_id, time DESC);
CREATE INDEX idx_node_metrics_type_time ON node_metrics (metric_type, time DESC);
CREATE INDEX gin_node_metrics_labels ON node_metrics USING GIN(labels);

-- 数据保留策略
SELECT add_retention_policy('node_metrics', INTERVAL '90 days');
```

#### 3.3.2 Clash API监控表 (clash_metrics)

```sql
CREATE TABLE clash_metrics (
    time        TIMESTAMPTZ NOT NULL,
    node_id     UUID NOT NULL,
    
    -- 连接统计
    total_connections INTEGER,
    active_connections INTEGER,
    
    -- 流量统计
    upload_total   BIGINT,
    download_total BIGINT,
    upload_speed   BIGINT,
    download_speed BIGINT,
    
    -- 代理状态
    proxy_mode     VARCHAR(20),
    current_proxy  VARCHAR(100),
    
    -- 延迟测试
    proxy_delays   JSONB,
    
    -- 元数据
    metadata       JSONB
);

-- 转换为TimescaleDB超表
SELECT create_hypertable('clash_metrics', 'time');

-- 索引
CREATE INDEX idx_clash_metrics_node_id_time ON clash_metrics (node_id, time DESC);
CREATE INDEX idx_clash_metrics_proxy_mode ON clash_metrics (proxy_mode);

-- 数据保留策略
SELECT add_retention_policy('clash_metrics', INTERVAL '30 days');
```

---

## 4. 数据库连接配置

### 4.1 连接池配置

```yaml
# PostgreSQL连接配置
postgres:
  # Web数据库
  web_db:
    host: localhost
    port: 5432
    database: sing_box_web
    username: sing_box_web_user
    password: "${POSTGRES_WEB_PASSWORD}"
    ssl_mode: require
    pool:
      max_open: 25
      max_idle: 10
      max_lifetime: 300s
      max_idle_time: 60s
  
  # API数据库
  api_db:
    host: localhost
    port: 5432
    database: sing_box_api
    username: sing_box_api_user
    password: "${POSTGRES_API_PASSWORD}"
    ssl_mode: require
    pool:
      max_open: 50
      max_idle: 20
      max_lifetime: 300s
      max_idle_time: 60s

# Redis连接配置
redis:
  host: localhost
  port: 6379
  password: "${REDIS_PASSWORD}"
  database: 0
  pool:
    max_active: 100
    max_idle: 50
    idle_timeout: 300s
    wait: true
```

### 4.2 数据库迁移策略

```sql
-- 创建迁移版本表
CREATE TABLE schema_migrations (
    version     BIGINT PRIMARY KEY,
    dirty       BOOLEAN NOT NULL DEFAULT FALSE,
    applied_at  TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 创建迁移锁表
CREATE TABLE schema_migration_lock (
    lock_id     INTEGER PRIMARY KEY DEFAULT 1,
    locked      BOOLEAN NOT NULL DEFAULT FALSE,
    locked_at   TIMESTAMP WITH TIME ZONE,
    locked_by   VARCHAR(255),
    
    CONSTRAINT single_lock CHECK (lock_id = 1)
);
```

---

## 5. 性能优化策略

### 5.1 索引优化

1. **复合索引设计**
   ```sql
   -- 用户查询优化
   CREATE INDEX idx_users_status_role_created ON users(status, role_id, created_at);
   
   -- 审计日志查询优化
   CREATE INDEX idx_audit_user_action_time ON audit_logs(user_id, action, created_at);
   
   -- 节点状态查询优化
   CREATE INDEX idx_nodes_status_region_heartbeat ON nodes(status, region, last_heartbeat);
   ```

2. **分区表优化**
   ```sql
   -- 审计日志按月分区
   CREATE TABLE audit_logs_default PARTITION OF audit_logs DEFAULT;
   
   -- 监控数据按时间分区（TimescaleDB自动处理）
   ```

### 5.2 查询优化

1. **预编译语句**
2. **连接池优化**
3. **读写分离**
4. **缓存策略**

### 5.3 数据清理策略

```sql
-- 清理过期会话
DELETE FROM user_sessions 
WHERE expires_at < CURRENT_TIMESTAMP - INTERVAL '7 days';

-- 清理旧审计日志（超过1年）
DELETE FROM audit_logs 
WHERE created_at < CURRENT_TIMESTAMP - INTERVAL '1 year';

-- 清理监控数据（通过TimescaleDB自动清理）
```

---

## 6. 数据安全与备份

### 6.1 敏感数据加密

```sql
-- 密码哈希存储
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- 敏感配置加密
CREATE OR REPLACE FUNCTION encrypt_sensitive_data(data TEXT, key TEXT)
RETURNS TEXT AS $$
BEGIN
    RETURN encode(encrypt(data::bytea, key::bytea, 'aes'), 'base64');
END;
$$ LANGUAGE plpgsql;
```

### 6.2 备份策略

1. **全量备份**：每日执行
2. **增量备份**：每小时执行
3. **WAL日志备份**：实时同步
4. **跨地域备份**：每日同步

### 6.3 权限控制

```sql
-- 创建专用用户
CREATE ROLE sing_box_web_user WITH LOGIN PASSWORD 'strong_password';
CREATE ROLE sing_box_api_user WITH LOGIN PASSWORD 'strong_password';

-- 授予最小权限
GRANT CONNECT ON DATABASE sing_box_web TO sing_box_web_user;
GRANT USAGE ON SCHEMA public TO sing_box_web_user;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO sing_box_web_user;

GRANT CONNECT ON DATABASE sing_box_api TO sing_box_api_user;
GRANT USAGE ON SCHEMA public TO sing_box_api_user;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO sing_box_api_user;
```

---

## 7. 总结

这个数据库设计方案为sing-box-web项目提供了：

### 7.1 清晰的数据分层
- **Web层**：用户认证、会话管理、操作审计
- **API层**：节点管理、配置模板、部署历史
- **监控层**：性能指标、Clash API数据

### 7.2 高性能架构
- PostgreSQL主数据库确保ACID特性
- Redis缓存提升响应速度
- TimescaleDB处理时序数据
- 合理的索引和分区策略

### 7.3 可扩展设计
- 支持水平分片
- 灵活的JSON字段存储
- 完善的迁移机制
- 模块化的数据结构

### 7.4 安全保障
- 敏感数据加密存储
- 细粒度权限控制
- 完整的审计跟踪
- 可靠的备份恢复

这个设计充分考虑了三个应用的职责分工，为构建高性能、高可用的sing-box管理平台提供了坚实的数据基础。 