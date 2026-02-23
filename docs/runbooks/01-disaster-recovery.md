# Runbook: Khôi Phục Sự Cố & Backup (Disaster Recovery)

Tài liệu này hướng dẫn cách phản ứng khi hệ thống gặp sự cố mất dữ liệu, hoặc khi cần chuyển đổi server (Migration).

## 1. Backup Thủ Công (Trước khi thực hiện thay đổi nguy hiểm)

Luôn luôn chạy backup toàn bộ cluster trước khi update version Database, xóa node, hoặc nâng cấp OS.

```bash
# Tạo bản backup toàn bộ (bao gồm thư mục volumes và sql dump)
swarm-ctl backup create --tag pre-upgrade
```

Bản backup sẽ được lưu tại `DATA_ROOT/backups/` trên node Master (Theo mặc định là `/opt/data/backups/`).

**Lưu ý:** Nếu bạn cần tải file backup này về máy tính cá nhân, hãy dùng SFTP/SCP:
```bash
scp root@<master-ip>:/opt/data/backups/backup-pre-upgrade-YYYYMMDD-HHMMSS/volumes.tar.gz ./local-dir
```

## 2. Restore (Khôi phục Dữ liệu)

Nếu ứng dụng bị lỗi dữ liệu (ví dụ: vô tình DROP TABLE hoặc format nhầm mount data):

```bash
# Xem danh sách các bản backup đang có
swarm-ctl backup list

# Lấy ID của bản backup (Ví dụ: backup-20260223-030000)
swarm-ctl backup restore backup-20260223-030000
```
*Hệ thống sẽ nạp lại file `.sql` vào MariaDB và giải nén thư mục cấu hình về `/opt/data`.*

## 3. Khôi Phục Master Node (Khi Server Master chết hoàn toàn)

Docker Swarm sử dụng cơ chế Raft consensus. Nếu bạn chỉ có 1 Manager Node và server đó chết (cháy ổ cứng, mất kết nối không thể khôi phục), bạn sẽ phải khởi tạo lại Cluster từ đầu:

1. Trỏ DNS Domain (ví dụ: `softty.net`) về IP của Server Master mới.
2. Từ máy chủ chứa source `swarm-ctl`, chạy lại lệnh:
   ```bash
   swarm-ctl cluster init --master <IP_MOI> --domain softty.net ...
   ```
3. Upload lại thư mục `/opt/data/backups` từ máy cá nhân lên server mới tại `/opt/data/backups/`.
4. Restore lại bản backup như hướng dẫn phần 2.
5. Thêm lại từng Worker Node cũ (Phải chạy `docker swarm leave --force` trên Worker cũ trước):
   ```bash
   swarm-ctl node add --ip <WORKER_IP> --role worker
   ```

*Tip: Nếu có trên 3 Manager Nodes thì 1 Manager chết không làm gián đoạn Swarm.*
