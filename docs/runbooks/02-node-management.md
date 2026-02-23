# Runbook: Quản lý hạ tầng Máy Chủ (Node Management)

Tài liệu hướng dẫn việc **Thêm**, **Xóa**, và **Bảo trì Máy chủ (Node)** nằm trong cluster.

## 1. Bảo trì (Patch Linux OS, Reboot máy chủ Worker)

Khi bạn muốn cập nhật bảo mật cho nhân Linux, tắt máy ảo để tăng RAM, hay nâng cấp phiên bản Docker:

**Bước 1: Drain máy chủ để di dời container**
Node phải chuyển vào trạng thái `Drain` để không có dịch vụ nào chạy trên nó trong lúc bảo trì. (Cluster Controller sẽ chuyển tải sang các server khác).
```bash
swarm-ctl node label remove <IP_SERVER> tier=app
# Drain: Ngừng ứng dụng
docker node update --availability drain <NODE_ID>
```

**Bước 2: Bảo trì (SSH / Reboot)**
Truy cập trực tiếp:
```bash
swarm-ctl node ssh <IP_SERVER>
# Thực hiện apt update && apt upgrade
# Sau đó reboot:
sudo reboot
```

**Bước 3: Gia nhập trở lại**
Khi máy tính khởi động lên, cấu hình Docker Daemon sẽ tự bật dịch vụ lại.
Mở chế độ hoạt động bình thường `Active`:
```bash
docker node update --availability active <NODE_ID>
swarm-ctl node label add <IP_SERVER> tier=app
```

## 2. Thêm mới Server Worker (Scale cụm Backend App)

Nếu máy chủ Appwrite bị đầy tải, CPU từ 80-100%, bạn thuê một Cloud Server mới.
Lắp nó vào cụm chỉ với 1 bước:

```bash
swarm-ctl node add \
  --ip 123.45.67.89 \
  --role worker \
  --label tier=app
```
Label `tier=app` giúp Swarm định tuyến những Ứng dụng/Web Service triển khai vào con mới cài đặt, thay vì triển khai chung Node chứa Database/Redis.

## 3. Loại bỏ máy chủ Worker an toàn
Nếu bạn không cần thuê nó nữa:
```bash
# Nó sẽ hỏi bạn IP để xác thực kỹ trước khi drain và remove.
swarm-ctl node remove --ip 123.45.67.89
```

*Cảnh báo: Không thể xóa Node đang là Quorum (Mảnh ghép Manager) nếu tổng số Manager giảm xuống chẵn (VD: Xuống 2, hoặc 4 Manager) vì nó sẽ gây split-brain cluster.*
