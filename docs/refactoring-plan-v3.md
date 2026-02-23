# Quy hoạch Kiến trúc Tái cấu trúc swarm-ctl (Bản 3.0)

Kế hoạch này vạch ra chiến lược Tái cấu trúc toàn diện (Refactoring Plan) cho công cụ CLI `swarm-ctl`, nhằm xóa bỏ những "khoảng mù" của cấu trúc cũ (Ansible ôm đồm cứng nhắc) và nâng cấp lên hệ thống triển khai phi trạng thái tiêu chuẩn Doanh nghiệp.

## 1. Tư Duy Kiến Trúc Lõi (Paradigm Shift)

Cụm Swarm 3.0 sẽ hoạt động dưới **Nguyên tắc Phân Tách Bề Mặt:**

*   **Stateless (Phi trạng thái):** Bản thân Ứng dụng, Config, Secret. Nó chạy ở RAM, dễ dệt thành gói `Bundle Thư mục` trên máy Client. Gỡ đi tải lại dễ dàng, có thể rớt mọi Node bất kỳ.
*   **Stateful (Có trạng thái):** Thư mục CSDL (`/opt/data/`), Upload,... Là thứ tuyệt đối bất khả xâm phạm. Cấu hình triển khai chỉ nhận ra chúng qua Môi trường `$DATA_ROOT`, không gắn chết Local Mount.

## 2. Kế Hoạch 4 Trụ Cột Tái Cấu Trúc (Action Plan)

Thiết kế kiến trúc Lệnh (CLI Command) sẽ được quy hoạch thành 4 mảng chính:

### Trụ Cột 1: PROMPT TƯƠNG TÁC THÔNG MINH (Interactive CLI)
*   **Vấn đề:** Lỗi "cụt ngủn" do quên Cờ (Flags). Cảm giác sử dụng khô khan.
*   **Giải pháp:** Code GoCLI sẽ chèn Prompt/Readline library (Vd: `survey` hoặc `huh`).
    *   Hỏi User các thông tin còn thiếu (IP Master, Domain, SSH Key).
    *   Làm đẹp màn hình bằng Component (Box, Banner, Checklist). Tự động nhớ Context.

### Trụ Cột 2: CORE CLEANUP (Thanh lọc Lệnh Triển khai Lõi `cluster init`)
*   **Vấn đề:** Quá nặng (15-20p init), ôm đồm Tier 2 (DB, Monitoring) vào Ansible, gây Permission Denied và Fail Password Hash do tạo Secret ẩn.
*   **Giải pháp:** Đập bỏ các cục tạ.
    *   Xóa triệt để file `deploy-tier2.yml`. Application Tier bị cấm can thiệp vào Init Phase.
    *   Thêm Flag `--traefik-domain` và `--portainer-domain` cho phép tự custom Route Public Lõi thay vì fix cứng.
    *   `init` giờ chỉ làm Tier 0 (Hạ tầng Máy Chủ/Network UFW/Docker/Swarm) và Tier 1 (Router Công Hành Traefik). Thời gian chạy rút xuống < 2 phút.

### Trụ Cột 3: BUNDLE DEPLOY (Sức mạnh Triển khai Ứng Dụng)
*   **Vấn đề:** Ansible không hợp để Deploy App, và không có tính linh hoạt (Config Cứng tại `playbooks/`).
*   **Giải pháp:** Đại tu Lệnh `swarm-ctl app deploy <đường_dẫn_folder_app> --name <stack_name>`.
    *   Tool đứng từ máy Client (Thư mục mã nguồn Cục bộ) đọc Cấu trúc chuẩn: File `.yml` + `configs/` + `secrets/`.
    *   Tool nén Zip (Tarball) Bundle này lại.
    *   Bắn sang Server Ảo hóa băng thông cao (SCP/SSH).
    *   Trên Host, gọi Docker Stack Deploy gốc rễ của Swarm thực thi thay hệ thống, tận dụng nguyên bản Load Balancing và Zero Downtime.

### Trụ Cột 4: SHIELD (Hệ Sinh Thái Bảo Vệ - Tương Lai)
*   **Backup Data:** Module `swarm-ctl backup` gọi Rsync nén toàn bộ `$DATA_ROOT` và đẩy Cloud S3/GDrive.
*   **Alerting Bot:** Tool sẽ thiết lập bot ngầm (Daemon) quét Event Swarm. Đập tan cấu trúc Monitoring phức tạp của Prometheus/Alfana-Manager bằng Bot CLI độc bản bắn thẳng Telegram/Email.

---

> Toàn bộ bản Tái cấu trúc này nhằm thay thế hệ thống Tool rối rắm cũ bằng chuẩn mực Triển khai Phi Trạng Thái, cực kỳ vững vàng và thân thiện với thao tác của System Administrator tương lai.
