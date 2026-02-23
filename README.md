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
# 1. Khởi tạo Kẻ Gác Cổng (Traefik Router + Portainer)
swarm-ctl cluster init --master 10.0.0.1 --key ~/.ssh/id_rsa --domain company.com

# 2. Triển khai Ứng dụng Bất Kỳ thông qua App Bundle (Thư mục YML + Secrets)
swarm-ctl app deploy ./my-mariadb-bundle --name mariadb_prod

# 3. Theo dõi toàn cảnh Sức khỏe Cluster thông qua TUI màu sắc
swarm-ctl dashboard
```

---

## 💻 Yêu cầu Hệ thống (System Requirements)

Để chạy trơn tru kiến trúc Cluster chuẩn, bạn cần thuê các Server/VPS **"Trắng tinh"** (Chưa cài đặt gì cả ngoài Hệ Điều Hành là **Ubuntu 20.04+** hoặc **Debian 11+** có mở SSH Port 22).

🚨 **ĐẶC BIỆT LƯU Ý:** BẠN KHÔNG CẦN (VÀ KHÔNG NÊN) TỰ CÀI DOCKER HAY DOCKER SWARM. Chỉ cần cấp mật khẩu Root để Tool đăng nhập cắm cờ vào, `swarm-ctl` sẽ tự động tải các gói Docker, tự động setup Mạng Ảo, tự chỉ định Leader hoàn toàn 100%. Mọi thứ cứ để công cụ tự lo!

### 1. Master/Manager Node (Khuyên dùng tối thiểu 1 máy)
Máy chứa số liệu cột lõi và điều khiển Swarm.
- **CPU:** Tối thiểu 2 vCores (Khuyến nghị 4 vCores cho production).
- **RAM:** Tối thiểu 4GB (Khuyến nghị 8GB+ cho production).
- **Disk:** Tối thiểu 50GB SSD/NVMe.

### 2. Worker Nodes (Tuỳ chọn ráp sau)
Dành cho ứng dụng Backend, Web, API để phân tán tải.
- **CPU:** 1-2 vCores / **RAM:** 2GB.
*(Tất cả kết nối qua LAN Nội bộ Private IP để giảm độ trễ)*

### 3. Máy cá nhân chạy lệnh (`swarm-ctl`)
- Linux / macOS (Chỉ hỗ trợ môi trường Unix-like).
- Tool yêu cầu bộ điều khiển phải có Khóa mã hóa SSH để xâm nhập các Server qua ngầm. Tuy nhiên, **BẠN KHÔNG CẦN LÀM GÌ CẢ**. Nếu máy tính bạn chưa có Khóa, công cụ `swarm-ctl` sẽ tự động sinh ra một chiếc chìa khóa `Ed25519` (chuẩn bảo mật quân đội) đặt tại đường dẫn `~/.ssh/id_ed25519` cho bạn.

---

## 📦 Cài đặt Công Cụ (`swarm-ctl`)

### Cách 1: Tải nhanh Binary (Khuyến nghị)
**Dành cho Linux (Ubuntu/Debian/CentOS)**
```bash
sudo curl -L https://github.com/LyVanBong/swarm-ctl/releases/latest/download/swarm-ctl-linux-amd64 \
  -o /usr/local/bin/swarm-ctl && sudo chmod +x /usr/local/bin/swarm-ctl
```

**Dành cho MacOS Apple Silicon (Chip M1/M2/M3)**
```bash
sudo curl -L https://github.com/LyVanBong/swarm-ctl/releases/latest/download/swarm-ctl-darwin-arm64 \
  -o /usr/local/bin/swarm-ctl && sudo chmod +x /usr/local/bin/swarm-ctl
```

**Dành cho MacOS Intel (Chip core i5/i7/i9)**
```bash
sudo curl -L https://github.com/LyVanBong/swarm-ctl/releases/latest/download/swarm-ctl-darwin-amd64 \
  -o /usr/local/bin/swarm-ctl && sudo chmod +x /usr/local/bin/swarm-ctl
```

Kiểm tra cài đặt:
```bash
swarm-ctl version
```

### Cách 2: Build từ Bản đồ Code (Source Code)
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

1. **Tier 1 (Hạ tầng Lõi - Infrastructure):**
   - **Traefik Proxy:** Cổng Router, Auto-HTTPS (Let's Encrypt), Rate-limit, Security Headers Enterprise.
   - **Portainer GUI:** Giao diện điều khiển Trực quan.
2. **Tier 2 (Nền tảng Tự do - App Bundle Pipeline):**
   - Tool `swarm-ctl` nay đã **trả lại 100% quyền kiểm soát Ứng dụng** cho Nhà Quản Trị. 
   - Từ Database (MariaDB, PostgreSQL) đến Lưu trữ (Minio) hay Dịch vụ Kênh Chat. Administrator chỉ cần tự thiết kế các Thư Mục Cấu hình (App Bundle) và vứt cho lệnh `swarm-ctl app deploy` để Engine tự đóng gói ném thẳng lên Server. Mọi thứ trong suốt và phi trạng thái!

---

## ✨ Tính Năng Nổi Bật (Ecosystem)

- 🔒 **Zero-Trust Connection:** Không mở cổng Docker Socket ra Internet. Tool hoàn toàn ra lệnh máy chủ thông qua kênh SSH mã hóa nội bộ.
- 🛡️ **Auto-Firewall (UFW):** Khi kết nạp Server, Tool tự đóng sập toàn bộ các Port truy cập trái phép của Hacker, chỉ chừa khe hở cho Web & SSH giao tiếp.
- ♻️ **Zero-Downtime Deploy:** Hỗ trợ `service update` (khởi chạy bản mới trước khi tắt bản cũ) hoặc `secret rotate` (đổi mật khẩu không downtime).
- ⚙️ **Config Management 23 Keys:** Lệnh `swarm-ctl config` quản lý toàn bộ cấu hình: Domain, SSL, Auth, Registry, Backup S3, Telegram Alert.
- 🗄️ **Backup & Restore:** Lệnh `backup create/restore` tự động zip toàn thư mục phân quyền.
- 📝 **Audit Trail:** Mọi lệnh can thiệp đều ghi nhật ký tên PC, thời gian vào file local `/audit.log`.
- 📊 **Giám Sát Trực Tiếp (TUI):** Gõ `swarm-ctl dashboard` để mở Bảng điện tử Live trên Terminal, thời gian thực, không cần trình duyệt.

---

## 📂 Mô Hình Triển Khai App Bundle (Kiến trúc Thay Thế Marketplace)

Thay vì bó buộc bạn vào một vài Ứng dụng "Cứng ngắc", `swarm-ctl v3.0` trình làng phương pháp **App Bundle Deployment**. Bạn đóng gói sẵn mã nguồn/file thiết kế vào một Thư mục máy cá nhân, Tool sẽ tự nén, mã hóa, vận chuyển qua SSH và dựng thành Hệ thống trên Cụm.

Cấu trúc Bundle Chuẩn Mực của một Ứng Dụng:
```text
my-minio-bundle/
├── docker-compose.yml       # Khung xương kiến trúc
├── .env                     # Khai báo các biến tự động ($DOMAIN)
└── secrets/
    ├── minio_root_user.txt  # Dữ liệu nhạy cảm được nhét cực kín vào Swarm-Secrets (KHÔNG lưu txt lên RAM public)
    └── minio_root_pass.txt
```

Deploy lên Production đơn giản bằng 1 nốt nhạc:
```bash
swarm-ctl app deploy /home/products/my-minio-bundle --name minio_storage
```

Để xem các Ví dụ kinh điển (Starter-Kit), vui lòng tham khảo [examples/README.md](examples/README.md)!

---

## 📌 Giải Thích Các Tham Số Đặc Biệt (Dành Cho Người Mới)

Khi chạy các lệnh của `swarm-ctl`, có một vài tham số bắt buộc. Dưới đây là giải thích chi tiết chúng là gì và lấy từ đâu:

1. **`--master` hoặc `--ip` (Địa chỉ IPv4 của Server)**
   * **Nó là gì:** Địa chỉ mạng IP Public hoặc Private của máy chủ Cloud/VPS bạn mua.
   * **Lấy ở đâu:** Đăng nhập vào bảng điều khiển (Dashboard) của nhà cung cấp Server (AWS EC2, DigitalOcean Droplet, Linode) và copy địa chỉ Public IP (vd: `103.25.x.x`).

2. **`--key` (Đường dẫn tới SSH Private Key)**
   * **Nó là gì:** Thay vì nhập mật khẩu dài ngoằng để kết nối Server, Linux dùng tệp tin "Chìa khóa mã hóa" (Thường tool sẽ ưu tiên tự sinh loại `id_ed25519` vì nó tân tiến và an toàn nhất hiện nay, thay thế cho chuẩn `id_rsa` cũ kĩ).
   * **Lấy ở đâu:** Công cụ tự sinh cho bạn và giấu ở thư mục: `~/.ssh/id_ed25519`

3. **`--pass` (Tự động Copy Khóa SSH)**
   * **Nó là gì:** Nếu bạn thuê Máy ảo mới toanh, bạn thường chỉ nhận được `Root Password`. Nếu không truyền cờ `--pass`, bạn phải cài SSH thủ công bằng lệnh `ssh-copy-id`. Với cờ `--pass "MậtKhẩu"`, Tool sẽ thay bạn làm công việc cực nhọc đó một cách hoàn toàn tự động!
   * **Lấy ở đâu:** Email cấp máy chủ từ nhà cung cấp Cloud.

4. **`--domain` (Tên miền chính thức)**
   * **Nó là gì:** Tên website bạn dùng để dẫn khách truy cập vào Server (vd: `congtycuaban.com` hay `api.congty.vn`).
   * **Lấy ở đâu:** Bạn phải Đăng ký mua trên các trang bán Domain (Tenten, Namecheap, Cloudflare). Sau khi mua, bạn vào cấu hình DNS của Domain đó, tạo Record **A** (bao gồm cả wildcard `*.congtycuaban.com`) và trỏ IP về đúng số `--master IP` bên trên. Khi đó Traefik mới tự cấp chứng chỉ HTTPS cho bạn được.

4. **`--role` (Vai trò của Node)**
   * **Nó là gì:** Khi cắm thêm Máy ảo Máy con vào mạng lưới, bạn phân công nó làm Nhiệm vụ gì.
   * **Sử dụng:** `manager` (Vừa chạy App vừa làm quản trị Database/Bầu cử) hoặc `worker` (Chỉ chuyên chạy App, chịu tải Web).

---

## 📋 Bảng Lệnh Tham Khảo (CLI Commands)

```text
swarm-ctl
├── cluster
│   ├── init     --master IP --domain DOMAIN [--pass PASSWORD]
│   ├── status
│   ├── upgrade
│   └── destroy  [--force]
├── node
│   ├── add      --ip IP [--role worker|manager] [--pass PASSWORD]
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
├── config
│   ├── show                          # Xem toàn bộ cấu hình
│   ├── keys                          # Liệt kê 23 config keys
│   ├── set      KEY VALUE             # Cập nhật giá trị
│   ├── reset    KEY                   # Đặt lại về mặc định
│   ├── apply                          # Áp dụng lên server đang chạy
│   ├── export   [FILE]                # Xuất ra YAML/JSON
│   ├── import   FILE                  # Nhập từ file
│   └── diff                           # So sánh local vs server
├── backup
│   ├── create
│   ├── restore  BACKUP-ID
│   └── list
├── app
│   └── deploy   THU_MUC_BUNDLE [--name SERVICE_NAME]
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
* [Triển khai Thực chiến (End-to-end Tutorial)](docs/runbooks/04-end-to-end-tutorial.md) - ⭐️ BÀI TẬP VÍ DỤ: Cài đặt cụm 3 Máy, Phân tải Backend API và Website!
* [⚙️ Quản Lý Cấu Hình (Config Management)](docs/runbooks/05-config-management.md) - 23 Config Keys, Export/Import, Diff, SSL, Auth, Backup S3, Telegram Alert.
* [📖 Bảng Tham Chiếu Lệnh (CLI Reference)](docs/runbooks/06-cli-reference.md) - Ví dụ cụ thể cho toàn bộ lệnh kèm output mẫu.
* [Kiến trúc Mô Hình Cluster](docs/architecture.md)
* [Lịch sử Cập Nhật (CHANGELOG)](CHANGELOG.md)

---

## 🤝 Đóng góp Mã Nguồn (Contributing)

Dự án là của Cộng đồng! Bạn có ý tưởng phát triển tính năng mới, hay Report lỗi, xin vui lòng xem [Hướng dẫn Đóng góp (CONTRIBUTING)](CONTRIBUTING.md). Cảm ơn vì đã giúp `swarm-ctl` trở nên tốt hơn!

---

## 📄 Bản Quyền (License)

Dự án được phân phối dưới giấy phép [MIT License](LICENSE). 

<div align="center">
<b>Được phát triển bền vững bởi <a href="https://github.com/LyVanBong">Ly Van Bong</a> & <a href="https://www.softty.net">Softty Net</a></b>
</div>
