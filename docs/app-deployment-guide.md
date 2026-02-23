# Hướng dẫn Chuẩn hóa Thư mục Triển khai Ứng dụng (App Bundle) cho swarm-ctl 3.0

Theo thiết kế kiến trúc chuẩn của `swarm-ctl 3.0`, toàn bộ ứng dụng (Application Tier) được quản lý dưới dạng **App Bundle** (Gói Thư Mục Phi Trạng Thái).

Thay vì phải chạy lệnh Ansible phân tán, bạn chỉ cần gộp toàn bộ những gì thuộc về cấu hình của một ứng dụng vào một thư mục duy nhất trên máy tính cá nhân (Client). Docker Swarm sẽ tự lo phần còn lại.

## 1. Cấu trúc Thư mục Chuẩn (Standard Folder Structure)

Một Bundle chuẩn mực khi dùng lệnh `swarm-ctl app deploy <folder>` nên theo định dạng như sau:

```text
/home/products/services/my-awesome-app/   <-- Tên Bundle (Folder Name)
 ├── docker-compose.yml       <-- (BẮT BUỘC) Mô tả kiến trúc Stack chính
 ├── docker-compose.prod.yml  <-- (Tùy chọn) Khái niệm Override: Dùng lưu cấu hình đè cực đại khi deploy lên cụm Production
 ├── .env                     <-- Biến môi trường chung nạp vào Compose Context (VD: ROOT_PASSWORD=abc)
 │
 ├── configs/                 <-- Bơm File cấu hình thuần Text, không mã hóa (Docker Swarm Native Configs)
 │    ├── prometheus.yml
 │    ├── nginx.conf
 │    └── ...
 │
 ├── secrets/                 <-- Bơm File chứa dữ liệu cực kỳ nhạy cảm, mã hóa TLS trong bộ nhớ RAM, không ghi Disk.
 │    ├── db_password.txt
 │    ├── api_key.txt
 │    └── ...
 │
 └── init/                    <-- (Tùy chọn) Gắn Bind-mount script khởi tạo database lần đầu
      ├── 01-schema.sql
      ├── 02-seed-data.sql
      └── setup.sh
```

**LƯU Ý:** TUYỆT ĐỐI KHÔNG chứa thư mục như `data/` trong Bundle này. Dữ liệu vật lý/Persistent Data phải được định tuyến qua các Biến Môi Trường (như `$DATA_ROOT`) lưu ở ổ cứng riêng của Cụm. Bundle chỉ chứa "Linh hồn" ứng dụng (Stateless), không chứa "Thể xác" (Stateful).

## 2. Cách viết `docker-compose.yml` theo chuẩn Mới

Thay vì cố gắng mount file cấu hình bằng đường dẫn cứng dễ lỗi phân quyền: `- ./configs/nginx.conf:/etc/nginx/nginx.conf`

Trong Swarm, bạn **phải** sử dụng khái niệm `configs` và `secrets`:

```yaml
version: '3.8'

services:
  web:
    image: nginx:alpine
    configs:
      - source: web_config
        target: /etc/nginx/nginx.conf  # Swarm tự động mount đè file vào vị trí này
    secrets:
      - source: api_key
        target: /run/secrets/api_key  # Trao quyền đọc cho App từ thư mục bí mật
    networks:
      - proxy_public
    deploy:
      replicas: 3
    volumes: # Volume Stateful ĐƯỢC CHỈ ĐỊNH VỀ DATA_ROOT
      - ${DATA_ROOT}/my-awesome-app/uploads:/var/www/uploads

configs:
  web_config:
    file: ./configs/nginx.conf

secrets:
  api_key:
    file: ./secrets/api_key.txt

networks:
  proxy_public:
    external: true
```

## 3. Quy trình Lệnh `swarm-ctl app deploy` Triển khai Bundle

Khi bạn hoàn thiện cấu trúc thư mục trên, chỉ cần chạy đúng 1 lệnh từ máy tính cá nhân:

```bash
swarm-ctl app deploy /home/products/services/my-awesome-app --name webapp
```

**Điều gì diễn ra đằng sau?**

1. **Phân tích ENV:** Tool tự lấy biến `$DATA_ROOT` đang config sẵn trên cụm, trộn với file `.env` cục bộ rải rác đều lên Bundle Context.
2. **Bundle/Zip Đóng Gói:** Tool nén toàn bộ thư mục `my-awesome-app` của bạn thành file `webapp.tar.gz`. *Không lo sợ dính rác hay cache.*
3. **Truyền Hình Ký (SSH Transport):** Tập lệnh nén gửi tốc hành từ Client Laptop -> thẳng qua Master Node / Manager Node qua đường hầm SSH bảo mật.
4. **Triển khai Native Swarm Deploy:** Tại máy chủ, Tool chỉ đạo Master Node giải nén, sau đó gõ đúng 1 lệnh Swarm Vô đối: 
   `docker stack deploy -c docker-compose.yml -c docker-compose.prod.yml webapp`
5. **Dọn dẹp (Cleanup):** Cấu hình Bundle cũ xóa bỏ, Docker Swarm phân rã các cấu hình, phân phối đi mọi Node bằng Internal Routing Mesh, tự thiết lập TLS, hoàn thiện App!

Với cơ chế quy hoạch Thư mục này:
- Bạn lỡ tay ghi config lỗi? Chỉ cần sửa config tại máy cục bộ (Local), chạy lại lệnh `app deploy`. Docker Swarm so sánh khác biệt mã Checksum Hash và kích hoạt Rolling Updates tự động (KHÔNG DOWN TIME).
- Quản trị Version Control? Chỉ việc Git Push/Pull folder là y nguyên Server ở Cụm mới.
