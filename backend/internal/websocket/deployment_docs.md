# WebSocket推送服务部署文档

## 概述

本文档描述了WebSocket推送服务的生产环境部署指南，包括系统要求、配置、部署步骤、监控和故障排查。

## 系统要求

### 硬件要求

- **CPU**: 4核心以上，推荐8核心
- **内存**: 8GB以上，推荐16GB
- **存储**: 50GB以上SSD
- **网络**: 1Gbps以上带宽

### 软件要求

- **操作系统**: Linux (Ubuntu 20.04+ / CentOS 8+)
- **Go版本**: 1.21+
- **依赖服务**: Redis, PostgreSQL (可选)

## 配置说明

### 基础配置

```yaml
# config.yaml
server:
  port: 8080
  read_buffer_size: 1024
  write_buffer_size: 1024

subscription:
  max_connections: 5000
  max_subscriptions: 50000
  cleanup_interval: 30s
  inactive_timeout: 5m

broadcast:
  max_queue_size: 50000
  worker_count: 50
  retry_attempts: 3
  retry_delay: 1s
  batch_size: 100
  batch_timeout: 100ms

heartbeat:
  heartbeat_interval: 30s
  pong_timeout: 60s
  max_missed_heartbeats: 3

reconnect:
  reconnect_interval: 5s
  max_reconnect_attempts: 10
  backoff_multiplier: 1.5
  max_backoff_interval: 60s

performance:
  sampling_rate: 0.1
  aggregation_interval: 1s
  retention_period: 24h
  enable_alerts: true
  alert_cooldown: 5m

log:
  level: info
  format: json
  output: /var/log/websocket/websocket.log
  max_size: 100
  max_backups: 3
  max_age: 7
  compress: true

environment: production
debug: false
```

### 环境变量配置

```bash
# 环境变量
export WEBSOCKET_PORT=8080
export WEBSOCKET_MAX_CONNECTIONS=5000
export WEBSOCKET_LOG_LEVEL=info
export WEBSOCKET_ENVIRONMENT=production
export REDIS_URL=redis://localhost:6379
export DATABASE_URL=postgres://user:pass@localhost:5432/websocket
```

## 部署步骤

### 1. 准备环境

```bash
# 更新系统
sudo apt update && sudo apt upgrade -y

# 安装依赖
sudo apt install -y curl wget git build-essential

# 安装Go
wget https://go.dev/dl/go1.21.0.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.0.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc
```

### 2. 安装Redis

```bash
# Ubuntu/Debian
sudo apt install redis-server

# CentOS/RHEL
sudo yum install redis

# 启动Redis
sudo systemctl start redis
sudo systemctl enable redis
```

### 3. 构建应用

```bash
# 克隆代码
git clone <repository-url>
cd websocket-push-service

# 构建应用
go mod tidy
go build -o websocket-service ./cmd/server

# 创建部署目录
sudo mkdir -p /opt/websocket-service
sudo cp websocket-service /opt/websocket-service/
sudo cp config.yaml /opt/websocket-service/
```

### 4. 创建系统服务

```bash
# 创建systemd服务文件
sudo tee /etc/systemd/system/websocket-service.service > /dev/null <<EOF
[Unit]
Description=WebSocket Push Service
After=network.target redis.service

[Service]
Type=simple
User=websocket
Group=websocket
WorkingDirectory=/opt/websocket-service
ExecStart=/opt/websocket-service/websocket-service
Restart=always
RestartSec=5
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
EOF

# 创建用户
sudo useradd -r -s /bin/false websocket
sudo chown -R websocket:websocket /opt/websocket-service

# 启动服务
sudo systemctl daemon-reload
sudo systemctl enable websocket-service
sudo systemctl start websocket-service
```

### 5. 配置Nginx反向代理

```bash
# 安装Nginx
sudo apt install nginx

# 创建配置文件
sudo tee /etc/nginx/sites-available/websocket > /dev/null <<EOF
upstream websocket {
    server 127.0.0.1:8080;
}

server {
    listen 80;
    server_name your-domain.com;

    location /ws {
        proxy_pass http://websocket;
        proxy_http_version 1.1;
        proxy_set_header Upgrade \$http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
        proxy_read_timeout 86400;
    }
}
EOF

# 启用站点
sudo ln -s /etc/nginx/sites-available/websocket /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl reload nginx
```

### 6. 配置SSL证书

```bash
# 安装Certbot
sudo apt install certbot python3-certbot-nginx

# 获取SSL证书
sudo certbot --nginx -d your-domain.com

# 自动续期
sudo crontab -e
# 添加: 0 12 * * * /usr/bin/certbot renew --quiet
```

## 监控和日志

### 1. 日志配置

```bash
# 创建日志目录
sudo mkdir -p /var/log/websocket
sudo chown websocket:websocket /var/log/websocket

# 配置日志轮转
sudo tee /etc/logrotate.d/websocket > /dev/null <<EOF
/var/log/websocket/*.log {
    daily
    missingok
    rotate 7
    compress
    delaycompress
    notifempty
    create 644 websocket websocket
    postrotate
        systemctl reload websocket-service
    endscript
}
EOF
```

### 2. 监控配置

```bash
# 安装Prometheus
wget https://github.com/prometheus/prometheus/releases/download/v2.40.0/prometheus-2.40.0.linux-amd64.tar.gz
tar xzf prometheus-2.40.0.linux-amd64.tar.gz
sudo mv prometheus-2.40.0.linux-amd64 /opt/prometheus

# 创建Prometheus配置
sudo tee /opt/prometheus/prometheus.yml > /dev/null <<EOF
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: 'websocket-service'
    static_configs:
      - targets: ['localhost:8080']
EOF

# 启动Prometheus
sudo /opt/prometheus/prometheus --config.file=/opt/prometheus/prometheus.yml &
```

### 3. 健康检查

```bash
# 创建健康检查脚本
sudo tee /opt/websocket-service/healthcheck.sh > /dev/null <<EOF
#!/bin/bash
curl -f http://localhost:8080/health || exit 1
EOF

sudo chmod +x /opt/websocket-service/healthcheck.sh

# 添加到crontab
sudo crontab -e
# 添加: */5 * * * * /opt/websocket-service/healthcheck.sh
```

## 性能优化

### 1. 系统优化

```bash
# 调整文件描述符限制
echo "* soft nofile 65536" >> /etc/security/limits.conf
echo "* hard nofile 65536" >> /etc/security/limits.conf

# 调整内核参数
echo "net.core.somaxconn = 65536" >> /etc/sysctl.conf
echo "net.ipv4.tcp_max_syn_backlog = 65536" >> /etc/sysctl.conf
echo "net.core.netdev_max_backlog = 5000" >> /etc/sysctl.conf
sysctl -p
```

### 2. 应用优化

```yaml
# 生产环境优化配置
server:
  read_buffer_size: 2048
  write_buffer_size: 2048

broadcast:
  worker_count: 100
  batch_size: 500
  batch_timeout: 50ms

performance:
  sampling_rate: 0.05  # 降低采样率
  aggregation_interval: 5s
```

## 故障排查

### 1. 常见问题

#### 连接数过多
```bash
# 检查连接数
netstat -an | grep :8080 | wc -l

# 调整连接限制
sudo sysctl -w net.core.somaxconn=65536
```

#### 内存使用过高
```bash
# 检查内存使用
free -h
ps aux --sort=-%mem | head

# 调整GC参数
export GOGC=100
export GOMEMLIMIT=8GiB
```

#### 消息延迟
```bash
# 检查队列大小
curl http://localhost:8080/metrics | grep queue_size

# 调整批处理参数
broadcast:
  batch_size: 1000
  batch_timeout: 10ms
```

### 2. 日志分析

```bash
# 查看服务日志
sudo journalctl -u websocket-service -f

# 分析错误日志
sudo journalctl -u websocket-service --since "1 hour ago" | grep ERROR

# 查看连接日志
tail -f /var/log/websocket/websocket.log | grep "connection"
```

### 3. 性能分析

```bash
# 使用pprof分析
go tool pprof http://localhost:8080/debug/pprof/profile
go tool pprof http://localhost:8080/debug/pprof/heap

# 监控goroutine
curl http://localhost:8080/debug/pprof/goroutine
```

## 备份和恢复

### 1. 配置备份

```bash
# 备份配置文件
sudo cp /opt/websocket-service/config.yaml /backup/websocket-config-$(date +%Y%m%d).yaml

# 备份系统服务配置
sudo cp /etc/systemd/system/websocket-service.service /backup/
```

### 2. 日志备份

```bash
# 创建日志备份脚本
sudo tee /opt/websocket-service/backup-logs.sh > /dev/null <<EOF
#!/bin/bash
DATE=\$(date +%Y%m%d)
tar -czf /backup/websocket-logs-\$DATE.tar.gz /var/log/websocket/
find /backup -name "websocket-logs-*.tar.gz" -mtime +30 -delete
EOF

sudo chmod +x /opt/websocket-service/backup-logs.sh

# 添加到crontab
sudo crontab -e
# 添加: 0 2 * * * /opt/websocket-service/backup-logs.sh
```

## 安全配置

### 1. 防火墙配置

```bash
# 配置UFW防火墙
sudo ufw allow 22/tcp
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
sudo ufw enable
```

### 2. SSL/TLS配置

```bash
# 生成自签名证书（测试环境）
sudo openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
  -keyout /etc/ssl/private/websocket.key \
  -out /etc/ssl/certs/websocket.crt
```

### 3. 访问控制

```yaml
# 配置访问控制
server:
  check_origin: true
  allowed_origins:
    - "https://yourdomain.com"
    - "https://app.yourdomain.com"
```

## 扩展和负载均衡

### 1. 水平扩展

```bash
# 使用Nginx负载均衡
upstream websocket_cluster {
    server 127.0.0.1:8080;
    server 127.0.0.1:8081;
    server 127.0.0.1:8082;
}
```

### 2. 数据库集群

```yaml
# Redis集群配置
redis:
  cluster:
    nodes:
      - "redis1:6379"
      - "redis2:6379"
      - "redis3:6379"
```

## 维护和更新

### 1. 滚动更新

```bash
# 更新服务
sudo systemctl stop websocket-service
sudo cp new-websocket-service /opt/websocket-service/
sudo systemctl start websocket-service
```

### 2. 版本管理

```bash
# 使用Docker进行版本管理
docker build -t websocket-service:v1.0.0 .
docker tag websocket-service:v1.0.0 websocket-service:latest
```

## 联系信息

如有问题，请联系：
- 技术支持: tech-support@company.com
- 紧急联系: +1-xxx-xxx-xxxx
- 文档更新: 2024-01-01
