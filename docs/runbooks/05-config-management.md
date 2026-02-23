# Hướng Dẫn Quản Lý Cấu Hình (`swarm-ctl config`)

## Tổng Quan

Hệ thống Config của `swarm-ctl` cho phép bạn quản lý toàn bộ cấu hình cluster từ một nơi duy nhất, bao gồm: tên miền, SSL, xác thực, Docker Registry, Backup S3 và Thông báo Telegram.

Mọi cấu hình được lưu tại: `~/.swarm-ctl/config.yml`

---

## Danh Sách Lệnh

| Lệnh | Chức năng |
|---|---|
| `config show` | Xem toàn bộ cấu hình hiện tại (chia theo 7 nhóm) |
| `config keys` | Liệt kê tất cả config keys kèm mô tả |
| `config set <key> <value>` | Thay đổi một giá trị cấu hình |
| `config reset <key>` | Đặt lại giá trị về mặc định |
| `config apply` | Áp dụng cấu hình mới lên server đang chạy |
| `config export [file]` | Xuất cấu hình ra file YAML hoặc JSON |
| `config import <file>` | Nhập cấu hình từ file đã xuất |
| `config diff` | So sánh cấu hình local với server đang chạy |

---

## Ví Dụ Sử Dụng Cụ Thể

### 📌 Ví dụ 1: Xem toàn bộ cấu hình hiện tại

```bash
$ swarm-ctl config show
```

Kết quả mẫu:
```
╭──────────────────────╮
│  ⚙️  CLUSTER CONFIG  │
╰──────────────────────╯

  THÔNG TIN CHUNG
  Cluster Name : production
  Master IP    : 81.17.101.123
  Domain       : company.com
  SSH User     : root
  SSH Key      : /root/.ssh/id_ed25519
  Data Root    : /opt/data

  SUBDOMAINS
  🌐 Traefik   : traefik.company.com
  📊 Portainer : portainer.company.com

  SSL / ACME
  Email        : admin@company.com
  Challenge    : tls

  TRAEFIK AUTH
  Username     : admin
  Password     : (không đặt)

  DOCKER REGISTRY
  Server       : docker.io
  Username     : (không đặt)

  BACKUP S3
  Endpoint     : (không đặt)
  Bucket       : (không đặt)
  Region       : us-east-1
  Schedule     : 0 2 * * *

  TELEGRAM ALERT
  Trạng thái   : ❌ Tắt
  Chat ID      : (không đặt)
  Bot Token    : (không đặt)
```

### 📌 Ví dụ 2: Liệt kê tất cả key cấu hình

```bash
$ swarm-ctl config keys
```

Kết quả mẫu:
```
╭──────────────────╮
│  🔑 CONFIG KEYS  │
╰──────────────────╯

  🏗️  Cluster
  domain                       Domain chính của cluster
  cluster-name                 Tên hiển thị cluster
  data-root                    Thư mục gốc lưu dữ liệu trên server
  ssh-user                     SSH username
  ssh-key                      Đường dẫn SSH private key

  🌐 Subdomains
  traefik-subdomain            Subdomain cho Traefik Dashboard
  portainer-subdomain          Subdomain cho Portainer

  🔒 SSL/ACME
  acme-email                   Email đăng ký SSL Let's Encrypt
  acme-challenge               Kiểu xác thực SSL: tls | http
  ...
```

### 📌 Ví dụ 3: Đổi subdomain Traefik & Portainer rồi áp dụng lên server

**Tình huống:** Bạn muốn truy cập Traefik Dashboard qua `dashboard.company.com` thay vì `traefik.company.com`, và Portainer qua `admin.company.com`.

```bash
# Bước 1: Cập nhật subdomain
$ swarm-ctl config set traefik-subdomain dashboard
✅ [traefik-subdomain]: traefik → dashboard
⚠️  Chạy 'swarm-ctl config apply' để áp dụng thay đổi lên server đang chạy

$ swarm-ctl config set portainer-subdomain admin
✅ [portainer-subdomain]: portainer → admin
⚠️  Chạy 'swarm-ctl config apply' để áp dụng thay đổi lên server đang chạy

# Bước 2: Áp dụng lên server đang chạy
$ swarm-ctl config apply
╭──────────────────╮
│  � CONFIG APPLY │
╰──────────────────╯

  🌐 Traefik   → dashboard.company.com
  📊 Portainer → admin.company.com

[1/3] Cập nhật routing Traefik Dashboard...
✅ Traefik → https://dashboard.company.com
[2/3] Cập nhật routing Portainer...
✅ Portainer → https://admin.company.com
[3/3] Xác minh services...

╭──────────────────────────────────────────────╮
│ ✅ Cấu hình đã được áp dụng thành công!      │
│                                              │
│ 🌐 Traefik Dashboard : https://dashboard.company.com │
│ 📊 Portainer         : https://admin.company.com     │
│                                              │
│ Lưu ý: Hãy đảm bảo DNS đã trỏ subdomain    │
│ mới về IP: 81.17.101.123                     │
╰──────────────────────────────────────────────╯
```

> **Quan trọng:** Sau khi đổi subdomain, bạn cần vào Cloudflare/Tenten tạo bản ghi DNS mới:
> - Loại `A` | Host `dashboard` | Content `81.17.101.123`
> - Loại `A` | Host `admin` | Content `81.17.101.123`

### 📌 Ví dụ 4: Cấu hình SSL cho môi trường Cloudflare Proxy

**Tình huống:** Bạn bật "Đám mây cam" (Proxy) trên Cloudflare, Traefik không xin được SSL vì TLS Challenge bị Cloudflare chặn.

```bash
# Chuyển sang HTTP Challenge — tương thích Cloudflare Proxy
$ swarm-ctl config set acme-challenge http
✅ [acme-challenge]: tls → http

# Đổi email SSL cho dễ quản lý
$ swarm-ctl config set acme-email devops@company.com
✅ [acme-email]:  → devops@company.com
```

> **Giải thích:**
> - `tls` (mặc định): Let's Encrypt kết nối trực tiếp tới cổng 443 của Server để xác thực. **Bắt buộc DNS trỏ thẳng** (không Proxy).
> - `http`: Let's Encrypt gọi HTTP trên cổng 80 để xác thực. **Hoạt động cả khi có Cloudflare Proxy**.

### 📌 Ví dụ 5: Đặt mật khẩu bảo vệ Traefik Dashboard

**Tình huống:** Traefik Dashboard đang mở công cộng, ai biết URL cũng vào được. Bạn muốn đặt Username/Password.

```bash
$ swarm-ctl config set traefik-auth-user operator
✅ [traefik-auth-user]: admin → operator

$ swarm-ctl config set traefik-auth-pass "MyStr0ngP@ss2026!"
✅ [traefik-auth-pass]: ******** → MyStr0ngP@ss2026!
```

### � Ví dụ 6: Cấu hình Docker Registry nội bộ cho app deploy

**Tình huống:** Công ty bạn có Docker Registry riêng tại `registry.company.com:5000`. Khi chạy `swarm-ctl app deploy`, cluster cần biết thông tin đăng nhập.

```bash
$ swarm-ctl config set registry-server registry.company.com:5000
✅ [registry-server]: docker.io → registry.company.com:5000

$ swarm-ctl config set registry-user deploy-bot
✅ [registry-user]:  → deploy-bot

$ swarm-ctl config set registry-pass "ghp_xxxxxxxxxxxx"
✅ [registry-pass]: ******** → ghp_xxxxxxxxxxxx
```

### � Ví dụ 7: Cấu hình Backup tự động lên AWS S3

**Tình huống:** Bạn muốn cluster tự động backup dữ liệu `/opt/data` lên Amazon S3 lúc 4h sáng mỗi ngày.

```bash
$ swarm-ctl config set backup-s3-endpoint s3.amazonaws.com
$ swarm-ctl config set backup-s3-bucket my-cluster-backup-prod
$ swarm-ctl config set backup-s3-access-key AKIAIOSFODNN7EXAMPLE
$ swarm-ctl config set backup-s3-secret-key wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY
$ swarm-ctl config set backup-s3-region ap-southeast-1
$ swarm-ctl config set backup-s3-schedule "0 4 * * *"
```

Hoặc nếu dùng MinIO tự dựng thay vì AWS:
```bash
$ swarm-ctl config set backup-s3-endpoint https://s3.company.com
$ swarm-ctl config set backup-s3-bucket backups
$ swarm-ctl config set backup-s3-access-key minioadmin
$ swarm-ctl config set backup-s3-secret-key minioadmin123
```

### � Ví dụ 8: Bật thông báo Telegram khi Container sập

**Tình huống:** Bạn muốn nhận tin nhắn Telegram khi có service bị Down.

**Bước 1: Tạo Bot Telegram**
1. Mở Telegram → Tìm `@BotFather` → Gõ `/newbot`
2. Đặt tên bot (vd: `Cluster Alert Bot`) → Nhận Token: `123456789:AAFx...`

**Bước 2: Lấy Chat ID**
1. Tạo nhóm Telegram tên `DevOps Alert`, mời Bot vào nhóm.
2. Gửi 1 tin nhắn bất kỳ vào nhóm.
3. Mở trình duyệt, truy cập: `https://api.telegram.org/bot123456789:AAFx.../getUpdates`
4. Tìm trường `"chat":{"id":-100123456789}` → Đó là Chat ID.

**Bước 3: Cấu hình swarm-ctl**
```bash
$ swarm-ctl config set alert-telegram-token "123456789:AAFx..."
✅ [alert-telegram-token]: ******** → 123456789:AAFx...

$ swarm-ctl config set alert-telegram-chat "-100123456789"
✅ [alert-telegram-chat]:  → -100123456789

$ swarm-ctl config set alert-telegram-enabled true
✅ [alert-telegram-enabled]: false → true
```

### 📌 Ví dụ 9: Export config backup chia sẻ cho đồng nghiệp

```bash
# Xuất ra file YAML
$ swarm-ctl config export cluster-prod.yml
✅ Config đã được xuất ra: cluster-prod.yml

# Xuất ra file JSON (nếu cần tích hợp API)
$ swarm-ctl config export cluster-prod.json
✅ Config đã được xuất ra: cluster-prod.json
```

Nội dung file `cluster-prod.yml`:
```yaml
name: production
master_ip: 81.17.101.123
ssh_key: /root/.ssh/id_ed25519
ssh_user: root
domain: company.com
data_root: /opt/data
subdomains:
  traefik: dashboard
  portainer: admin
acme:
  email: devops@company.com
  challenge: http
```

### 📌 Ví dụ 10: Import config trên máy quản trị mới

**Tình huống:** Đồng nghiệp mới gia nhập team, cần setup máy quản trị nhanh.

```bash
# Trên máy đồng nghiệp (đã cài swarm-ctl)
$ swarm-ctl config import cluster-prod.yml
✅ Đã import config cluster 'production' từ cluster-prod.yml
ℹ️  Chạy 'swarm-ctl config show' để xem chi tiết

# Kiểm tra ngay
$ swarm-ctl config show
$ swarm-ctl cluster status
```

### 📌 Ví dụ 11: Kiểm tra cấu hình đồng bộ (Drift Detection)

**Tình huống:** Bạn nghi ngờ ai đó đã sửa config trên server bằng tay mà không thông qua tool.

```bash
$ swarm-ctl config diff
╭──────────────────╮
│  🔍 CONFIG DIFF  │
╰──────────────────╯

  TRAEFIK DASHBOARD
  Server : traefik.company.com
  Local  : dashboard.company.com
  ⚠️  KHÁC BIỆT — Chạy 'config apply' để đồng bộ

  PORTAINER
  ✅ Đồng bộ: admin.company.com

  VERSIONS
  Portainer : portainer/portainer-ce:2.38.1
  Traefik   : traefik:v3.6.2
```

Nếu phát hiện drift:
```bash
$ swarm-ctl config apply   # Đồng bộ lại từ config local lên server
```

### 📌 Ví dụ 12: Reset giá trị về mặc định

```bash
# Đặt lại subdomain Traefik về mặc định (traefik.{domain})
$ swarm-ctl config reset traefik-subdomain
✅ [traefik-subdomain] đã được đặt lại về giá trị mặc định

# Đặt lại kiểu xác thực SSL về mặc định (tls)
$ swarm-ctl config reset acme-challenge
✅ [acme-challenge] đã được đặt lại về giá trị mặc định

# Áp dụng lại
$ swarm-ctl config apply
```

---

## Danh Sách Đầy Đủ 23 Config Keys

### 🏗️ Cluster

| Key | Mặc định | Mô tả |
|---|---|---|
| `domain` | *(bắt buộc khi init)* | Domain chính của cluster |
| `cluster-name` | `production` | Tên hiển thị cluster |
| `data-root` | `/opt/data` | Thư mục gốc lưu dữ liệu trên server |
| `ssh-user` | `root` | SSH username kết nối server |
| `ssh-key` | `~/.ssh/id_ed25519` | Đường dẫn SSH private key |

### 🌐 Subdomains

| Key | Mặc định | Mô tả |
|---|---|---|
| `traefik-subdomain` | `traefik` | Subdomain Traefik → `traefik.{domain}` |
| `portainer-subdomain` | `portainer` | Subdomain Portainer → `portainer.{domain}` |

### 🔒 SSL / ACME

| Key | Mặc định | Mô tả |
|---|---|---|
| `acme-email` | `admin@{domain}` | Email đăng ký SSL Let's Encrypt |
| `acme-challenge` | `tls` | Kiểu xác thực: `tls` (trực tiếp) hoặc `http` (qua Cloudflare) |

### 🛡️ Traefik Auth

| Key | Mặc định | Mô tả |
|---|---|---|
| `traefik-auth-user` | `admin` | Username BasicAuth Traefik Dashboard |
| `traefik-auth-pass` | *(không đặt)* | Password BasicAuth Traefik Dashboard |

### 📦 Docker Registry

| Key | Mặc định | Mô tả |
|---|---|---|
| `registry-server` | `docker.io` | Địa chỉ Docker Registry |
| `registry-user` | *(không đặt)* | Username đăng nhập Registry |
| `registry-pass` | *(không đặt)* | Password đăng nhập Registry |

### 💾 Backup S3

| Key | Mặc định | Mô tả |
|---|---|---|
| `backup-s3-endpoint` | *(không đặt)* | S3 endpoint |
| `backup-s3-bucket` | *(không đặt)* | Tên bucket |
| `backup-s3-access-key` | *(không đặt)* | S3 Access Key |
| `backup-s3-secret-key` | *(không đặt)* | S3 Secret Key |
| `backup-s3-region` | `us-east-1` | S3 Region |
| `backup-s3-schedule` | `0 2 * * *` | Lịch Cron (mặc định: 2h sáng) |

### 🔔 Telegram Alert

| Key | Mặc định | Mô tả |
|---|---|---|
| `alert-telegram-token` | *(không đặt)* | Bot Token (từ @BotFather) |
| `alert-telegram-chat` | *(không đặt)* | Chat ID nhóm hoặc cá nhân |
| `alert-telegram-enabled` | `false` | Bật/tắt: `true` / `false` |

---

## Lưu Ý Bảo Mật

- **Mật khẩu luôn được ẩn:** Khi chạy `config show`, các trường nhạy cảm chỉ hiện `********`.
- **File config lưu local:** Toàn bộ cấu hình nằm tại `~/.swarm-ctl/config.yml` trên máy quản trị, **không** gửi lên server hay Internet.
- **Export cẩn thận:** File export có chứa thông tin nhạy cảm → Không commit vào Git public. Nên mã hóa trước khi gửi qua mạng cho đồng nghiệp.
