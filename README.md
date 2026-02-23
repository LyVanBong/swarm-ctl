# swarm-ctl

<div align="center">

```text
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

`swarm-ctl` là công cụ CLI giúp quản lý Docker Swarm cluster ở quy mô Doanh nghiệp (Enterprise) từ **một điểm duy nhất**. Tool tự động hóa toàn bộ công đoạn khó nhằn nhất của DevOps: cấu hình HA, Load Balancing tự động SSL, Zero-downtime deploy và Giám sát máy chủ chuyên nghiệp.

**Chỉ cần cung cấp SSH key → Tool tự động làm mọi thứ còn lại.**

```bash
# 1. Khởi tạo toàn bộ cluster tự động (Docker + Swarm + Traefik + Monitoring)
swarm-ctl cluster init --master 10.0.0.1 --key ~/.ssh/id_rsa --domain softty.net

# 2. Triển khai Ứng dụng mẫu Mã nguồn mở từ Marketplace
swarm-ctl app install nocodb --domain data.softty.net

# 3. Theo dõi toàn cảnh Sức khỏe Cluster thông qua TUI màu sắc
swarm-ctl dashboard
```

---

## 💻 Yêu cầu Hệ thống (System Requirements)

Để chạy trơn tru kiến trúc Cluster chuẩn, bạn cần chuẩn bị server với hệ điều hành **Ubuntu 20.04+** hoặc **Debian 11+** có mở SSH Port 22:

### 1. Master/Manager Node (Khuyên dùng tối thiểu 1 máy)
Máy chứa Database, Monitoring và điều khiển Swarm.
- **CPU:** Tối thiểu 2 vCores (Khuyến nghị 4 vCores cho production).
- **RAM:** Tối thiểu 4GB (Khuyến nghị 8GB+ để chạy đủ MinIO, MariaDB, Prometheus).
- **Disk:** Tối thiểu 50GB SSD/NVMe.

### 2. Worker Nodes (Tuỳ chọn ráp sau)
Dành cho ứng dụng Backend, Web, API để phân tán tải.
- **CPU:** 1-2 vCores / **RAM:** 2GB.
*(Tất cả kết nối qua LAN Nội bộ Private IP để giảm độ trễ)*

### 3. Máy cá nhân chạy lệnh (`swarm-ctl`)
- Linux / macOS / Windows.
- Khóa SSH (SSH Key) có quyền đăng nhập vào các Servers.

---

## 📦 Cài đặt Công Cụ (`swarm-ctl`)

### Cách 1: Tải nhanh Binary (Khuyến nghị)
```bash
# Dành cho Linux AMD64 (Ubuntu/Debian/CentOS)
curl -L https://github.com/LyVanBong/swarm-ctl/releases/latest/download/swarm-ctl-linux-amd64 \
  -o /usr/local/bin/swarm-ctl && chmod +x /usr/local/bin/swarm-ctl

# Kiểm tra cài đặt
swarm-ctl version
```

### Cách 2: Build từ Source Code
*Yêu cầu Go 1.23+*
```bash
git clone https://github.com/LyVanBong/swarm-ctl
cd swarm-ctl
make build
sudo mv swarm-ctl /usr/local/bin/
```

---

## 🏗️ Kiến trúc & Dịch vụ Nền tảng Mặc định

Tool phân chia Cluster thành **3 Tier (3 Lớp mạng)**. Ngay khi lệnh `cluster init` thành công, các dịch vụ này tự động chạy ẩn:

1. **Tier 1 (Hạ tầng - Ingress):**
   - Traefik: Cổng Router, Load Balancer tự động cấp HTTPS (Let's Encrypt).
   - Portainer: Giao diện web quản lý Docker.
2. **Tier 2 (Platform - Dữ liệu vĩnh viễn):**
   - **Database & Cache:** MariaDB Galera (CSDL rập khuôn), Redis Sentinel.
   - **Lưu trữ:** MinIO (Private S3 Cloud).
   - **Giám sát (Observability Stack):** Prometheus (Metrics) + Grafana (Dashboard) + Loki (Gom Logs) + Alertmanager.
3. **Tier 3 (Applications):** Chứa các App do bạn tự Deploy (Wordpress, Web, NodeJS, API...).

---

## ✨ Tính Năng Nổi Bật (Ecosystem)

- 🔒 **Zero-Trust Connection:** Không mở cổng Docker Socket ra Internet. Tool hoàn toàn ra lệnh máy chủ thông qua kênh SSH mã hóa nội bộ.
- ♻️ **Zero-Downtime Deploy:** Hỗ trợ tính năng `service update` (khởi chạy bản mới trước khi tắt bản cũ) hoặc `secret rotate` (đổi mật khẩu database mà API đang kết nối không chết).
- 📂 **Distributed Storage (GlusterFS):** Hỗ trợ `storage init-glusterfs` để đồng bộ ổ đĩa giữa 3 con Worker/Master chống cháy nổ vật lý.
- 🗄️ **Cứu Hộ Dữ Liệu:** Lệnh `backup create/restore` tự động zip toàn thư mục phân quyền và tạo SQL Export cứu sống Cluster phút mốt.
- 📝 **Nhật ký Kiểm Toán (Audit Trail):** Mọi lệnh can thiệp thay đổi (`node remove`, `service add`) đều ghi nhật ký tên PC, Thời gian vào file local `/audit.log` và loại bỏ hoàn toàn các Argument mang tính nhạy cảm như (Password/Key). 

---

## 🛒 Chợ Ứng Dụng (1-Click Marketplace)

Với hơn 15+ phần mềm Enterprise/Open-source nổi tiếng. Bạn không cần tự viết `docker-compose`. Chỉ cần gõ:
```bash
swarm-ctl app list
swarm-ctl app install [tên-app] --domain sub.congty.com
```

Một số App đỉnh nhất tích hợp sẵn:
* **Tự động hóa:** `n8n` (Chuẩn Zapier)
* **Website & CMS:** `wordpress`, `ghost`
* **Công cụ nội bộ:** `nocodb` (Airtable-like), `metabase` (Trực quan Data), `nextcloud` (Google Drive tự cài).
* **Bảo mật & Dev:** `vaultwarden` (Quản lý Mật khẩu), `gitea` (Trạm Code nội bộ), `uptime-kuma` (Đo Ping Web).

---

## 📋 Bảng Lệnh Tham Khảo (CLI Commands)

```text
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
│   ├── init-glusterfs --nodes IP1,IP2..
│   └── expand   --node IP
├── backup
│   ├── create
│   ├── restore  BACKUP-ID
│   └── list
├── app
│   ├── list     (Trình chiếu Marketplace)
│   └── install  APP-ID [--domain DOMAIN]
├── dashboard    (Live Terminal UI)
├── audit        (Xem nhật ký thao tác)
└── version
```

---

## 📚 Tài Liệu Hướng Dẫn Vận Hành (Runbooks) & Liên kết Phụ

Tham khảo thêm các hướng dẫn cấu hình kỹ thuật sâu hơn tại thư mục [docs/runbooks](docs/runbooks/):

* [Khôi phục Thảm Họa (Disaster Recovery & Backup)](docs/runbooks/01-disaster-recovery.md) - Cứu hoả khi sập DB, sập toàn bộ Master.
* [Quản lý Máy chủ (Node Management)](docs/runbooks/02-node-management.md) - Cách Add thêm máy ảo, cách cập nhật Kernel Linux an toàn.
* [Cập nhật Dịch vụ (Service Updates)](docs/runbooks/03-service-updates.md) - Re-deploy, Rotate Mật khẩu không Downtime.
* [Kiến trúc Mô Hình Cluster](docs/architecture.md)
* [Lịch sử Cập Nhật (CHANGELOG)](CHANGELOG.md)

---

## 🤝 Đóng góp Mã Nguồn (Contributing)

Dự án là của Cộng đồng! Bạn có ý tưởng phát triển App mới vào Marketplace, hay Report lỗi, xin vui lòng xem [Hướng dẫn Đóng góp (CONTRIBUTING)](CONTRIBUTING.md). Cảm ơn vì đã giúp `swarm-ctl` trở nên tốt hơn!

---

## 📄 Bản Quyền (License)

Dự án được phân phối dưới giấy phép [MIT License](LICENSE). 

<div align="center">
<b>Được phát triển bền vững bởi <a href="https://github.com/LyVanBong">Ly Van Bong</a> & <a href="https://www.softty.net">Softty Net</a></b>
</div>
