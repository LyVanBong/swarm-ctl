# Contributing to `swarm-ctl`

Cảm ơn bạn đã quan tâm đóng góp cho dự án `swarm-ctl`! Công cụ này được phát triển để giúp mọi người quản lý hệ thống Docker Swarm một cách dễ dàng, nhanh chóng và chuẩn Enterprise.

Chúng tôi rất hoan nghênh báo cáo lỗi (Bug Reports), yêu cầu tính năng (Feature Requests), và các Pull Requests (PRs).

## Mục Lục

1. [Phát hiện Lỗi (Bug Reports)](#1-phát-hiện-lỗi-bug-reports)
2. [Đề xuất Tính Năng (Feature Requests)](#2-đề-xuất-tính-năng)
3. [Thiết lập Môi Trường Phát Triển](#3-thiết-lập-môi-trường-phát-triển)
4. [Hướng dẫn gửi Pull Request (PR)](#4-hướng-dẫn-gửi-pull-request)
5. [Cấu trúc Project](#5-cấu-trúc-project)

---

## 1. Phát hiện Lỗi (Bug Reports)

Nếu bạn gặp vấn đề hoặc lỗi không mong muốn, vui lòng xem qua mục **[Issues](https://github.com/LyVanBong/swarm-ctl/issues)** để chắc chắn rằng lỗi này chưa được báo cáo.

Khi tạo một Issue mới, vui lòng cung cấp thông tin sau:
- Phiên bản đang sử dụng: `swarm-ctl version`
- Phiên bản Docker và hệ điều hành (OS) Server.
- Các bước để sao chép lại lỗi (Steps to reproduce).
- Output hoặc Log lỗi chi tiết.

## 2. Đề xuất Tính Năng

Nếu bạn có ý tưởng cải tiến `swarm-ctl`, hãy mở một Issue với tag **[Enhancement]** hoặc **[Feature]**. 
- Trình bày rõ ràng bài toán bạn đang giải quyết (Use-case).
- Tại sao tính năng này hữu ích với những người khác trong cộng đồng.
- Đề xuất giải pháp (nếu có).

## 3. Thiết lập Môi Trường Phát Triển

Bạn cần có **Go 1.23+** để biên dịch công cụ.

1. **Clone project:**
   ```bash
   git clone https://github.com/LyVanBong/swarm-ctl.git
   cd swarm-ctl
   ```

2. **Cài đặt dependencies:**
   ```bash
   make tidy
   # hoặc `go mod tidy`
   ```

3. **Biên dịch và chạy thử (Run tests/build):**
   ```bash
   make build    # Build ra binary `swarm-ctl`
   make lint     # Quét Code Linter
   make test     # Chạy Unit Tests
   ```

## 4. Hướng dẫn gửi Pull Request

Chúng tôi áp dụng mô hình Fork & Pull Request.

1. **Fork** repository này về tài khoản của bạn.
2. Cập nhật nhánh branch phát triển từ nhánh `master`.
   *Dùng tên branch gợi nhớ: ví dụ: `feature/auto-scaling` hoặc `fix/ssh-login-bug`.*
3. Viết code, ghi nhớ hãy tuân thủ kiểu viết của Golang và sử dụng Linter:
   - Hãy chạy `make lint` và sửa hết các warning.
   - Khi implement CLI Command mới, cần update **Runbooks / README** nếu thiết yếu.
4. **Commit** với Convention rõ ràng (Ví dụ: `feat: Thêm tính năng Auto-scale`, `fix: Khắc phục lỗi drain node không chờ time-out`).
5. Đẩy (Push) nhánh branch của bạn lên kho lấy về (Forked Repo).
6. Mở một **Pull Request (PR)** hướng tới nhánh `master` của repository gốc.

## 5. Cấu trúc Project
Dành cho bạn muốn đọc Source Code:
- `cmd/`: Các thành phần CLI sử dụng thư viện `Cobra`. Đây là điểm vào của từng Command.
- `internal/`: Logic xử lý của tool:
  - `ansible/`: Wrapper gọi playbook tự động.
  - `ssh/`: Kết nối và thực thi Shell từ xa bằng chuẩn Zero-Trust.
  - `template/`: Generator mẫu file (Compose files, config).
  - `config/`: Cơ chế đọc YAML file lưu Context của Multi-clusters.
  - `ui/`: Các thành phần vẽ giao diện TUI với Bubbletea và Lipgloss.
- `ansible/`: Kho Playbooks chứa mã cấu hình OS (Cài docker, deploy Traefik, setup DB...).
- `docs/runbooks/`: Tài liệu xử lý sự cố nâng cao.

Cảm ơn bạn! Bất kỳ một đóng góp nhỏ nào cũng giúp `swarm-ctl` phát triển bền vững.
