# Runbook: Cập Nhật Ứng Dụng & Thay Mật Khẩu Không Gián Đoạn (Zero-Downtime)

Docker Swarm sử dụng chiến lược `Rolling Update` trên Manager Node để chạy Container phiên bản mới trước khi tắt bản cũ. Do đó sẽ không bị rớt bất cứ Request nào từ khách hàng (Zero Downtime).

## 1. Cập nhật Phiên bản Phần mềm (Images/Tag)

Ví dụ khi bạn muốn push bản build phiên bản `v2.0` cho `my-api`:
```bash
# Push bản mới lên Docker Registry
docker build -t your-registry.com/my-api:v2.0 .
docker push your-registry.com/my-api:v2.0

# Kích hoạt Rolling Update trên cụm Cluster
swarm-ctl service update my-api --image your-registry.com/my-api:v2.0
```
Tiến trình diễn ra: 1 Service Replica (Thực thể) bị tắt > 1 Thực thể v2.0 khởi chạy > Kiểm tra HealthCheck > Success > Tắt thực thể v1.1 tiếp theo... và cuối cùng tất cả Container App đều là `v2.0`.

Hệ thống Traefik Gateway sẽ tự cân bằng tải.

## 2. Lắp thêm Node và Scale Dịch Vụ
Nếu sự kiện lớn ập đến và trang web / Mobile App bị quá tải (Giả sử bạn đã có thêm 10 Worker Nodes):

```bash
# Tăng tải Replica Container từ 3 lên 15 bản gốc (Tùy cấu hình RAM máy chủ)
swarm-ctl service scale my-api=15
```

## 3. Cập nhật Mật khẩu Database (Rotate Secret)

Cực kỳ quan trọng: Việc thay đổi mật khẩu Root/App của Database sẽ phải cập nhật cho hàng loạt Service liên quan. 

**Quy trình chuẩn bằng swarm-ctl:**
1. Rotating Zero-downtime: 
   Công cụ thiết lập `secret mới` -> Đẩy ứng dụng kết nối lại = Secret mới -> Lần lượt xóa mật khẩu cũ đi.
   
```bash
# Rotate bí mật Database một cách không gián đoạn
swarm-ctl secret rotate mariadb_password "Th1s!s@N3wStr0ngP@ss"
```
Khi chạy `swarm-ctl secret rotate`, công cụ sẽ tìm tất cả các Application phụ thuộc vào secret trên, bắt buộc nó tiến hành update thông qua cơ chế `Rolling Update` với cấu hình nhận mật khẩu mới.

## 4. Revert / Rollback do lỗi sản xuất
Sau khi release, phát hiện bug (Crash / Load CPU 100%), bạn có thể hủy tác vụ Update và trả về bản ổn định liền trước đó:

```bash
swarm-ctl service rollback my-api
```
Bản Update gần nhất trên Registry Image sẽ được gọi ra.
