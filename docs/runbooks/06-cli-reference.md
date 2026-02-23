# 📖 Bảng Tham Chiếu Lệnh Toàn Diện (CLI Reference)

Tài liệu này chứa mô tả chi tiết và **ví dụ cụ thể** cho toàn bộ lệnh của `swarm-ctl`.

---

## Mục Lục

1. [cluster — Quản lý Cluster](#1-cluster--quản-lý-cluster)
2. [node — Quản lý Node](#2-node--quản-lý-node)
3. [service — Quản lý Service](#3-service--quản-lý-service)
4. [secret — Quản lý Docker Secrets](#4-secret--quản-lý-docker-secrets)
5. [app — Triển khai App Bundle](#5-app--triển-khai-app-bundle)
6. [config — Quản lý Cấu hình](#6-config--quản-lý-cấu-hình)
7. [backup — Sao lưu & Phục hồi](#7-backup--sao-lưu--phục-hồi)
8. [dashboard — Giám sát Trực tiếp](#8-dashboard--giám-sát-trực-tiếp)
9. [audit — Nhật ký Kiểm toán](#9-audit--nhật-ký-kiểm-toán)
10. [doctor — Chẩn đoán Sự cố](#10-doctor--chẩn-đoán-sự-cố)
11. [registry — Docker Registry](#11-registry--docker-registry)
12. [update — Tự cập nhật](#12-update--tự-cập-nhật)
13. [version — Phiên bản](#13-version--phiên-bản)

---

## 1. `cluster` — Quản lý Cluster

### `cluster init` — Khởi tạo cluster mới

Tự động cài Docker, khởi tạo Swarm, deploy Traefik + Portainer.

```bash
# Cách 1: Truyền đầy đủ tham số
swarm-ctl cluster init \
  --master 192.168.1.10 \
  --domain company.com \
  --pass "RootPassword123"

# Cách 2: Chế độ tương tác (Tool tự hỏi từng thông tin)
swarm-ctl cluster init
# 👉 Nhập Master IP (vd: 1.2.3.4): 192.168.1.10
# 👉 Nhập Domain gốc (vd: example.com): company.com
# 👉 Nhập Mật khẩu Root Server: ****

# Cách 3: Tuỳ chỉnh nâng cao
swarm-ctl cluster init \
  --master 10.0.0.1 \
  --domain api.company.com \
  --key ~/.ssh/id_rsa \
  --user deploy \
  --data-root /data/swarm \
  --name staging
```

**Kết quả:**
```
╭──────────────────────────────────────────────────────╮
│ ✅ Cluster "production" đã được khởi tạo thành công!  │
│                                                      │
│   🌐 Traefik Dashboard : https://traefik.company.com │
│   📊 Portainer         : https://portainer.company.com│
╰──────────────────────────────────────────────────────╯
```

### `cluster status` — Xem trạng thái cluster

```bash
swarm-ctl cluster status
```

**Kết quả mẫu:**
```
╭───────────────────────────────────╮
│  🐳 Cluster: production          │
╰───────────────────────────────────╯
  Master: 192.168.1.10  |  Domain: company.com

  NODES
  HOSTNAME        STATUS    AVAILABILITY   MANAGER STATUS
  manager-01      Ready     Active         Leader
  worker-01       Ready     Active
  worker-02       Ready     Active

  SERVICES
  NAME                    REPLICAS   IMAGE
  traefik_traefik         1/1        traefik:v3.6.2
  portainer_portainer     1/1        portainer/portainer-ce:2.38.1
```

### `cluster upgrade` — Nâng cấp Docker Engine

```bash
# Nâng cấp Docker trên tất cả nodes (zero-downtime, lần lượt từng node)
swarm-ctl cluster upgrade
```

### `cluster destroy` — Xóa trắng cluster

```bash
# Xem cảnh báo trước
swarm-ctl cluster destroy

# Xác nhận xóa (NGUY HIỂM — mất toàn bộ dữ liệu!)
swarm-ctl cluster destroy --force
```

---

## 2. `node` — Quản lý Node

### `node add` — Thêm node vào cluster

```bash
# Thêm Worker node (phổ biến nhất)
swarm-ctl node add --ip 192.168.1.11 --role worker --pass "WorkerPass123"

# Thêm Manager node (cho HA, cần số lẻ: 3, 5, 7)
swarm-ctl node add --ip 192.168.1.12 --role manager --pass "ManagerPass123"

# Thêm node kèm label (dùng để constraint placement)
swarm-ctl node add \
  --ip 192.168.1.13 \
  --role worker \
  --pass "Pass123" \
  --label "tier=frontend,region=hcm"
```

**Kết quả:**
```
╭──────────────────────────────────────────────────────╮
│ ✅ Node 192.168.1.11 đã được thêm thành công!       │
│    Role   : worker                                   │
│    Labels : tier=frontend, region=hcm                │
╰──────────────────────────────────────────────────────╯
```

### `node list` — Xem danh sách nodes

```bash
swarm-ctl node list
# Viết tắt:
swarm-ctl node ls
```

**Kết quả:**
```
  NODES — Cluster: production
HOSTNAME          STATUS    AVAILABILITY   MANAGER STATUS   ENGINE VERSION
192.168.1.11      Ready     Active                          29.2.1
192.168.1.12      Ready     Active                          29.2.1
master.company    Ready     Active         Leader           29.2.1
```

### `node remove` — Gỡ node khỏi cluster

```bash
# Gỡ node an toàn (drain trước, chờ reschedule, rồi mới xóa)
swarm-ctl node remove --ip 192.168.1.13

# Tool sẽ yêu cầu xác nhận bằng cách nhập lại IP
# Nhập IP node để xác nhận xóa: 192.168.1.13

# Bỏ qua xác nhận (dùng trong script tự động)
swarm-ctl node remove --ip 192.168.1.13 --force
```

### `node ssh` — SSH trực tiếp vào node

```bash
# SSH vào node bất kỳ (tự dùng key từ config)
swarm-ctl node ssh 192.168.1.11
```

### `node label` — Quản lý labels

```bash
# Xem labels trên 1 node
swarm-ctl node label list --ip 192.168.1.11

# Thêm label
swarm-ctl node label add --ip 192.168.1.11 --label "gpu=true"

# Xóa label
swarm-ctl node label remove --ip 192.168.1.11 --label "gpu"
```

---

## 3. `service` — Quản lý Service

### `service add` — Triển khai service mới

```bash
# Deploy 1 website React (2 replicas, tự động SSL)
swarm-ctl service add \
  --name frontend-web \
  --image mycompany/website:v1.0.0 \
  --domain www.company.com \
  --port 80 \
  --replicas 2

# Deploy API backend (1 replica, port nội bộ 3000)
swarm-ctl service add \
  --name backend-api \
  --image mycompany/api:latest \
  --domain api.company.com \
  --port 3000 \
  --replicas 1
```

### `service list` — Xem danh sách services

```bash
swarm-ctl service list
# Viết tắt:
swarm-ctl service ls
```

**Kết quả:**
```
  SERVICES
  NAME                    REPLICAS   IMAGE                         DOMAIN
  frontend-web            2/2        mycompany/website:v1.0.0      www.company.com
  backend-api             1/1        mycompany/api:latest          api.company.com
  traefik_traefik         1/1        traefik:v3.6.2
  portainer_portainer     1/1        portainer/portainer-ce:2.38.1
```

### `service scale` — Tăng/giảm số lượng replicas

```bash
# Tăng frontend lên 4 replicas (gánh tải cao điểm)
swarm-ctl service scale frontend-web=4

# Giảm xuống 1 replica (ngoài giờ cao điểm)
swarm-ctl service scale frontend-web=1
```

### `service update` — Cập nhật image (zero-downtime)

```bash
# Cập nhật lên phiên bản mới
swarm-ctl service update frontend-web --image mycompany/website:v2.0.0

# Docker Swarm sẽ tự rolling update: khởi bản mới → tắt bản cũ → không downtime
```

### `service rollback` — Quay về phiên bản trước

```bash
# Phiên bản mới bị lỗi? Rollback ngay!
swarm-ctl service rollback frontend-web
```

### `service logs` — Xem logs

```bash
# Xem 50 dòng log gần nhất
swarm-ctl service logs frontend-web

# Xem logs theo thời gian thực (follow)
swarm-ctl service logs frontend-web --follow

# Xem N dòng log gần nhất
swarm-ctl service logs frontend-web --tail 100
```

### `service remove` — Xóa service

```bash
swarm-ctl service remove frontend-web
```

---

## 4. `secret` — Quản lý Docker Secrets

### `secret add` — Tạo secret mới

```bash
# Tạo secret chứa mật khẩu database
swarm-ctl secret add db_password "MySecretDbP@ss2026"

# Tạo secret chứa API key
swarm-ctl secret add stripe_api_key "sk_live_xxxxxxxxxxxx"
```

### `secret list` — Xem danh sách secrets

```bash
swarm-ctl secret list
```

**Kết quả:**
```
  SECRETS
  ID                          NAME              CREATED
  abc123def456                db_password       2 hours ago
  xyz789ghi012                stripe_api_key    5 minutes ago
```

### `secret rotate` — Đổi secret không downtime

```bash
# Đổi mật khẩu database mà API đang kết nối KHÔNG bị restart
swarm-ctl secret rotate db_password "NewP@ssword2026!"

# Tool sẽ:
# 1. Tạo secret mới (db_password_v2)
# 2. Cập nhật service dùng secret này
# 3. Xóa secret cũ
# → Không downtime!
```

### `secret remove` — Xóa secret

```bash
swarm-ctl secret remove stripe_api_key
```

---

## 5. `app` — Triển khai App Bundle

### `app deploy` — Deploy ứng dụng từ thư mục Bundle

```bash
# Deploy MariaDB bundle
swarm-ctl app deploy ./examples/mariadb-bundle --name mariadb_prod

# Deploy MinIO bundle
swarm-ctl app deploy ./examples/minio-bundle --name storage_s3

# Deploy bundle tùy chỉnh từ đường dẫn tuyệt đối
swarm-ctl app deploy /home/products/my-api-bundle --name backend_api

# Deploy với compose file override
swarm-ctl app deploy ./my-bundle \
  --name web_app \
  --compose-file docker-compose.prod.yml
```

**Cấu trúc App Bundle mẫu:**
```text
my-api-bundle/
├── docker-compose.yml       # Kiến trúc services
├── .env                     # Biến môi trường ($DOMAIN, $DATA_ROOT tự inject)
└── secrets/
    ├── db_root_password.txt # Swarm Secret (mount vào /run/secrets/)
    └── api_key.txt
```

### `app remove` — Gỡ bỏ ứng dụng

```bash
swarm-ctl app remove mariadb_prod
```

---

## 6. `config` — Quản lý Cấu hình

*(Xem chi tiết đầy đủ tại [05-config-management.md](05-config-management.md))*

```bash
# Xem toàn bộ cấu hình
swarm-ctl config show

# Liệt kê 23 config keys
swarm-ctl config keys

# Đổi subdomain Traefik
swarm-ctl config set traefik-subdomain dashboard

# Đổi subdomain Portainer
swarm-ctl config set portainer-subdomain admin

# Chuyển SSL sang HTTP Challenge (cho Cloudflare Proxy)
swarm-ctl config set acme-challenge http

# Đặt mật khẩu Traefik Dashboard
swarm-ctl config set traefik-auth-user operator
swarm-ctl config set traefik-auth-pass "StrongPass123"

# Cấu hình Docker Registry nội bộ
swarm-ctl config set registry-server registry.company.com:5000
swarm-ctl config set registry-user deploy-bot
swarm-ctl config set registry-pass "token_xxx"

# Cấu hình Backup S3
swarm-ctl config set backup-s3-endpoint s3.amazonaws.com
swarm-ctl config set backup-s3-bucket my-backup
swarm-ctl config set backup-s3-access-key AKIAEXAMPLE
swarm-ctl config set backup-s3-secret-key wJalrXUtnFEMI/EXAMPLEKEY

# Bật Telegram Alert
swarm-ctl config set alert-telegram-token "123456789:AAFx..."
swarm-ctl config set alert-telegram-chat "-100123456789"
swarm-ctl config set alert-telegram-enabled true

# Áp dụng thay đổi lên server
swarm-ctl config apply

# Reset về mặc định
swarm-ctl config reset traefik-subdomain

# So sánh local vs server
swarm-ctl config diff

# Xuất config
swarm-ctl config export cluster-backup.yml
swarm-ctl config export cluster-backup.json

# Nhập config trên máy mới
swarm-ctl config import cluster-backup.yml
```

---

## 7. `backup` — Sao lưu & Phục hồi

### `backup create` — Tạo bản sao lưu

```bash
# Backup toàn bộ thư mục /opt/data trên master
swarm-ctl backup create

# Backup riêng 1 service
swarm-ctl backup create --service mariadb_prod
```

### `backup list` — Xem danh sách bản sao lưu

```bash
swarm-ctl backup list
```

**Kết quả:**
```
  BACKUPS
  ID          SIZE      DATE                  TYPE
  bk-001      245MB     2026-02-23 02:00:00   full
  bk-002      12MB      2026-02-23 14:30:00   service:mariadb_prod
```

### `backup restore` — Khôi phục từ bản sao lưu

```bash
# Khôi phục từ bản backup cụ thể
swarm-ctl backup restore bk-001
```

---

## 8. `dashboard` — Giám sát Trực tiếp (TUI)

```bash
# Mở bảng điều khiển trực tiếp trên Terminal
swarm-ctl dashboard
```

**Giao diện mẫu:**
```
╭──────────────────────────────────────────────╮
│  🐳 SWARM DASHBOARD — production            │
│  Master: 192.168.1.10 | Domain: company.com │
╰──────────────────────────────────────────────╯

  📊 NODES                          CPU    MEM    DISK
  ● manager-01 (Leader)             23%    4.2G   45%
  ● worker-01                       67%    3.1G   32%
  ● worker-02                       12%    1.8G   28%

  🔧 SERVICES                       REPLICAS   STATUS
  traefik_traefik                    1/1        ✅ Running
  portainer_portainer                1/1        ✅ Running
  frontend-web                       2/2        ✅ Running
  backend-api                        1/1        ✅ Running

  Nhấn 'q' để thoát | 'r' để refresh | '↑↓' để cuộn
```

---

## 9. `audit` — Nhật ký Kiểm toán

```bash
# Xem 30 dòng log gần nhất (mặc định)
swarm-ctl audit

# Xem 100 dòng log gần nhất
swarm-ctl audit --tail 100
```

**Kết quả mẫu:**
```
  📝 AUDIT LOG (~/.swarm-ctl/audit.log)

  2026-02-23 19:06:09  cluster init --master 192.168.1.10 --domain company.com --pass ****
  2026-02-23 19:12:33  node add --ip 192.168.1.11 --role worker --pass ****
  2026-02-23 19:13:45  node add --ip 192.168.1.12 --role worker --pass ****
  2026-02-23 19:15:00  config set traefik-subdomain dashboard
  2026-02-23 19:15:05  config set portainer-subdomain admin
  2026-02-23 19:15:10  config apply
  2026-02-23 19:20:00  app deploy ./examples/mariadb-bundle --name mariadb_prod
  2026-02-23 19:25:00  service scale frontend-web=4
```

> Lưu ý: Mật khẩu và secret luôn được thay thế bằng `****` trong audit log.

---

## 10. `doctor` — Chẩn đoán Sự cố

```bash
# Kiểm tra toàn bộ sức khỏe cluster
swarm-ctl doctor
```

**Kết quả mẫu:**
```
  🏥 CLUSTER HEALTH CHECK

  ✅ SSH kết nối Master            OK
  ✅ Docker Engine                 v29.2.1
  ✅ Swarm Status                  Active (Leader)
  ✅ Nodes Online                  3/3
  ⚠️  Disk Usage Master            78% (cảnh báo > 80%)
  ✅ Traefik                       Running (1/1)
  ✅ Portainer                     Running (1/1)
  ✅ SSL Certificate               Valid (còn 67 ngày)
  ❌ Telegram Alert                Chưa cấu hình

  Tổng: 8 ✅  1 ⚠️  1 ❌
```

---

## 11. `registry` — Docker Registry

```bash
# Đăng nhập Docker Hub
swarm-ctl registry login --user mycompany --pass "dckr_pat_xxx"

# Đăng nhập registry nội bộ
swarm-ctl registry login \
  --server registry.company.com:5000 \
  --user deploy-bot \
  --pass "token_xxx"
```

---

## 12. `update` — Tự cập nhật swarm-ctl

```bash
# Cập nhật lên phiên bản mới nhất từ GitHub
swarm-ctl update
```

**Kết quả:**
```
  🔄 Đang kiểm tra phiên bản mới...
  Phiên bản hiện tại: v1.0.38
  Phiên bản mới nhất: v1.0.40

  ✅ Đã cập nhật swarm-ctl lên v1.0.40
```

---

## 13. `version` — Xem phiên bản

```bash
swarm-ctl version
```

**Kết quả:**
```
  swarm-ctl v1.0.38 (built: 2026-02-23)
```

---

## Cờ Toàn Cục (Global Flags)

Các cờ này áp dụng cho **tất cả** lệnh:

| Cờ | Mô tả | Ví dụ |
|---|---|---|
| `--verbose` / `-v` | In log chi tiết (debug) | `swarm-ctl cluster init -v` |
| `--config` | Đường dẫn file config tùy chỉnh | `swarm-ctl --config /path/to/config.yml cluster status` |
| `--no-color` | Tắt màu sắc (dùng trong CI/CD pipeline) | `swarm-ctl --no-color service list` |
| `--help` / `-h` | Xem trợ giúp cho bất kỳ lệnh nào | `swarm-ctl node add --help` |

---

## Kịch Bản Thực Tế (End-to-End)

### Kịch bản: Setup cluster 3 máy + deploy website + giám sát

```bash
# 1. Khởi tạo cluster trên Master
swarm-ctl cluster init --master 192.168.1.10 --domain company.com --pass "RootPass"

# 2. Thêm 2 Worker
swarm-ctl node add --ip 192.168.1.11 --role worker --pass "Worker1Pass"
swarm-ctl node add --ip 192.168.1.12 --role worker --pass "Worker2Pass"

# 3. Kiểm tra cluster
swarm-ctl node list

# 4. Cấu hình subdomain & SSL
swarm-ctl config set traefik-subdomain dashboard
swarm-ctl config set acme-email devops@company.com
swarm-ctl config apply

# 5. Deploy database
swarm-ctl app deploy ./bundles/mariadb-bundle --name database

# 6. Deploy website (2 replicas cho HA)
swarm-ctl service add \
  --name website \
  --image company/web:v1.0 \
  --domain www.company.com \
  --port 80 \
  --replicas 2

# 7. Cấu hình Telegram Alert
swarm-ctl config set alert-telegram-token "123456:AAFx..."
swarm-ctl config set alert-telegram-chat "-10012345"
swarm-ctl config set alert-telegram-enabled true

# 8. Backup config cho đồng nghiệp
swarm-ctl config export cluster-prod.yml

# 9. Giám sát trực tiếp
swarm-ctl dashboard
```
