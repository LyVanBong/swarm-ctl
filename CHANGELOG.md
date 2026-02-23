# Lịch sử Phiên bản (Changelog)

Tất cả những thay đổi quan trọng của công cụ này sẽ được liệt kê tại đây.
Dự án tuân theo quy tắc Đánh số Phiên bản [Semantic Versioning](https://semver.org/).

---

## [1.0.0] - Enterprise-Ready Release (Ngày phát hành hiện tại)
Phiên bản 1.0.0 đánh dấu sự hoàn thiện toàn bộ các tính năng cốt lõi bắt buộc, đáp ứng được tiêu chuẩn An toàn và Độ tin cậy trong môi trường doanh nghiệp.

### Thêm Mới (Added)
- **[Storage]** Hỗ trợ lệnh `storage init-glusterfs` cài đặt GlusterFS phân tán, sao lưu dữ liệu sang 3 máy chỉ với 1 lệnh Ansible.
- **[Audit]** Tích hợp Logger kiểm toán tự động. Lệnh `swarm-ctl audit` truy vết lịch sử gõ lệnh ở cấp độ Local (Tự động che giấu mật khẩu).
- **[Dashboard]** Tích hợp Terminal UI Live Dashboard bằng thư viện Bubbletea với đầy đủ màu sắc, tự động refresh, hiển thị tài nguyên thời gian thực.
- **[Secrets]** `secret rotate` cho phép xoay mật khẩu cấu hình Database an toàn không downtime.
- **[Unit Tests]** Bổ sung Unit Testing vào Pipeline Config và Auditing.
- **[Docs]** Hoàn thiện 3 sách hướng dẫn Runbooks (Disaster Recovery, Node Management, Service Updates).

### Sửa lỗi (Fixed)
- Sửa lỗi LoadConfig bị tham chiếu sai địa chỉ file cục bộ.
- Khắc phục lỗi quyền phát hành GitHub Actions Release thiếu `contents: write`.
- Chỉnh sửa Pipeline SSH giúp thực thi interactive shell thay vì shell ẩn.

---

## [0.2.0] - Networking & Dashboard
### Thêm Mới (Added)
- Lệnh `backup create/restore/list` tự động tar.gz toàn bộ dữ liệu tại `/opt/data/` kèm mysqldump.
- Lệnh `node ssh` và `node label` hỗ trợ gán thẻ (Tag) cho các Worker Node chỉ định.
- Lệnh `doctor` tự động khám lỗi (Quét kết nối, quét Porttainer, quét Network) toàn bộ dàn máy.
- Bổ sung cấu hình Makefile cho Developer.

---

## [0.1.0] - Khởi tạo Dự Án (Initial Release)
- Khung sườn kiến trúc lệnh CLI bằng `Cobra`.
- Parser SSH và Wrapper Ansible tự động gọi Playbook.
- Tự động hóa `cluster init` cấp phát Traefik, Let's Encrypt Certs, và Monitoring Platform (Prometheus + Grafana).
- Tích hợp Git repo và Github actions xuất xưởng tự động Binaries cho 4 nền tảng: Linux/Mac(M1)/Win.
