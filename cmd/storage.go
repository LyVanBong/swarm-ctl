package cmd

import (
	"fmt"
	"strings"

	"github.com/LyVanBong/swarm-ctl/internal/config"
	"github.com/LyVanBong/swarm-ctl/internal/ssh"
	"github.com/LyVanBong/swarm-ctl/internal/ui"
	"github.com/spf13/cobra"
)

var storageCmd = &cobra.Command{
	Use:   "storage",
	Short: "Quản lý distributed storage (MinIO)",
}

// ──────────────────────────────────────────────
// swarm-ctl storage status
// ──────────────────────────────────────────────
var storageStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Xem trạng thái MinIO storage",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		cluster, err := cfg.GetCurrentCluster()
		if err != nil {
			return err
		}

		client := ssh.NewClient(cluster.MasterIP, cluster.SSHUser, cluster.SSHKey)
		if err := client.Connect(); err != nil {
			return err
		}
		defer client.Close()

		fmt.Println(ui.SectionHeader.Render(" STORAGE STATUS "))

		// Kiểm tra MinIO container
		minioID, err := client.Run(
			"docker ps --filter 'name=minio_minio' --format '{{.ID}}|{{.Status}}' | head -1")
		if err != nil || strings.TrimSpace(minioID) == "" {
			fmt.Println(ui.RenderWarning("MinIO chưa chạy"))
			fmt.Println(ui.RenderInfo("Deploy: swarm-ctl service deploy minio"))
			return nil
		}

		parts := strings.Split(strings.TrimSpace(minioID), "|")
		fmt.Printf("  MinIO Container: %s\n", ui.Success.Render(parts[0]))
		if len(parts) > 1 {
			fmt.Printf("  Status         : %s\n", ui.Info.Render(parts[1]))
		}

		// Disk usage
		fmt.Println()
		fmt.Println(ui.SectionHeader.Render(" DISK USAGE "))
		diskOut, _ := client.Run(fmt.Sprintf("df -h %s/minio 2>/dev/null | tail -1", cluster.DataRoot))
		if diskOut != "" {
			fmt.Println("  " + strings.TrimSpace(diskOut))
		}

		// Buckets list (gọi MinIO client nếu có)
		minioClientExists, _ := client.Run("which mc 2>/dev/null")
		if strings.TrimSpace(minioClientExists) != "" {
			fmt.Println()
			fmt.Println(ui.SectionHeader.Render(" BUCKETS "))
			bucketsOut, _ := client.Run("mc ls local/ 2>/dev/null | head -20")
			if bucketsOut != "" {
				fmt.Println(bucketsOut)
			}
		}

		return nil
	},
}

// ──────────────────────────────────────────────
// swarm-ctl storage create-bucket
// ──────────────────────────────────────────────
var storageCreateBucketCmd = &cobra.Command{
	Use:   "create-bucket [name]",
	Short: "Tạo bucket mới trên MinIO",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		bucketName := args[0]

		cfg, err := config.Load()
		if err != nil {
			return err
		}
		cluster, err := cfg.GetCurrentCluster()
		if err != nil {
			return err
		}

		client := ssh.NewClient(cluster.MasterIP, cluster.SSHUser, cluster.SSHKey)
		if err := client.Connect(); err != nil {
			return err
		}
		defer client.Close()

		// Tạo bucket qua MinIO client hoặc API
		minioID, err := client.Run(
			"docker ps --filter 'name=minio_minio' -q | head -1")
		if err != nil || strings.TrimSpace(minioID) == "" {
			return fmt.Errorf("MinIO chưa chạy — deploy trước: swarm-ctl service deploy minio")
		}

		containerID := strings.TrimSpace(minioID)
		createCmd := fmt.Sprintf(
			"docker exec %s mc mb /data/%s --ignore-existing 2>/dev/null || "+
				"docker exec %s mkdir -p /data/%s",
			containerID, bucketName, containerID, bucketName)

		if _, err := client.Run(createCmd); err != nil {
			return fmt.Errorf("tạo bucket thất bại: %w", err)
		}

		fmt.Println(ui.RenderSuccess(fmt.Sprintf("Bucket '%s' đã được tạo", bucketName)))
		return nil
	},
}

// ──────────────────────────────────────────────
// swarm-ctl storage expand
// ──────────────────────────────────────────────
var storageExpandCmd = &cobra.Command{
	Use:   "expand --node [ip]",
	Short: "Thêm storage capacity bằng cách thêm MinIO node",
	Long: `Mở rộng MinIO storage cluster bằng cách thêm node mới.

Lưu ý: MinIO distributed mode yêu cầu số node là bội số của erasure set size.
Thông thường: 4, 8, 12, 16 nodes.

Ví dụ:
  swarm-ctl storage expand --node 10.0.0.6`,

	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println(ui.RenderInfo("Storage expand: Cần plan MinIO distributed mode trước"))
		fmt.Println()
		fmt.Println(`Để scale MinIO trên Docker Swarm:
1. Deploy MinIO với distributed mode từ đầu (không thể convert từ single)
2. Dùng MinIO Gateway hoặc NAS gateway
3. Hoặc add storage via volume expansion trên existing node

Xem thêm: https://min.io/docs/minio/linux/operations/install-deploy-manage/expand-minio-deployment.html`)
		return nil
	},
}

func init() {
	storageCmd.AddCommand(storageStatusCmd)
	storageCmd.AddCommand(storageCreateBucketCmd)
	storageCmd.AddCommand(storageExpandCmd)
}
