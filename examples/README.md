# Swarm-ctl Starter Bundles (Application Templates)

Chào mừng bạn đến với kho lưu trữ Template Ứng Dụng dành cho **Docker Swarm 3.0**!

Theo thiết kế mới, hệ thống `swarm-ctl init` đã dọn dẹp sạch sẽ và chỉ lưu giữ Lớp Hạ tầng Mạng (Traefik và Portainer). Mọi ứng dụng khác như Database, Storage, Monitoring sẽ được trả về hoàn toàn cho Người dùng quản lý thông qua cơ chế **Thư Mục Bundle (Bundle Folders)**.

## Cơ Chế Hoạt Động Của Bundle
Thay vì viết một file `docker-compose.yml` quá khổ và gắn cứng mật khẩu vào text, Bundle chia Application của bạn ra làm 3 phần sạch sẽ:
1. `docker-compose.yml`: Kiến trúc ứng dụng, không chứa file cấu hình hay secret cứng.
2. `configs/`: Thư mục chứa các file config text cần nạp đè vào App (như `nginx.conf`, `prometheus.yml`). Swarm sẽ tự map file.
3. `secrets/`: Thư mục chứa file text các mật khẩu. Nó không bao giờ bị lộ ra log, và chỉ lưu vào RAM của Container đích.

### 🌟 Cách Sử Dụng Cực Nhanh (1 Phút Deploy)
Bạn muốn dựng Database MariaDB? 
1. Chép thư mục `mariadb-bundle` về thành dự án của bạn:
   ```bash
   cp -r examples/mariadb-bundle /home/products/services/my-mariadb
   ```
2. Vào sửa file `/home/products/services/my-mariadb/.env` theo ý thích (đặt pass mới).
3. Đứng từ máy cá nhân, Gõ 1 lệnh thần thánh:
   ```bash
   swarm-ctl app deploy /home/products/services/my-mariadb --name mariadb_prod
   ```

Tool `swarm-ctl` sẽ nén toàn bộ thư mục đó (thay vì bắt bạn copy lắt nhắt), bắn bằng mã hóa qua máy chủ Master, và Docker Swarm sẽ tiếp quản, nặn Mật khẩu trong folder `secrets` của bạn thành Swarm-Secrets, rồi dựng hệ thống siêu cấp an toàn.

Chúc bạn thiết kế và làm chủ vương quốc Microservices của riêng mình!
