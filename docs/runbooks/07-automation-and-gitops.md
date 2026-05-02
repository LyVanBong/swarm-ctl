# 🚀 Hướng Dẫn Thiết Lập CI/CD Tự Động Hóa Với Github Actions

Để cấu hình Github Actions tự động nhảy vào Server của bạn và gõ lệnh deploy mỗi khi có code mới, bạn chỉ cần thực hiện đúng 2 bước sau đây. Hệ thống sẽ mất khoảng 5 phút để thiết lập lần đầu tiên và dùng mãi mãi về sau.

---

## Bước 1: Khai báo Chìa khóa bí mật (Github Secrets)

Github cần biết địa chỉ IP của máy chủ và cần một chiếc "Chìa khóa" để mở cửa Server của bạn một cách an toàn mà không cần nhập mật khẩu.

1. Lấy khóa trên Server: Bạn SSH vào máy chủ Master của bạn, gõ lệnh sau để in mã khóa ra màn hình:
   ```bash
   cat /root/.ssh/id_ed25519
   ```
   *(Hãy copy toàn bộ đoạn văn bản hiện ra, bao gồm cả dòng `-----BEGIN...` và `-----END...`)*

2. Mở trình duyệt, truy cập vào Repo mã nguồn của bạn trên **Github.com**.
3. Chuyển sang tab **Settings** ➔ **Secrets and variables** ➔ **Actions**.
4. Bấm nút màu xanh **New repository secret**, lần lượt tạo 2 biến sau:
   *   **Tên biến 1:** `MASTER_IP`
       **Giá trị:** Điền IP máy chủ của bạn (Vd: `81.17.101.123`)
   *   **Tên biến 2:** `SSH_PRIVATE_KEY`
       **Giá trị:** Dán toàn bộ nội dung file chìa khóa mà bạn vừa copy ở trên vào đây.

---

## Bước 2: Nhúng File Kịch Bản (Workflow YAML) vào Mã Nguồn

Trong dự án mã nguồn của bạn trên máy tính cá nhân (ví dụ thư mục code `softty-ecosystem`), hãy tạo một cấu trúc thư mục đặc biệt như sau: `.github/workflows/deploy.yml`.

Mở file `deploy.yml` lên và dán nguyên xi đoạn mã này vào:

```yaml
name: Tự động Triển khai lên Docker Swarm (Production)

# Kịch bản này sẽ tự động chạy khi có ai đó Push code lên nhánh "main"
on:
  push:
    branches:
      - main

jobs:
  build-and-deploy:
    runs-on: ubuntu-latest
    steps:
      - name: 1. Checkout Mã Nguồn
        uses: actions/checkout@v3

      # ------------------------------------------------------------------
      # [TÙY CHỌN] NẾU BẠN CÓ BƯỚC BUILD DOCKER IMAGE THÌ ĐỂ Ở ĐÂY
      # - name: Build và Push Docker Image
      #   run: |
      #     docker build -t my-dockerhub-user/my-app:${{ github.sha }} .
      #     docker push my-dockerhub-user/my-app:${{ github.sha }}
      # ------------------------------------------------------------------

      - name: 2. Cài đặt Swarm-ctl CLI vào máy ảo Github
        run: |
          curl -sL https://github.com/LyVanBong/swarm-ctl/releases/latest/download/swarm-ctl-linux-amd64 -o swarm-ctl
          chmod +x swarm-ctl
          sudo mv swarm-ctl /usr/local/bin/

      - name: 3. Lắp Chìa khóa SSH & Khởi tạo Kết nối
        env:
          SSH_KEY: ${{ secrets.SSH_PRIVATE_KEY }}
          IP: ${{ secrets.MASTER_IP }}
        run: |
          # 1. Tạo file khóa bảo mật
          mkdir -p ~/.ssh
          echo "$SSH_KEY" > ~/.ssh/id_ed25519
          chmod 600 ~/.ssh/id_ed25519
          
          # 2. Khai báo cho Swarm-ctl biết Server nằm ở đâu
          swarm-ctl config set master-ip "$IP"
          swarm-ctl config set ssh-key "~/.ssh/id_ed25519"

      - name: 4. Chốt hạ! Bắn lệnh Deploy sang máy chủ
        run: |
          # Chọn 1 TRONG 2 cách dưới đây tùy nhu cầu của bạn:
          
          # CÁCH A: Dùng để Deploy/Update nguyên một Thư mục Bundle (có file docker-compose.yml) 
          #         Lệnh này có kèm theo kéo mật khẩu tự động từ Infisical
          swarm-ctl app deploy ./ --infisical-project "proj_xyz" --infisical-env "prod"

          # ---------------- HOẶC ----------------
          
          # CÁCH B: Chỉ thay đổi mỗi cái phiên bản Image của 1 Service đang chạy sẵn
          # swarm-ctl service update ecosystem_web --image my-dockerhub-user/my-app:${{ github.sha }}
```

## Bước 3: Hưởng thụ thành quả
Bây giờ, bạn hãy lưu file lại và gõ lệnh Push code lên Github:
```bash
git add .
git commit -m "Thêm kịch bản CI/CD Github Actions"
git push origin main
```

Ngay lập tức, bạn hãy mở trình duyệt lên, vào tab **Actions** trên repo Github của bạn. Bạn sẽ thấy một tiến trình màu vàng đang xoay vòng vòng. Khi nó báo tích xanh (✅) là ứng dụng của bạn đã được cập nhật thành công rực rỡ lên Server mà tay bạn không cần đụng vào một giọt mồ hôi nào!
