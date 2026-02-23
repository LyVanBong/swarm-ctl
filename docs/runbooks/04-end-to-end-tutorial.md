# BÀI TẬP THỰC CHIẾN (END-TO-END TUTORIAL)

Đây là tài liệu Runbook hướng dẫn **Từ A đến Z** một kịch bản sử dụng cực kì phổ biến trong môi trường Doanh nghiệp (Product/Enterprise). Kịch bản được thiết kế hoàn toàn mạch lạc, dễ hiểu và "bấm là chạy".

## 🎯 Kịch bản Giả định (Use Case)

Công ty bạn (Tên miền: `softty.net`) vừa phát triển xong dự án Website tin tức/thương mại. 
Giám đốc kĩ thuật yêu cầu bạn thiết kế cụm Cloud Server gồm 3 máy để chịu tải được 1 triệu lượt truy cập và không bị rớt mạng khi 1 máy bị chập cháy phần cứng.

**Kiến trúc mong muốn:**
- Hệ thống BaaS (Backend-as-a-service): **Appwrite** (chạy tại `api.softty.net`).
- Giao diện người dùng: **Frontend Web / ReactJS** (chạy tại `www.softty.net`).

---

## 🟢 BƯỚC 1: TIỀN TRẠM (CHUẨN BỊ MÁY MÓC & DNS)

Bạn bắt đầu bằng việc đi thuê 3 chiếc Server ảo (Cloud VPS) có chạy hệ điều hành Ubuntu 22.04 LTS (Hoặc Rocky/Debian tùy ý). Cấu hình:
* **Server 1 (Làm Manager):** IP `103.11.22.M` (Phải có ít nhất 4 Core, 8GB RAM vì nó gánh DB cốt lõi của Appwrite).
* **Server 2 (Làm Worker 1):** IP `103.11.22.W1` (2 Core, 4GB RAM).
* **Server 3 (Làm Worker 2):** IP `103.11.22.W2` (2 Core, 4GB RAM).
*(Bạn sẽ nhận được 3 cái Mật khẩu `root` tương ứng từ Email của nhà mạng phân phối, ví dụ: PassRoot123).*

**Vào trang quản lý tên miền (Cloudflare/Tenten), tạo 2 Bản ghi phân giải:**
* Loại: `A` | Host: `api` | Content: `103.11.22.M`
* Loại: `A` | Host: `www` | Content: `103.11.22.M`
*(Lưu ý: Mọi tên miền đều trỏ thẳng về cánh cửa của Máy Mẹ số 1).*

---

## 🟢 BƯỚC 2: KHỞI TẠO NÃO BỘ MẠNG LƯỚI (MANAGER)

Mở màn hình đen (Terminal) trên Máy tính Laptop/Cán bộ Quản trị của bạn (Nơi đã cài `swarm-ctl`). Bạn tiến hành cướp quyền điều khiển con VPS số 1 để gộp hệ thống:

```bash
swarm-ctl cluster init --master 103.11.22.M --domain softty.net --pass "PassRoot123"
```

**Điều gì diễn ra phía sau?**
Khi bạn truyền cờ `--pass`, tool `swarm-ctl` sẽ giải phóng đôi tay bạn. Nó tự sinh chứng chỉ rsa, nhồi password vào Máy Mẹ, sau đó tự bật Ansible để cài Docker, sinh Cụm Swarm Manager, Rắc hạt định tuyến mạng `Traefik`... mà không yêu cầu bạn phải ấn Enter thêm cái nào. (Chờ khoảng 5-10 phút).

---

## 🟢 BƯỚC 3: MỞ RỘNG MẠNG LƯỚI QUA WORKERS

Máy Mẹ đã tỉnh giấc. Hãy vung gậy điều binh 2 máy tính Trâu cày (Worker) vào đọi hình mạng lưới. Nhập 2 lệnh liên tiếp:

```bash
swarm-ctl node add --ip 103.11.22.W1 --role worker --pass "KhoaBomVuW1_123"
swarm-ctl node add --ip 103.11.22.W2 --role worker --pass "KhoaBomVuW2_123"
```

🔥 **Khám nghiệm thành công:** Bạn gõ `swarm-ctl dashboard` sẽ thấy màn hình TUI hiện vạch sóng xanh lá cây hiển thị Máy Mẹ và 2 Worker đang online nối mạng LAN với nhau trơn tru.

---

## 🟢 BƯỚC 4: TRIỂN KHAI VŨ KHÍ NẶNG (APPWRITE)

Appwrite (Backend) sinh ra 1 lúc hơn 20 cái Container nhỏ lẻ siêu phức tạp. Bạn không cần làm gì ngoài gõ:

```bash
swarm-ctl app install appwrite --domain api.softty.net
```

**Vì sao kịch bản này lại an toàn tuyệt đối?**
Trong thiết kế ngầm (`Template`) của lệnh Cài đặt: Cơ sở dữ liệu MariaDB, Cấu hình File tĩnh đã bị "khoá còng số 8" (placement constraints) không cho chạy đi đâu ngoài MÁY SỐ 1 MẸ để chống phân mảnh dữ liệu. 
Còn lõi xử lý PHP `Appwrite Core API` thì được thả cho chạy nhảy trên Cả 3 máy (Tuỳ thằng nào Cổ RAM rảnh rỗi hơn thì ưu tiên ném qua thằng đấy chạy giùm). Người dùng truy cập `https://api.softty.net` sẽ được Traefik SSL rải thảm HTTP đưa tin vô mượt mà.

---

## 🟢 BƯỚC 5: NẠP ĐẠN GIAO DIỆN XUẤT TRẬN (FRONTEND WEB)

Giả định Bộ phận lập trình của bạn đã nhồi xong Code HTML CSS vào cái cục tên là `softty/my-react-website:v1.0.0` (Đẩy lên DockerHub công khai hoặc nội bộ).
Bạn muốn chạy Website này trên 2 con Worker để gánh cho nhau khi có sự cố? Hãy nã phát súng cuối cùng:

```bash
swarm-ctl service add \
  --name frontend-web \
  --image softty/my-react-website:v1.0.0 \
  --domain www.softty.net \
  --port 80 \
  --replicas 2
```

**Thành quả tối cao (High Availability / 99.99% Uptime):**
Chữ số `--replicas 2` trong lệnh trên đã ra chỉ thị sao chép Code của bạn thành 2 bản phân bổ rải rác.
Nếu Datacenter số 2 của Máy Worker 2 bị chuột cắn cụt dây mạng (Máy W2 cháy đen/sít off), các khách truy cập vào trang web `www.softty.net` lúc 2 giờ sáng vẫn **HIỆN RÕ VÀ KHÔNG BỊ DISCONNECT**. Vì Traefik tại Gateway số 1 Đã nhìn thấy máy 2 chết nên chẻ luồng Mạng rẽ sang Máy Worker số 1.

Sức mạnh của Hạ tầng Cluster phân mảnh đến từ những Kỹ sư Vận Hành hiểu luật như bạn! 🎯 Mọi thứ đã Live, xin mời truy cập bằng trình duyệt Chrome vào HTTPs. Khoá xanh đã bảo mật.
