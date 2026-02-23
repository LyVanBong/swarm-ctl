# swarm-ctl

<div align="center">

```
 ███████╗██╗    ██╗ █████╗ ██████╗ ███╗   ███╗      ██████╗████████╗██╗     
 ██╔════╝██║    ██║██╔══██╗██╔══██╗████╗ ████║     ██╔════╝╚══██╔══╝██║     
 ███████╗██║ █╗ ██║███████║██████╔╝██╔████╔██║     ██║        ██║   ██║     
 ╚════██║██║███╗██║██╔══██║██╔══██╗██║╚██╔╝██║     ██║        ██║   ██║     
 ███████║╚███╔███╔╝██║  ██║██║  ██║██║ ╚═╝ ██║     ╚██████╗   ██║   ███████╗
 ╚══════╝ ╚══╝╚══╝ ╚═╝  ╚═╝╚═╝  ╚═╝╚═╝     ╚═╝      ╚═════╝   ╚═╝   ╚══════╝
```

**Enterprise Docker Swarm Management CLI**

[![Go Version](https://img.shields.io/badge/go-1.23+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)
[![Release](https://img.shields.io/github/v/release/LyVanBong/swarm-ctl)](https://github.com/LyVanBong/swarm-ctl/releases)

</div>

---

## 🚀 Giới thiệu

`swarm-ctl` là công cụ CLI giúp quản lý Docker Swarm cluster ở quy mô enterprise từ **một điểm duy nhất**.

**Chỉ cần cung cấp SSH key → Tool tự động làm mọi thứ còn lại.**

```bash
# Khởi tạo cluster mới hoàn toàn tự động
swarm-ctl cluster init --master 10.0.0.1 --key ~/.ssh/id_rsa --domain example.com

# Thêm node mới (provision + join swarm tự động)
swarm-ctl node add --ip 10.0.0.5 --role worker

# Deploy service với SSL tự động
swarm-ctl service add --name my-api --image nginx:latest --domain api.example.com --port 3000

# Scale khi traffic tăng
swarm-ctl service scale my-api=10
```

---

## ✨ Tính năng

### 🖥️ Cluster Management
- `cluster init` — Bootstrap cluster đầy đủ (Docker + Swarm + Traefik + Portainer + Monitoring)
- `cluster status` — Tổng quan health của cluster
- `cluster upgrade` — Nâng cấp Docker zero-downtime *(v1.1)*

### 📦 Node Management
- `node add` — Provision + join node mới hoàn toàn tự động qua SSH + Ansible
- `node remove` — Drain an toàn rồi xóa node
- `node list` — Danh sách nodes với status real-time
- `node ssh` — SSH nhanh vào node

### 🚀 Service Management
- `service add` — Tạo service mới với wizard (tự generate compose + labels Traefik)
- `service deploy` — Deploy từ services.yml
- `service scale` — Scale replicas
- `service update` — Rolling update zero-downtime
- `service rollback` — Rollback về version trước
- `service logs` — Live logs
- `service remove` — Xóa service

### 🔐 Secret Management
- `secret add/list/remove` — Quản lý Docker Secrets
- `secret rotate` — Rotate secret không downtime *(v1.1)*

### 💾 Storage & Backup *(v1.1)*
- `storage status` — MinIO/GlusterFS health
- `backup create/restore/list` — Backup automation

### 📊 Dashboard *(v1.2)*
- `dashboard` — Live TUI dashboard (Bubbletea)

---

## 📦 Cài đặt

### Download binary (Khuyến nghị)

```bash
# Linux AMD64
curl -L https://github.com/LyVanBong/swarm-ctl/releases/latest/download/swarm-ctl-linux-amd64 \
  -o /usr/local/bin/swarm-ctl && chmod +x /usr/local/bin/swarm-ctl

# Kiểm tra
swarm-ctl version
```

### Build từ source

```bash
git clone https://github.com/LyVanBong/swarm-ctl
cd swarm-ctl
go build -o swarm-ctl .
sudo mv swarm-ctl /usr/local/bin/
```

---

## 🏗️ Kiến trúc

Tool hoạt động theo mô hình **2 lớp**:

```
swarm-ctl (CLI)          Ansible (Engine)         Target Nodes
     │                        │                        │
     ├─ node add ────────────►│── SSH ────────────────►│
     │                        │   Install Docker       │
     │                        │   Setup directories    │
     │                        │   Join Swarm          │
     │◄────────── Done ───────┤                        │
```

### Tier 1: Infrastructure (Auto-deployed khi `cluster init`)
- **Traefik** — Reverse proxy, SSL (Let's Encrypt wildcard), load balancing
- **Portainer** — Management UI

### Tier 2: Platform (Auto-deployed khi `cluster init`)  
- **MinIO** — S3-compatible distributed storage
- **MariaDB Galera** — HA database cluster (3 nodes)
- **Redis Sentinel** — HA cache/queue
- **Prometheus + Grafana + Loki** — Observability stack

### Tier 3: Applications (User tự deploy)
- Appwrite, WordPress, n8n, NocoDB, và bất kỳ service nào khác

---

## 📋 Commands Reference

```
swarm-ctl
├── cluster
│   ├── init     --master IP --key PATH --domain DOMAIN [--name NAME]
│   ├── status
│   └── upgrade
├── node
│   ├── add      --ip IP [--role worker|manager] [--label KEY=VAL]
│   ├── remove   --ip IP [--force]
│   ├── list
│   └── ssh      IP
├── service
│   ├── add      --name N --image IMG --domain DOM --port PORT [--replicas N]
│   ├── deploy   SERVICE-NAME
│   ├── scale    SERVICE=REPLICAS
│   ├── update   SERVICE [--image IMG]
│   ├── rollback SERVICE
│   ├── logs     SERVICE
│   ├── list
│   └── remove   SERVICE
├── secret
│   ├── add      NAME VALUE
│   ├── list
│   ├── remove   NAME
│   └── rotate   NAME NEW-VALUE
├── storage
│   ├── status
│   └── expand   --node IP
├── backup
│   ├── create
│   ├── restore  BACKUP-ID
│   └── list
├── dashboard
└── version
```

---

## ⚙️ Yêu cầu

### Máy chạy swarm-ctl
- Linux/macOS/Windows
- SSH key có quyền truy cập các nodes
- Ansible (tự động cài nếu chưa có)

### Các nodes trong cluster
- Ubuntu 20.04+ hoặc Debian 11+
- SSH port 22 mở
- Quyền sudo

---

## 🤝 Contributing

Đóng góp luôn được chào đón! Xem [CONTRIBUTING.md](CONTRIBUTING.md).

---

## 📄 License

MIT License — xem [LICENSE](LICENSE)

---

**Tác giả**: [Ly Van Bong](https://github.com/LyVanBong)  
**Website**: https://www.softty.net
