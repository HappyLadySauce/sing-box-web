# 数据库设计说明 - sing-box-web

## 1. 数据库选型

### 1.1 推荐数据库

- **主要推荐**: PostgreSQL 15+
- **轻量部署**: SQLite 3.38+
- **缓存存储**: Redis 7.0+（用于会话和监控数据缓存）

### 1.2 选择依据

| 数据库类型 | 适用场景 | 优势 | 劣势 |
|-----------|---------|------|------|
| **PostgreSQL** | 生产环境、多用户、大规模节点 | 功能完整、支持JSON、事务完整、高并发 | 部署复杂度相对较高 |
| **SQLite** | 单机部署、小规模节点（<100个） | 零配置、单文件、轻量 | 不支持并发写入、功能有限 |
| **Redis** | 监控数据缓存、会话存储 | 高性能、丰富数据结构 | 非持久化（需配置） |

---

## 2. 核心数据表设计

### 2.1 用户与认证表

#### 2.1.1 用户表 (users)

```sql
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,  -- bcrypt hash
    full_name VARCHAR(100),
    role VARCHAR(20) DEFAULT 'admin',     -- admin, viewer (未来扩展)
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_login_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_active ON users(is_active);
```

#### 2.1.2 会话表 (user_sessions)

```sql
CREATE TABLE user_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(255) NOT NULL,     -- JWT token hash
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_used_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    ip_address INET,
    user_agent TEXT
);

CREATE INDEX idx_sessions_user_id ON user_sessions(user_id);
CREATE INDEX idx_sessions_token_hash ON user_sessions(token_hash);
CREATE INDEX idx_sessions_expires_at ON user_sessions(expires_at);
```

### 2.2 节点管理表

#### 2.2.1 节点表 (nodes)

```sql
CREATE TABLE nodes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    ip_address INET NOT NULL,
    port INTEGER DEFAULT 22,
    status VARCHAR(20) DEFAULT 'offline',  -- online, offline, error, maintenance
    agent_version VARCHAR(20),
    os_type VARCHAR(20),                   -- linux, windows, darwin
    os_version VARCHAR(50),
    architecture VARCHAR(20),              -- amd64, arm64, 386
    cpu_cores INTEGER,
    total_memory BIGINT,                   -- bytes
    total_disk BIGINT,                     -- bytes
    labels JSONB DEFAULT '{}',             -- 自定义标签
    metadata JSONB DEFAULT '{}',           -- 扩展元数据
    registered_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_heartbeat_at TIMESTAMP WITH TIME ZONE,
    last_config_sync_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_nodes_status ON nodes(status);
CREATE INDEX idx_nodes_ip ON nodes(ip_address);
CREATE INDEX idx_nodes_heartbeat ON nodes(last_heartbeat_at);
CREATE INDEX idx_nodes_labels ON nodes USING GIN(labels);
```

#### 2.2.2 节点标签表 (node_labels) - 可选的关系表方式

```sql
CREATE TABLE node_labels (
    id SERIAL PRIMARY KEY,
    node_id UUID NOT NULL REFERENCES nodes(id) ON DELETE CASCADE,
    key VARCHAR(50) NOT NULL,
    value VARCHAR(200) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_node_labels_unique ON node_labels(node_id, key);
CREATE INDEX idx_node_labels_key ON node_labels(key);
CREATE INDEX idx_node_labels_value ON node_labels(value);
```

### 2.3 配置管理表

#### 2.3.1 配置模板表 (config_templates)

```sql
CREATE TABLE config_templates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    content JSONB NOT NULL,               -- sing-box 配置JSON
    version INTEGER DEFAULT 1,
    is_default BOOLEAN DEFAULT false,
    checksum VARCHAR(64) NOT NULL,        -- SHA256 of content
    created_by INTEGER REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_config_templates_name ON config_templates(name);
CREATE INDEX idx_config_templates_created_by ON config_templates(created_by);
CREATE INDEX idx_config_templates_checksum ON config_templates(checksum);
```

#### 2.3.2 配置版本历史表 (config_template_versions)

```sql
CREATE TABLE config_template_versions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    template_id UUID NOT NULL REFERENCES config_templates(id) ON DELETE CASCADE,
    version_number INTEGER NOT NULL,
    content JSONB NOT NULL,
    checksum VARCHAR(64) NOT NULL,
    change_description TEXT,
    created_by INTEGER REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_template_versions_unique ON config_template_versions(template_id, version_number);
CREATE INDEX idx_template_versions_template_id ON config_template_versions(template_id);
```

#### 2.3.3 节点配置应用记录表 (node_config_deployments)

```sql
CREATE TABLE node_config_deployments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    node_id UUID NOT NULL REFERENCES nodes(id) ON DELETE CASCADE,
    template_id UUID NOT NULL REFERENCES config_templates(id),
    template_version INTEGER NOT NULL,
    status VARCHAR(20) DEFAULT 'pending',  -- pending, success, failed, timeout
    config_content JSONB NOT NULL,
    error_message TEXT,
    deployed_by INTEGER REFERENCES users(id),
    deployed_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    completed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_deployments_node_id ON node_config_deployments(node_id);
CREATE INDEX idx_deployments_template_id ON node_config_deployments(template_id);
CREATE INDEX idx_deployments_status ON node_config_deployments(status);
CREATE INDEX idx_deployments_deployed_at ON node_config_deployments(deployed_at);
```

### 2.4 监控数据表

#### 2.4.1 节点监控数据表 (node_metrics)

```sql
CREATE TABLE node_metrics (
    id BIGSERIAL PRIMARY KEY,
    node_id UUID NOT NULL REFERENCES nodes(id) ON DELETE CASCADE,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    
    -- CPU 监控
    cpu_usage_percent DECIMAL(5,2),
    cpu_cores INTEGER,
    load_average_1m DECIMAL(5,2),
    load_average_5m DECIMAL(5,2),
    load_average_15m DECIMAL(5,2),
    
    -- 内存监控
    memory_total_bytes BIGINT,
    memory_used_bytes BIGINT,
    memory_available_bytes BIGINT,
    memory_usage_percent DECIMAL(5,2),
    swap_total_bytes BIGINT,
    swap_used_bytes BIGINT,
    
    -- 磁盘监控
    disk_total_bytes BIGINT,
    disk_used_bytes BIGINT,
    disk_available_bytes BIGINT,
    disk_usage_percent DECIMAL(5,2),
    disk_io_read_bytes_per_sec BIGINT,
    disk_io_write_bytes_per_sec BIGINT,
    
    -- 网络监控
    network_bytes_sent_per_sec BIGINT,
    network_bytes_recv_per_sec BIGINT,
    network_packets_sent_per_sec BIGINT,
    network_packets_recv_per_sec BIGINT,
    
    -- sing-box 特定监控
    singbox_status VARCHAR(20),            -- running, stopped, error
    singbox_uptime_seconds BIGINT,
    singbox_connections_count INTEGER,
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 按节点和时间分区（PostgreSQL 12+）
CREATE INDEX idx_node_metrics_node_timestamp ON node_metrics(node_id, timestamp DESC);
CREATE INDEX idx_node_metrics_timestamp ON node_metrics(timestamp);

-- 定期清理旧数据的函数
CREATE OR REPLACE FUNCTION cleanup_old_metrics(retention_days INTEGER DEFAULT 30)
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    DELETE FROM node_metrics 
    WHERE created_at < NOW() - INTERVAL '1 day' * retention_days;
    
    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;
```

#### 2.4.2 监控数据聚合表 (node_metrics_hourly) - 可选

```sql
CREATE TABLE node_metrics_hourly (
    id BIGSERIAL PRIMARY KEY,
    node_id UUID NOT NULL REFERENCES nodes(id) ON DELETE CASCADE,
    hour_timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    
    -- 聚合统计
    cpu_usage_avg DECIMAL(5,2),
    cpu_usage_max DECIMAL(5,2),
    memory_usage_avg DECIMAL(5,2),
    memory_usage_max DECIMAL(5,2),
    disk_usage_avg DECIMAL(5,2),
    disk_usage_max DECIMAL(5,2),
    network_bytes_sent_total BIGINT,
    network_bytes_recv_total BIGINT,
    
    sample_count INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_metrics_hourly_unique ON node_metrics_hourly(node_id, hour_timestamp);
```

### 2.5 系统配置表

#### 2.5.1 系统设置表 (system_settings)

```sql
CREATE TABLE system_settings (
    key VARCHAR(100) PRIMARY KEY,
    value JSONB NOT NULL,
    description TEXT,
    is_public BOOLEAN DEFAULT false,      -- 是否可通过API公开访问
    updated_by INTEGER REFERENCES users(id),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 插入默认设置
INSERT INTO system_settings (key, value, description, is_public) VALUES
('heartbeat_interval', '"30s"', 'Agent心跳间隔', true),
('heartbeat_timeout', '"90s"', 'Agent心跳超时时间', true),
('metrics_retention_days', '30', '监控数据保留天数', true),
('max_nodes_per_user', '500', '单用户最大节点数', false),
('enable_node_auto_register', 'true', '是否允许节点自动注册', false);
```

#### 2.5.2 操作审计日志表 (audit_logs)

```sql
CREATE TABLE audit_logs (
    id BIGSERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id),
    action VARCHAR(50) NOT NULL,          -- login, logout, create_node, update_config, etc.
    resource_type VARCHAR(50),            -- node, config, user
    resource_id VARCHAR(100),
    details JSONB DEFAULT '{}',
    ip_address INET,
    user_agent TEXT,
    success BOOLEAN DEFAULT true,
    error_message TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX idx_audit_logs_action ON audit_logs(action);
CREATE INDEX idx_audit_logs_created_at ON audit_logs(created_at);
CREATE INDEX idx_audit_logs_resource ON audit_logs(resource_type, resource_id);
```

---

## 3. 索引优化策略

### 3.1 查询性能优化

```sql
-- 复合索引用于常见查询
CREATE INDEX idx_nodes_status_heartbeat ON nodes(status, last_heartbeat_at DESC);
CREATE INDEX idx_deployments_node_status_time ON node_config_deployments(node_id, status, deployed_at DESC);

-- 部分索引（只索引有用的数据）
CREATE INDEX idx_nodes_online ON nodes(id) WHERE status = 'online';
CREATE INDEX idx_deployments_failed ON node_config_deployments(id) WHERE status = 'failed';
```

### 3.2 监控数据分区（PostgreSQL）

```sql
-- 按月分区监控数据表
CREATE TABLE node_metrics_y2024m01 PARTITION OF node_metrics
    FOR VALUES FROM ('2024-01-01') TO ('2024-02-01');

CREATE TABLE node_metrics_y2024m02 PARTITION OF node_metrics
    FOR VALUES FROM ('2024-02-01') TO ('2024-03-01');
```

---

## 4. 数据库连接与配置

### 4.1 连接池配置

```yaml
database:
  postgres:
    host: "localhost"
    port: 5432
    user: "singbox_user"
    password: "secure_password"
    database: "singbox_web"
    sslmode: "require"
    
    # 连接池配置
    max_open_conns: 25
    max_idle_conns: 5
    conn_max_lifetime: "1h"
    conn_max_idle_time: "30m"
    
  sqlite:
    file: "./data/singbox.db"
    pragma:
      - "journal_mode=WAL"
      - "synchronous=NORMAL"
      - "cache_size=1000"
      - "foreign_keys=true"
      - "temp_store=memory"
```

### 4.2 备份策略

```bash
# PostgreSQL 备份脚本
pg_dump -h localhost -U singbox_user -d singbox_web \
    --no-owner --no-privileges --clean --if-exists \
    | gzip > "singbox_backup_$(date +%Y%m%d_%H%M%S).sql.gz"

# SQLite 备份
sqlite3 ./data/singbox.db ".backup ./backups/singbox_$(date +%Y%m%d_%H%M%S).db"
```

---

## 5. 迁移脚本示例

### 5.1 初始化脚本

```sql
-- migrations/001_initial_schema.up.sql
BEGIN;

-- 启用必要的扩展
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_trgm";

-- 创建所有表（省略具体DDL，参考上面的表定义）

-- 创建默认管理员用户
INSERT INTO users (username, email, password_hash, full_name, role)
VALUES (
    'admin',
    'admin@example.com',
    '$2b$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/lewdBWX3Bm4t1H3Ze',  -- "admin123"
    'System Administrator',
    'admin'
);

COMMIT;
```

### 5.2 版本升级脚本

```sql
-- migrations/002_add_node_labels.up.sql
BEGIN;

-- 添加新的标签功能
ALTER TABLE nodes ADD COLUMN IF NOT EXISTS labels JSONB DEFAULT '{}';
CREATE INDEX IF NOT EXISTS idx_nodes_labels ON nodes USING GIN(labels);

-- 更新现有节点的默认标签
UPDATE nodes SET labels = '{"environment": "production"}' WHERE labels = '{}';

COMMIT;
```

---

## 6. 性能监控查询

### 6.1 关键性能指标 (KPIs) 查询

```sql
-- 活跃节点统计
SELECT 
    status,
    COUNT(*) as count,
    COUNT(*) * 100.0 / SUM(COUNT(*)) OVER() as percentage
FROM nodes 
GROUP BY status;

-- 最近24小时配置部署成功率
SELECT 
    DATE_TRUNC('hour', deployed_at) as hour,
    COUNT(*) as total_deployments,
    SUM(CASE WHEN status = 'success' THEN 1 ELSE 0 END) as successful,
    ROUND(
        SUM(CASE WHEN status = 'success' THEN 1 ELSE 0 END) * 100.0 / COUNT(*), 
        2
    ) as success_rate
FROM node_config_deployments 
WHERE deployed_at >= NOW() - INTERVAL '24 hours'
GROUP BY DATE_TRUNC('hour', deployed_at)
ORDER BY hour;

-- 节点平均响应时间（基于心跳间隔）
SELECT 
    n.name,
    n.ip_address,
    EXTRACT(EPOCH FROM (NOW() - n.last_heartbeat_at)) as seconds_since_heartbeat
FROM nodes n
WHERE n.status = 'online'
ORDER BY seconds_since_heartbeat DESC;
```

### 6.2 监控数据查询示例

```sql
-- 获取节点最近1小时的平均CPU使用率
SELECT 
    DATE_TRUNC('minute', timestamp) as minute,
    AVG(cpu_usage_percent) as avg_cpu,
    MAX(cpu_usage_percent) as max_cpu
FROM node_metrics 
WHERE node_id = $1 
    AND timestamp >= NOW() - INTERVAL '1 hour'
GROUP BY DATE_TRUNC('minute', timestamp)
ORDER BY minute;
```

这个数据库设计充分考虑了sing-box-web的核心功能需求，包括节点管理、配置分发、监控数据收集等，同时兼顾了性能优化和未来扩展的需要。 