# Kiến trúc Hệ thống Docker Swarm Enterprise

Tài liệu này mô tả kiến trúc của cụm Docker Swarm được triển khai thông qua công cụ `swarm-ctl`.

## 1. Tổng quan Kiến trúc (High-Level Architecture)

Cụm Swarm chia làm 3 lớp (Tiers) tương đương 3 nhóm Network và Server Roles khác nhau:

```text
       [ Khách hàng / End Users ]
                  │
                  ▼ (Port 80/443)
┌──────────────────────────────────────────────┐
│                [ LỚP INGRESS ]               │
│               Manager Nodes                  │
│                                              │
│  Traefik (Reverse Proxy & Load Balancer)     │
│  Lấy SSL Let's Encrypt tự động               │
└─────────────────┬────────────────────────────┘
                  │ (proxy_public overlay network)
                  ▼
┌──────────────────────────────────────────────┐
│            [ LỚP APPLICATION ]               │
│             Worker/Manager Nodes             │
│                                              │
│  Appwrite | Frontend Web | Backend API       │
│  (Các Services tự động Scale Replica)        │
└─────────────────┬────────────────────────────┘
                  │ (app_internal overlay network)
                  ▼
┌──────────────────────────────────────────────┐
│               [ LỚP DATA ]                   │
│         Manager Nodes (Pinned Data)          │
│                                              │
│  MariaDB | Redis | MinIO (Storage S3)        │
│  Prometheus | Grafana | Loki (Logging)       │
└──────────────────────────────────────────────┘
```

## 2. Quy hoạch Mạng lưới (Docker Networks)

Chúng ta không bao giờ phơi API nội bộ ra ngoài Internet, vì vậy hệ thống chia làm các mạng `Overlay` sau:

1. **`proxy_public`**: Mạng lưới DMZ. Traefik sẽ được đính vào mạng lưới này cùng với các Web Backend (Những app cần external traffic).
2. **`app_internal`**: Mạng lưới xương sống của App. Backend API giao tiếp với Database/Redis thông qua mạng này. Traefik không có quyền truy cập vào mạng này.
3. **`data_net`**: Mạng đồng bộ dữ liệu riêng của các DB (Cluster Galera, Redis Sentinel). Nếu triển khai Galera cluster đa node, traffic đồng bộ data chạy qua ngõ này.

## 3. Kiến trúc Triển khai (Tier 1 & Tier 2)

Quá trình `swarm-ctl cluster init` sẽ cài đặt các thành phần theo thứ tự thiết yếu nhất:

### Tier 1: Infrastructure & Traffic Routing
* **Traefik (Gateway)**: Phân giải SSL thông qua HTTP Challenge/DNS Challenge và tự động khám phá các services triển khai vào cụm. 
* **Portainer** (Tùy chọn): Quản lý Swarm thông qua WebUI trực quan (dù chúng ta đã có `swarm-ctl dashboard`).

### Tier 2: Platform Services (Stateful)
*Lớp này phải chạy trên Manager có mount thư mục `/opt/data`.*
* **MinIO Object Storage**: Thay thế Amazon S3, để lưu hình ảnh, video của người dùng (`s3.yourdomain.com`).
* **MariaDB**: Quản trị CSDL (`mariadb_password` được mount qua Docker Secrets).
* **Redis**: Cache in-memory, session storage, Background Jobs Queue.
* **Monitoring Stack**: Prometheus đi gom rác Metric toàn cụm, Grafana lên Chart thống kê tài nguyên hệ thống, Loki hứng log của toàn bộ Container và AlertManager gửi thông báo về Telegram/Slack khi CPU > 80% hoặc Node tạch.

## 4. Bảo mật (Security Model)

* **Secret Management Zero-Trust**: Không khai báo mật khẩu trong file config `docker-compose.yml`. Mật khẩu được mã hóa và chích thẳng vào Memory (`/run/secrets/`) của container (Sách trắng Docker Swarm Secrets). Lệnh `swarm-ctl secret rotate` cho phép đổi password không sập app.
* **Ansible over SSH**: Công cụ `swarm-ctl` chỉ nói chuyện với Master node bằng giao thức SSH sử dụng Private Key. Không mở port Docker Socket ra Public.
* **Traefik Security Headers**: Các Middlewares chặn iframe, ép buộc HSTS (TLS) trọn đời cho mọi sub-domain.

## 5. Persistence Storage (Dữ liệu cố định)

Toàn bộ dữ liệu gắn liền với ổ đĩa cứng sẽ mặc định lưu tại Server chứa `role=manager` nằm tại thư mục:
```text
/opt/data/
├── appwrite/
│   ├── mariadb/
│   ├── redis/
│   ├── uploads/
├── backups/
├── minio/
├── monitoring/
└── traefik/
    └── certs/
```
Chính vì vậy, lệnh `swarm-ctl backup create` sẽ tự động zip toàn bộ thư mục `/opt/data/` (ngoại trừ DB đang chạy dở phải dùng mysqldump riêng trước). Mọi App của bạn hoàn toàn Stateless và bay tự do khắp cụm Cluster.
