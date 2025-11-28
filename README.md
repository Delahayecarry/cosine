# Cosine2API

一个将 [Cosine](https://cosine.sh) API 转换为 OpenAI 兼容格式的代理服务，集成 LinuxDo OAuth 登录认证。

## 功能特性

- **OpenAI API 兼容**: 将 Cosine API 转换为标准的 OpenAI API 格式，可无缝集成到现有的 OpenAI 客户端中
- **流式响应支持**: 完整支持 Server-Sent Events (SSE) 流式响应
- **LinuxDo OAuth 认证**: 集成 LinuxDo 社区 OAuth 登录
- **JWT 令牌认证**: 安全的 JWT 令牌管理
- **PostgreSQL 数据库**: 持久化存储用户数据
- **Docker 支持**: 提供完整的 Docker 容器化部署方案
- **健康检查**: 内置服务健康检查端点

## 技术栈

- **后端框架**: Go 1.24 + Gin
- **数据库**: PostgreSQL 16
- **认证**: OAuth2 + JWT
- **容器化**: Docker + Docker Compose

## 快速开始

### 前置要求

- Go 1.24+
- PostgreSQL 16+
- Docker & Docker Compose (可选)

### 使用 Docker Compose 部署（推荐）

1. **克隆仓库**

```bash
git clone <repository-url>
cd cosine
```

2. **配置环境变量**

```bash
# 复制环境变量模板
cp .env.example .env

# 编辑 .env 文件，设置数据库密码等
nano .env
```

3. **配置应用**

```bash
# 复制配置文件模板
cp config.yaml.example config.yaml

# 编辑配置文件
nano config.yaml
```

在 [config.yaml](config.yaml) 中配置以下关键信息：

```yaml
server:
  port: 7643

database:
  host: db  # Docker Compose 中使用 'db'
  port: 5432
  user: cosine
  password: cosine123
  dbname: cosine2api
  sslmode: disable

upstream:
  base_url: https://api.cosine.sh

linuxdo:
  client_id: your_client_id           # 在 LinuxDo 获取
  client_secret: your_client_secret   # 在 LinuxDo 获取
  backend_base_url: "http://your-domain:7643"

jwt:
  secret: "your_jwt_secret_key_here_change_me"  # 请修改为强密码
```

4. **启动服务**

```bash
docker compose up -d
```

5. **检查服务状态**

```bash
# 查看服务日志
docker compose logs -f

# 检查健康状态
curl http://localhost:7643/health
```

### 本地开发部署

1. **安装依赖**

```bash
go mod download
```

2. **配置数据库**

创建 PostgreSQL 数据库并执行初始化脚本：

```bash
psql -U postgres -f init.sql
```

3. **配置文件**

参考上述 Docker 部署步骤配置 [config.yaml](config.yaml)，注意将 `database.host` 改为实际的数据库地址。

4. **运行服务**

```bash
go run main.go
```

## API 使用

### OpenAI 兼容端点

服务提供以下 OpenAI 兼容的 API 端点：

#### 1. 获取模型列表

```bash
GET /v1/models
```

示例：

```bash
curl http://localhost:7643/v1/models
```

#### 2. 聊天补全（流式）

```bash
POST /v1/chat/completions
Content-Type: application/json
Authorization: Bearer YOUR_COSINE_AUTH_COOKIE

{
  "model": "gpt-4",
  "messages": [
    {
      "role": "user",
      "content": "Hello, how are you?"
    }
  ],
  "stream": true
}
```

示例：

```bash
curl -X POST http://localhost:7643/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_COSINE_AUTH_COOKIE" \
  -d '{
    "model": "gpt-4",
    "messages": [{"role": "user", "content": "Hello"}],
    "stream": true
  }'
```

#### 3. 聊天补全（非流式）

设置 `"stream": false` 即可获取完整响应：

```bash
curl -X POST http://localhost:7643/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_COSINE_AUTH_COOKIE" \
  -d '{
    "model": "gpt-4",
    "messages": [{"role": "user", "content": "Hello"}],
    "stream": false
  }'
```

### 认证端点

#### 1. 获取 LinuxDo OAuth 授权 URL

```bash
GET /api/auth/linuxdo/url
```

返回示例：

```json
{
  "url": "https://connect.linux.do/oauth2/authorize?client_id=..."
}
```

#### 2. OAuth 回调处理

```bash
GET /api/auth/linuxdo/callback?code=xxx&state=xxx
```

成功后返回包含 JWT 令牌的响应。

### 在 OpenAI 客户端中使用

你可以在任何支持自定义 API 端点的 OpenAI 客户端中使用本服务：

**Python 示例**:

```python
from openai import OpenAI

client = OpenAI(
    base_url="http://localhost:7643/v1",
    api_key="YOUR_COSINE_AUTH_COOKIE"  # 使用 Cosine auth cookie
)

response = client.chat.completions.create(
    model="gpt-4",
    messages=[
        {"role": "user", "content": "Hello!"}
    ],
    stream=True
)

for chunk in response:
    if chunk.choices[0].delta.content:
        print(chunk.choices[0].delta.content, end="")
```

**Node.js 示例**:

```javascript
import OpenAI from 'openai';

const client = new OpenAI({
  baseURL: 'http://localhost:7643/v1',
  apiKey: 'YOUR_COSINE_AUTH_COOKIE'
});

const stream = await client.chat.completions.create({
  model: 'gpt-4',
  messages: [{ role: 'user', content: 'Hello!' }],
  stream: true,
});

for await (const chunk of stream) {
  process.stdout.write(chunk.choices[0]?.delta?.content || '');
}
```

## 项目结构

```
.
├── auth/                  # 认证相关逻辑
│   ├── jwt.go            # JWT 令牌处理
│   ├── linuxdo.go        # LinuxDo OAuth
│   └── middleware.go     # 认证中间件
├── config/               # 配置管理
│   └── config.go
├── database/             # 数据库操作
│   ├── postgres.go       # 数据库初始化
│   └── linuxdo_user.go   # 用户数据操作
├── handlers/             # HTTP 处理器
│   ├── auth.go          # 认证相关路由
│   ├── chat.go          # 聊天接口
│   ├── health.go        # 健康检查
│   └── models.go        # 模型列表
├── models/              # 数据模型
│   └── types.go
├── upstream/            # 上游 API 客户端
│   └── cosine.go        # Cosine API 客户端
├── docker-compose.yml   # Docker Compose 配置
├── Dockerfile           # Docker 镜像构建
├── init.sql            # 数据库初始化脚本
├── config.yaml.example  # 配置文件模板
└── main.go             # 程序入口
```

## 配置说明

### 配置文件参数

| 参数 | 说明 | 示例 |
|------|------|------|
| `server.port` | 服务监听端口 | `7643` |
| `database.host` | 数据库主机 | `db` 或 `localhost` |
| `database.port` | 数据库端口 | `5432` |
| `database.user` | 数据库用户名 | `cosine` |
| `database.password` | 数据库密码 | `cosine123` |
| `database.dbname` | 数据库名称 | `cosine2api` |
| `upstream.base_url` | Cosine API 地址 | `https://api.cosine.sh` |
| `linuxdo.client_id` | LinuxDo OAuth 客户端 ID | - |
| `linuxdo.client_secret` | LinuxDo OAuth 客户端密钥 | - |
| `linuxdo.backend_base_url` | 服务的公网地址 | `http://your-domain:7643` |
| `jwt.secret` | JWT 签名密钥 | 强随机字符串 |

### 获取 LinuxDo OAuth 凭据

1. 访问 [LinuxDo](https://linux.do)
2. 前往开发者设置创建 OAuth 应用
3. 设置回调 URL 为: `http://your-domain:7643/api/auth/linuxdo/callback`
4. 获取 `client_id` 和 `client_secret`

## 开发

### 运行测试

```bash
go test ./...
```

### 构建

```bash
# 本地构建
go build -o cosine2api

# Docker 构建
docker build -t cosine2api:latest .
```

### 代码格式化

```bash
go fmt ./...
```

## 部署建议

### 生产环境注意事项

1. **安全性**:
   - 修改所有默认密码
   - 使用强随机字符串作为 JWT 密钥
   - 启用 HTTPS（建议使用 Nginx 反向代理）
   - 配置防火墙规则

2. **性能优化**:
   - 配置适当的数据库连接池
   - 启用 Gzip 压缩
   - 使用 CDN 加速静态资源

3. **监控**:
   - 配置日志收集
   - 设置服务监控和告警
   - 定期备份数据库

### Nginx 反向代理示例

```nginx
server {
    listen 80;
    server_name your-domain.com;

    location / {
        proxy_pass http://localhost:7643;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_cache_bypass $http_upgrade;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    }
}
```

## 故障排查

### 常见问题

**1. 数据库连接失败**

检查数据库配置是否正确，确保 PostgreSQL 服务正在运行：

```bash
docker compose logs db
```

**2. OAuth 回调失败**

确保 [config.yaml](config.yaml) 中的 `backend_base_url` 配置正确，并且可以从外部访问。

**3. 流式响应中断**

检查 Cosine API 的 auth cookie 是否有效，是否需要刷新。

### 查看日志

```bash
# Docker 部署
docker compose logs -f app

# 本地部署
# 日志输出到标准输出
```

## 贡献

欢迎提交 Issue 和 Pull Request！

## 许可证

MIT License

## 相关链接

- [Cosine 官网](https://cosine.sh)
- [LinuxDo 社区](https://linux.do)
- [OpenAI API 文档](https://platform.openai.com/docs/api-reference)

---

如有问题或建议，请提交 Issue 或联系维护者。
