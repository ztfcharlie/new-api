# New-API Docker 部署指南

## 🚀 快速部署

### 1. 环境准备

确保你的Linux服务器已安装：
- Docker
- Docker Compose
- Git

### 2. 克隆项目

```bash
git clone <your-repo-url>
cd burncloud-aiapi-limit
```

### 3. 初始化环境

```bash
# 给脚本执行权限
chmod +x init.sh

# 运行初始化脚本
./init.sh
```

### 4. 配置环境变量

复制并编辑 `.env` 文件：

```bash
cp .env.example .env
```

编辑 `.env` 文件，设置以下必要变量：

```bash
# 数据库连接
SQL_DSN=root:burncloud123456!qwf@tcp(burncloud-aiapi-mysql:3306)/new-api?parseTime=true&charset=utf8mb4&collation=utf8mb4_unicode_ci

# Redis连接
REDIS_CONN_STRING=redis://burncloud-aiapi-redis:6379

# JWT密钥（必须修改！）
CONSOLE_JWT_SECRET=your-secret-key-here

# API域名
CONSOLE_API_DOMAIN=http://your-domain.com:3000

# 生成默认token
GENERATE_DEFAULT_TOKEN=true

# 节点类型
NODE_TYPE=master

# API请求追踪（新增功能）
API_REQUEST_LOG_ENABLED=true
```

### 5. 创建Docker网络

```bash
# 创建外部网络（与docker-compose.yml中的配置对应）
docker network create nginx-network
```

### 6. 启动服务

```bash
# 构建并启动所有服务
docker-compose up -d --build

# 查看服务状态
docker-compose ps

# 查看日志
docker-compose logs -f burncloud-aiapi
```

## 📊 API请求追踪功能

部署完成后，新的API请求追踪功能会自动启用：

- ✅ **功能已启用**: `API_REQUEST_LOG_ENABLED=true`
- 📊 **数据存储**: MySQL数据库中的 `api_request_logs` 表
- 🔄 **自动清理**: 超过10,000条记录时自动删除旧数据
- 🛡️ **安全防护**: 自动截取超长数据，防止数据库错误

### 查看追踪数据

```sql
-- 连接到MySQL容器
docker exec -it burncloud-aiapi-mysql mysql -u root -p

-- 查看API请求日志
USE new-api;
SELECT * FROM api_request_logs ORDER BY created_at DESC LIMIT 10;

-- 查看特定用户的请求
SELECT * FROM api_request_logs WHERE user_id = 1 ORDER BY created_at DESC;

-- 查看有错误的请求
SELECT * FROM api_request_logs WHERE error_message IS NOT NULL;
```

## 🔧 常用操作

### 查看服务状态

```bash
# 查看所有服务
docker-compose ps

# 查看特定服务日志
docker-compose logs -f burncloud-aiapi
docker-compose logs -f burncloud-aiapi-mysql
docker-compose logs -f burncloud-aiapi-redis
```

### 重启服务

```bash
# 重启所有服务
docker-compose restart

# 重启特定服务
docker-compose restart burncloud-aiapi
```

### 停止服务

```bash
# 停止所有服务
docker-compose down

# 停止并删除数据卷（⚠️ 会删除所有数据）
docker-compose down -v
```

### 数据备份

```bash
# 备份MySQL数据
docker exec burncloud-aiapi-mysql mysqldump -u root -pburncloud123456!qwf new-api > backup.sql

# 备份Redis数据
docker exec burncloud-aiapi-redis redis-cli BGSAVE
```

## 🛠️ 故障排除

### 常见问题

1. **网络连接错误**
   ```bash
   # 确保nginx-network存在
   docker network ls | grep nginx-network

   # 如果不存在，重新创建
   docker network create nginx-network
   ```

2. **数据库连接失败**
   ```bash
   # 检查MySQL容器状态
   docker-compose logs burncloud-aiapi-mysql

   # 检查数据库是否创建成功
   docker exec -it burncloud-aiapi-mysql mysql -u root -p -e "SHOW DATABASES;"
   ```

3. **权限问题**
   ```bash
   # 重新设置权限
   chmod -R 777 logs public/static public/uploads
   ```

4. **API请求追踪不工作**
   - 检查环境变量 `API_REQUEST_LOG_ENABLED=true`
   - 检查数据库表是否存在：`SHOW TABLES LIKE 'api_request_logs';`

### 日志位置

- 应用日志: `./logs/`
- Docker日志: `docker-compose logs`
- MySQL数据: `./mysql_data/`
- Redis数据: `./redis_data/`

## 📈 性能监控

API请求追踪功能会自动：

- 🔄 **异步保存**: 不影响API响应性能
- 📊 **数据截取**: 自动截取超长数据（限制60KB）
- 🧹 **自动清理**: 超过10,000条记录时自动清理
- 🛡️ **异常处理**: 所有错误都有日志记录，不影响主程序

## 🔐 安全提醒

1. **修改默认密码**:
   - MySQL root密码: `burncloud123456!qwf`
   - JWT密钥: 必须设置强密码

2. **网络安全**:
   - 生产环境建议关闭MySQL端口映射
   - 使用强密码和SSL连接

3. **数据保护**:
   - 定期备份数据库
   - 监控API请求日志大小

---

🎉 **部署完成！**

访问 `http://your-server-ip:3000` 开始使用 New-API！