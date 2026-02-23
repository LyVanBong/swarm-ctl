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
- `app install` — Cài đặt nhanh ứng dụng mẫu (n8n, nocodb, metabase...) *(v1.0)*
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
├── app
│   ├── list     (Marketplace)
│   └── install  APP-ID [--domain DOMAIN]
├── dashboard
└── version
```

---

## ⚙️ Yêu cầu

### Máy chạy swarm-ctl
- Linux/macOS/Windows
- SSH key có quyền truy cập

## 📚 Documentation (Runbooks)

Để vận hành hệ thống trơn tru và chuyên nghiệp trong production, `swarm-ctl` cung cấp các Runbooks tài liệu (Khắc phục sự cố theo từng bước):

* [Disaster Recovery & Backup](docs/runbooks/01-disaster-recovery.md) - Cứu hoả khi sập DB, sập toàn bộ Master, thao tác với Backup/Restore.
* [Node Management](docs/runbooks/02-node-management.md) - Cách Add thêm máy ảo, thu hẹp ứng dụng, cách cập nhật Kernel Linux cho máy Worker Node an toàn.
* [Service Updates](docs/runbooks/03-service-updates.md) - Cách deploy lại 1 app, tự động đổi mật khẩu Database (zero-downtime rotation), rollback phiên bản bị lỗi.

## 🔗 Các Liên Kết Quan Trọng (Important Links)
* [Kiến trúc Mô Hình Cluster](docs/architecture.md)
* [Lịch sử Cập Nhật (CHANGELOG)](CHANGELOG.md)
* [Hướng dẫn Đóng góp Mã nguồn (CONTRIBUTING)](CONTRIBUTING.md)
* [Giấy phép MIT (LICENSE)](LICENSE)

## 📦 Công nghệ sử dụng
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

## 🚀 Các dịch vụ (Services) được triển khai mặc định

Khi bạn chạy lệnh `swarm-ctl cluster init` để khởi tạo Cluster, `swarm-ctl` sẽ tự động thiết lập 2 nhóm dịch vụ (Tier 1 và Tier 2) chuẩn Enterprise để bạn có ngay hệ sinh thái hoàn chỉnh mà không cần cấu hình thủ công:

### Tier 1: Hạ tầng (Infrastructure) - *Bắt buộc*
* **Traefik (Gateway):** Router, Load Balancer siêu cấp tự động xin chứng chỉ SSL/TLS (Let's Encrypt) cho mọi tên miền. (Có trang Dashboard quản lý + Auth).
* **Portainer (Management):** WebUI quản lý Docker/Swarm trực quan ngoài Dashboard CLI.
* **Middlewares Security:** Thiết lập sẵn các bộ lọc chống Ddos (RateLimit), Security Headers ngăn chặn iFrame, XSS.

### Tier 2: Dịch vụ nền tảng (Platform Services) - *Bắt buộc*
* **MariaDB Database:** Hệ quản trị CSDL siêu tối ưu cho Production, thiết đặt sẵn tự động backup.
* **Redis Cache:** Cache In-Memory và Session Storage tiêu chuẩn.
* **MinIO (S3-compatible):** Server lưu trữ vệ tinh phân tán (Chứa file upload, ảnh, video của users).
* **Monitoring Stack trọn bộ:**
  * **Prometheus:** Gom metric (CPU, RAM, Network) toàn bộ cụm.
  * **Grafana:** Vẽ biểu đồ giám sát Data real-time.
  * **Loki & Promtail:** Thu gom Logs (Console Logs) của tất cả containers dồn về 1 mối tập trung. Không cần SSH vào từng node đọc log.
  * **Alertmanager:** Server cấu hình kịch bản báo động tới Telegram/Slack khi hệ thống sập.

## 💻 Yêu cầu hệ thống tối thiểu (System Requirements)

Để chạy trơn tru kiến trúc bên trên, bạn cần chuẩn bị tài nguyên cơ bản (Cloud Server/VPS):

### 1. Master/Manager Node (Tối thiểu 1 máy)
Máy chứa Database, Monitoring và điều khiển Swarm.
- **CPU:** Tối thiểu 2 vCores (Khuyến nghị 4 vCores cho production).
- **RAM:** Tối thiểu 4GB (Khuyến nghị 8GB+ để chạy đủ MinIO, MariaDB, Prometheus).
- **Disk:** Tối thiểu 50GB SSD/NVMe (Nên mở rộng >100GB nếu nạp lượng lớn File Upload).
- **OS:** Linux (Ubuntu 20.04+, Debian 11+).

### 2. Worker Nodes (Dành cho Ứng dụng/Web/API) - Có thể thêm sau
Máy chuyên chạy ứng dụng (Backend, Frontend).
- **CPU:** 1-2 vCores là đủ.
- **RAM:** 2GB.
- **OS:** Linux.

*Lưu ý: Tất cả các máy tính cần kết nối qua Mạng LAN Nội bộ (Private IP) để tốc độ đồng bộ Cluster tối ưu nhất và giảm độ trễ (Latency).*
