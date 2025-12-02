# Nginx WebSocket 代理配置指南

## 📋 配置概述

本文档说明如何配置 Nginx 代理 ttpos-websocket 服务的 WebSocket 连接。

## 🔧 配置文件位置

- **项目配置文件**: `/home/coder/workspaces/ttpos-server-go/docker/nginx/conf.d/ttpos-websocket.conf`
- **生产环境**: `/etc/nginx/conf.d/websocket.conf`

## 📝 基础配置

### 1. 单独的 WebSocket 配置文件

```nginx
# /etc/nginx/conf.d/websocket.conf

# WebSocket 连接代理
location /ws {
    # 后端 WebSocket 服务地址
    proxy_pass http://10.0.10.45:14051;
    
    # 使用 HTTP/1.1
    proxy_http_version 1.1;
    
    # WebSocket 必需的头部
    proxy_set_header Upgrade $http_upgrade;
    proxy_set_header Connection "upgrade";
    
    # 基础代理头部
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;
    
    # 超时配置（重要：WebSocket 需要较长的超时时间）
    proxy_connect_timeout 60s;
    proxy_send_timeout 3600s;      # 1小时
    proxy_read_timeout 3600s;      # 1小时
    
    # 禁用缓冲
    proxy_buffering off;
    proxy_cache off;
}
```

### 2. 完整的 Server 配置示例

```nginx
# /etc/nginx/sites-available/ttpos.conf

upstream websocket_backend {
    # WebSocket 后端服务器
    server 10.0.10.45:14051;
    
    # 可选：负载均衡配置（多个 WebSocket 实例）
    # server 10.0.10.46:14051;
    # server 10.0.10.47:14051;
    
    # 保持连接
    keepalive 32;
}

server {
    listen 80;
    server_name your-domain.com;
    
    # 重定向到 HTTPS（生产环境推荐）
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name your-domain.com;
    
    # SSL 证书配置
    ssl_certificate /etc/nginx/ssl/your-domain.crt;
    ssl_certificate_key /etc/nginx/ssl/your-domain.key;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers HIGH:!aNULL:!MD5;
    
    # 日志配置
    access_log /var/log/nginx/ttpos_access.log;
    error_log /var/log/nginx/ttpos_error.log;
    
    # WebSocket 路径
    location /ws {
        proxy_pass http://websocket_backend;
        proxy_http_version 1.1;
        
        # WebSocket 升级
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        
        # 代理头部
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        # 超时配置
        proxy_connect_timeout 60s;
        proxy_send_timeout 3600s;
        proxy_read_timeout 3600s;
        
        # 禁用缓冲
        proxy_buffering off;
    }
    
    # WebSocket gRPC API（如果需要）
    location /api/v1/websocket {
        proxy_pass http://10.0.10.45:14052;
        proxy_http_version 1.1;
        
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    }
    
    # 其他 API 路径
    location /api/ {
        proxy_pass http://your-api-server;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    }
    
    # 静态文件
    location / {
        root /var/www/html;
        index index.html;
        try_files $uri $uri/ /index.html;
    }
}
```

## 🚀 高级配置

### 1. 负载均衡配置

```nginx
# 多个 WebSocket 实例的负载均衡
upstream websocket_backend {
    # IP Hash：确保同一客户端始终连接到同一后端
    ip_hash;
    
    server 10.0.10.45:14051 weight=1 max_fails=3 fail_timeout=30s;
    server 10.0.10.46:14051 weight=1 max_fails=3 fail_timeout=30s;
    server 10.0.10.47:14051 weight=1 max_fails=3 fail_timeout=30s;
    
    keepalive 32;
}

location /ws {
    proxy_pass http://websocket_backend;
    # ... 其他配置
}
```

### 2. 路径重写

```nginx
# 将 /websocket 重写为 /ws
location /websocket {
    rewrite ^/websocket(.*)$ /ws$1 break;
    proxy_pass http://10.0.10.45:14051;
    proxy_http_version 1.1;
    proxy_set_header Upgrade $http_upgrade;
    proxy_set_header Connection "upgrade";
    # ... 其他配置
}
```

### 3. 基于子域名的配置

```nginx
# WebSocket 专用子域名
server {
    listen 443 ssl http2;
    server_name ws.your-domain.com;
    
    # SSL 配置
    ssl_certificate /etc/nginx/ssl/ws.your-domain.crt;
    ssl_certificate_key /etc/nginx/ssl/ws.your-domain.key;
    
    # 所有请求都代理到 WebSocket 服务
    location / {
        proxy_pass http://10.0.10.45:14051;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        # ... 其他配置
    }
}
```

### 4. 限流配置

```nginx
# 定义限流区域
limit_conn_zone $binary_remote_addr zone=websocket_conn:10m;
limit_req_zone $binary_remote_addr zone=websocket_req:10m rate=10r/s;

location /ws {
    # 每个 IP 最多 10 个并发连接
    limit_conn websocket_conn 10;
    
    # 每秒最多 10 个请求
    limit_req zone=websocket_req burst=20 nodelay;
    
    proxy_pass http://10.0.10.45:14051;
    # ... 其他配置
}
```

### 5. 安全配置

```nginx
location /ws {
    # 只允许特定 IP 访问（可选）
    # allow 192.168.1.0/24;
    # allow 10.0.0.0/8;
    # deny all;
    
    # 防止跨站 WebSocket 劫持
    if ($http_origin !~* (https?://your-domain\.com)) {
        return 403;
    }
    
    proxy_pass http://10.0.10.45:14051;
    # ... 其他配置
}
```

## 📊 Map 配置（升级头部）

在 `http {}` 块中添加：

```nginx
http {
    # WebSocket 升级头部映射
    map $http_upgrade $connection_upgrade {
        default upgrade;
        '' close;
    }
    
    server {
        location /ws {
            proxy_pass http://10.0.10.45:14051;
            proxy_http_version 1.1;
            proxy_set_header Upgrade $http_upgrade;
            proxy_set_header Connection $connection_upgrade;
            # ... 其他配置
        }
    }
}
```

## 🔍 重要配置项说明

### 1. 超时配置

| 配置项 | 建议值 | 说明 |
|-------|--------|------|
| `proxy_connect_timeout` | 60s | 与后端建立连接的超时时间 |
| `proxy_send_timeout` | 3600s | 发送数据到后端的超时时间 |
| `proxy_read_timeout` | 3600s | 从后端读取数据的超时时间 |

**注意**：WebSocket 是长连接，需要设置较长的超时时间（如 1 小时或更长）。

### 2. 必需的头部

```nginx
proxy_set_header Upgrade $http_upgrade;        # 升级协议
proxy_set_header Connection "upgrade";         # 升级连接
proxy_http_version 1.1;                       # 使用 HTTP/1.1
```

### 3. 缓冲配置

```nginx
proxy_buffering off;    # 禁用缓冲（实时传输）
proxy_cache off;        # 禁用缓存
```

## 🧪 测试配置

### 1. 检查配置语法

```bash
# 测试 Nginx 配置
nginx -t

# 或
sudo nginx -t
```

### 2. 重载配置

```bash
# 重载 Nginx 配置（不中断服务）
nginx -s reload

# 或
sudo systemctl reload nginx
```

### 3. 测试 WebSocket 连接

```bash
# 使用 websocat 测试
websocat ws://your-domain.com/ws

# 使用 wscat 测试
wscat -c ws://your-domain.com/ws

# 使用 curl 测试 WebSocket 握手
curl -i -N \
  -H "Connection: Upgrade" \
  -H "Upgrade: websocket" \
  -H "Sec-WebSocket-Version: 13" \
  -H "Sec-WebSocket-Key: $(echo -n 'test' | base64)" \
  http://your-domain.com/ws
```

## 🐛 故障排查

### 1. 502 Bad Gateway

**原因**：
- 后端服务未启动
- 后端服务地址错误
- 防火墙阻止连接

**解决**：
```bash
# 检查后端服务状态
docker ps | grep websocket

# 测试后端连接
telnet 10.0.10.45 14051

# 检查防火墙
iptables -L -n
```

### 2. 连接立即断开

**原因**：
- 超时配置过短
- 缺少 WebSocket 升级头部

**解决**：
```nginx
# 增加超时时间
proxy_read_timeout 3600s;
proxy_send_timeout 3600s;

# 确保有升级头部
proxy_set_header Upgrade $http_upgrade;
proxy_set_header Connection "upgrade";
```

### 3. 无法升级到 WebSocket

**原因**：
- HTTP 版本不是 1.1
- 缺少升级头部

**解决**：
```nginx
# 必须使用 HTTP/1.1
proxy_http_version 1.1;

# 必须设置升级头部
proxy_set_header Upgrade $http_upgrade;
proxy_set_header Connection "upgrade";
```

## 📝 生产环境检查清单

- [ ] SSL/TLS 证书已配置
- [ ] 超时时间已设置（建议 3600s）
- [ ] 升级头部已配置正确
- [ ] HTTP 版本设置为 1.1
- [ ] 禁用了缓冲和缓存
- [ ] 配置了适当的日志
- [ ] 测试了 WebSocket 连接
- [ ] 配置了负载均衡（如有多实例）
- [ ] 配置了限流保护
- [ ] 配置了安全策略（CORS、IP 白名单等）

## 🔗 相关命令

```bash
# 启动 Nginx
sudo systemctl start nginx

# 停止 Nginx
sudo systemctl stop nginx

# 重启 Nginx
sudo systemctl restart nginx

# 重载配置
sudo systemctl reload nginx

# 查看 Nginx 状态
sudo systemctl status nginx

# 查看 Nginx 日志
tail -f /var/log/nginx/access.log
tail -f /var/log/nginx/error.log
tail -f /var/log/nginx/websocket_access.log
tail -f /var/log/nginx/websocket_error.log
```

## 📚 参考资料

- [Nginx WebSocket 官方文档](http://nginx.org/en/docs/http/websocket.html)
- [Nginx 反向代理文档](http://nginx.org/en/docs/http/ngx_http_proxy_module.html)
- [WebSocket RFC 6455](https://tools.ietf.org/html/rfc6455)

---

**文档版本**：1.0  
**最后更新**：2025-11-13  
**维护者**：ttpos-server-go 团队

