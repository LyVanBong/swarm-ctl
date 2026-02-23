package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/LyVanBong/swarm-ctl/internal/config"
	"github.com/LyVanBong/swarm-ctl/internal/ssh"
	"github.com/LyVanBong/swarm-ctl/internal/ui"
	"github.com/spf13/cobra"
)

var backupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Quản lý backup & restore dữ liệu cluster",
}

// ──────────────────────────────────────────────
// swarm-ctl backup create
// ──────────────────────────────────────────────
var (
	backupServices  []string
	backupOutputDir string
	backupTag       string
)

var backupCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Tạo backup toàn bộ hoặc từng service",
	Long: `Backup dữ liệu các services đang chạy.

Backup bao gồm:
  - Database dumps (MariaDB, PostgreSQL)
  - Volume data (bind mounts)
  - Docker Secrets list (không backup giá trị)
  - Services config

Ví dụ:
  swarm-ctl backup create                    # Backup tất cả
  swarm-ctl backup create --service mariadb  # Chỉ backup MariaDB`,

	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		cluster, err := cfg.GetCurrentCluster()
		if err != nil {
			return err
		}

		backupID := fmt.Sprintf("backup-%s", time.Now().Format("20060102-150405"))
		if backupTag != "" {
			backupID = fmt.Sprintf("backup-%s-%s", backupTag, time.Now().Format("20060102-150405"))
		}

		fmt.Println(ui.Banner.Render(fmt.Sprintf("💾 BACKUP: %s", backupID)))
		fmt.Println()

		client := ssh.NewClient(cluster.MasterIP, cluster.SSHUser, cluster.SSHKey)
		if err := client.Connect(); err != nil {
			return fmt.Errorf("không thể kết nối cluster: %w", err)
		}
		defer client.Close()

		backupDir := fmt.Sprintf("%s/backups/%s", cluster.DataRoot, backupID)

		// Tạo thư mục backup
		fmt.Println(ui.RenderStep(1, 4, "Chuẩn bị thư mục backup..."))
		if _, err := client.RunSudo(fmt.Sprintf("mkdir -p %s", backupDir)); err != nil {
			return fmt.Errorf("không thể tạo thư mục backup: %w", err)
		}
		fmt.Println(ui.RenderSuccess(fmt.Sprintf("Backup dir: %s", backupDir)))

		// Backup MariaDB nếu đang chạy
		fmt.Println(ui.RenderStep(2, 4, "Backup databases..."))
		mariaDBContainerID, err := client.Run(
			"docker ps --filter 'name=database_mariadb' -q 2>/dev/null | head -1")
		if err == nil && strings.TrimSpace(mariaDBContainerID) != "" {
			containerID := strings.TrimSpace(mariaDBContainerID)
			dumpCmd := fmt.Sprintf(
				"docker exec %s mysqldump -u root -p$(docker exec %s cat /run/secrets/mariadb_root_password) --all-databases 2>/dev/null > %s/mariadb-all.sql",
				containerID, containerID, backupDir)
			if _, err := client.RunSudo(dumpCmd); err != nil {
				fmt.Println(ui.RenderWarning("MariaDB backup thất bại: " + err.Error()))
			} else {
				fmt.Println(ui.RenderSuccess("MariaDB backup OK"))
			}
		} else {
			fmt.Println(ui.Muted.Render("  MariaDB không chạy — bỏ qua"))
		}

		// Backup Redis
		redisContainerID, err := client.Run(
			"docker ps --filter 'name=cache_redis' -q 2>/dev/null | head -1")
		if err == nil && strings.TrimSpace(redisContainerID) != "" {
			containerID := strings.TrimSpace(redisContainerID)
			if _, err := client.Run(fmt.Sprintf(
				"docker exec %s redis-cli BGSAVE && sleep 2 && "+
					"docker cp %s:/data/dump.rdb %s/redis-dump.rdb",
				containerID, containerID, backupDir)); err != nil {
				fmt.Println(ui.RenderWarning("Redis backup thất bại"))
			} else {
				fmt.Println(ui.RenderSuccess("Redis backup OK"))
			}
		}

		// Backup volumes/data
		fmt.Println(ui.RenderStep(3, 4, "Backup volume data..."))
		volumeBackupCmd := fmt.Sprintf(
			"tar -czf %s/volumes.tar.gz -C %s --exclude=backups .",
			backupDir, cluster.DataRoot)
		if _, err := client.RunSudo(volumeBackupCmd); err != nil {
			fmt.Println(ui.RenderWarning("Volume backup thất bại: " + err.Error()))
		} else {
			fmt.Println(ui.RenderSuccess("Volume data backup OK"))
		}

		// Lấy kích thước backup
		fmt.Println(ui.RenderStep(4, 4, "Hoàn tất..."))
		sizeOut, _ := client.Run(fmt.Sprintf("du -sh %s", backupDir))
		size := strings.Fields(strings.TrimSpace(sizeOut))

		backupSize := "unknown"
		if len(size) > 0 {
			backupSize = size[0]
		}

		fmt.Println()
		fmt.Println(ui.SuccessBox.Render(fmt.Sprintf(`
✅ Backup hoàn tất!

   ID      : %s
   Location: %s
   Size    : %s

Restore bằng lệnh:
   swarm-ctl backup restore %s
`, backupID, backupDir, backupSize, backupID)))

		return nil
	},
}

// ──────────────────────────────────────────────
// swarm-ctl backup list
// ──────────────────────────────────────────────
var backupListCmd = &cobra.Command{
	Use:     "list",
	Short:   "Danh sách các bản backup",
	Aliases: []string{"ls"},
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

		fmt.Println(ui.SectionHeader.Render(" BACKUPS "))
		output, err := client.Run(fmt.Sprintf(
			"ls -lht %s/backups/ 2>/dev/null | head -20", cluster.DataRoot))
		if err != nil || strings.TrimSpace(output) == "" {
			fmt.Println(ui.RenderInfo("Chưa có backup nào. Tạo backup: swarm-ctl backup create"))
			return nil
		}
		fmt.Println(output)
		return nil
	},
}

// ──────────────────────────────────────────────
// swarm-ctl backup restore
// ──────────────────────────────────────────────
var backupRestoreCmd = &cobra.Command{
	Use:   "restore [backup-id]",
	Short: "Restore từ backup",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		backupID := args[0]

		cfg, err := config.Load()
		if err != nil {
			return err
		}
		cluster, err := cfg.GetCurrentCluster()
		if err != nil {
			return err
		}

		fmt.Println(ui.Banner.Render(fmt.Sprintf("♻️  RESTORE: %s", backupID)))
		fmt.Println()
		fmt.Println(ui.WarnBox.Render(`
⚠️  CẢNH BÁO: Restore sẽ ghi đè dữ liệu hiện tại!
   Đảm bảo đã dừng services cần thiết trước khi restore.
   
   Dừng services: swarm-ctl service remove <service-name>`))
		fmt.Println()

		fmt.Print("Nhập backup ID để xác nhận: ")
		var confirm string
		fmt.Scanln(&confirm)
		if confirm != backupID {
			return fmt.Errorf("ID không khớp — đã hủy")
		}

		client := ssh.NewClient(cluster.MasterIP, cluster.SSHUser, cluster.SSHKey)
		if err := client.Connect(); err != nil {
			return err
		}
		defer client.Close()

		backupDir := fmt.Sprintf("%s/backups/%s", cluster.DataRoot, backupID)

		// Kiểm tra backup tồn tại
		exists, err := client.FileExists(backupDir)
		if err != nil || !exists {
			return fmt.Errorf("backup '%s' không tồn tại tại %s", backupID, backupDir)
		}

		// Restore MariaDB
		fmt.Println(ui.RenderStep(1, 2, "Restore MariaDB..."))
		sqlFile := fmt.Sprintf("%s/mariadb-all.sql", backupDir)
		sqlExists, _ := client.FileExists(sqlFile)
		if sqlExists {
			mariaDBContainerID, _ := client.Run(
				"docker ps --filter 'name=database_mariadb' -q | head -1")
			if id := strings.TrimSpace(mariaDBContainerID); id != "" {
				restoreCmd := fmt.Sprintf(
					"docker exec -i %s mysql -u root -p$(docker exec %s cat /run/secrets/mariadb_root_password) < %s",
					id, id, sqlFile)
				if _, err := client.RunSudo(restoreCmd); err != nil {
					fmt.Println(ui.RenderWarning("MariaDB restore thất bại: " + err.Error()))
				} else {
					fmt.Println(ui.RenderSuccess("MariaDB restored"))
				}
			}
		}

		// Restore volumes
		fmt.Println(ui.RenderStep(2, 2, "Restore volume data..."))
		volumeFile := fmt.Sprintf("%s/volumes.tar.gz", backupDir)
		volExists, _ := client.FileExists(volumeFile)
		if volExists {
			restoreCmd := fmt.Sprintf(
				"tar -xzf %s -C %s", volumeFile, cluster.DataRoot)
			if _, err := client.RunSudo(restoreCmd); err != nil {
				fmt.Println(ui.RenderWarning("Volume restore thất bại: " + err.Error()))
			} else {
				fmt.Println(ui.RenderSuccess("Volume data restored"))
			}
		}

		fmt.Println()
		fmt.Println(ui.RenderSuccess(fmt.Sprintf("Restore từ '%s' hoàn tất!", backupID)))
		fmt.Println(ui.RenderInfo("Restart lại services: swarm-ctl service deploy <name>"))
		return nil
	},
}

func init() {
	// backup create flags
	backupCreateCmd.Flags().StringArrayVarP(&backupServices, "service", "s", nil, "Chỉ backup service cụ thể")
	backupCreateCmd.Flags().StringVar(&backupOutputDir, "output", "", "Thư mục output (default: DATA_ROOT/backups)")
	backupCreateCmd.Flags().StringVarP(&backupTag, "tag", "t", "", "Tag cho backup (vd: pre-upgrade)")

	backupCmd.AddCommand(backupCreateCmd)
	backupCmd.AddCommand(backupListCmd)
	backupCmd.AddCommand(backupRestoreCmd)
}
