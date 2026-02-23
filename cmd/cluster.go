package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"github.com/LyVanBong/swarm-ctl/internal/ansible"
	"github.com/LyVanBong/swarm-ctl/internal/config"
	"github.com/LyVanBong/swarm-ctl/internal/ssh"
	"github.com/LyVanBong/swarm-ctl/internal/ui"
)

var clusterCmd = &cobra.Command{
	Use:   "cluster",
	Short: "Quản lý Docker Swarm cluster",
	Long:  "Khởi tạo, kiểm tra và quản lý Docker Swarm cluster",
}

// ──────────────────────────────────────────────
// swarm-ctl cluster init
// ──────────────────────────────────────────────
var (
	initMasterIP  string
	initSSHKey    string
	initSSHUser   string
	initDomain    string
	initDataRoot  string
	initClusterName string
	initPass        string // Mật khẩu để chạy ssh-copy-id
)

var clusterInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Khởi tạo cluster mới từ đầu",
	Long: `Bootstrap toàn bộ Docker Swarm cluster trên server của bạn.

Quá trình này sẽ tự động:
  1. Kiểm tra kết nối SSH
  2. Cài đặt Docker Engine
  3. Khởi tạo Docker Swarm
  4. Tạo overlay networks
  5. Deploy Tier 1: Traefik + Portainer
  6. Deploy Tier 2: MinIO + MariaDB Galera + Redis + Monitoring

Ví dụ:
  swarm-ctl cluster init --master 10.0.0.1 --key ~/.ssh/id_rsa --domain example.com`,

	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println(ui.Banner.Render("🐳 SWARM-CTL — Cluster Init"))
		fmt.Println()

		// ── Step 1: Validate inputs ──
		if initMasterIP == "" {
			return fmt.Errorf("--master IP là bắt buộc")
		}
		if initSSHKey == "" {
			return fmt.Errorf("--key SSH key path là bắt buộc")
		}
		if initDomain == "" {
			return fmt.Errorf("--domain là bắt buộc")
		}

		// ── Step 1.5: Đảm bảo đã có Khóa SSH (Tự sinh nếu thiếu) ──
		expandedKeyPath, err := ssh.EnsureSSHKeyExists(initSSHKey)
		if err != nil {
			return fmt.Errorf("lỗi khởi tạo SSH key: %w", err)
		}
		initSSHKey = expandedKeyPath

		fmt.Println(ui.SectionHeader.Render(" THÔNG TIN CLUSTER "))
		fmt.Printf("  Master IP  : %s\n", ui.Info.Render(initMasterIP))
		fmt.Printf("  Domain     : %s\n", ui.Info.Render(initDomain))
		fmt.Printf("  SSH User   : %s\n", ui.Info.Render(initSSHUser))
		fmt.Printf("  Data Root  : %s\n", ui.Info.Render(initDataRoot))
		fmt.Println()

		// ── Step 2: Check SSH connectivity ──
		fmt.Println(ui.RenderStep(1, 6, "Kiểm tra kết nối SSH..."))
		
		// Tự động sao chép Khóa SSH nếu có truyền Mật khẩu
		if initPass != "" {
			fmt.Println("  Phát hiện cờ --pass, đang thực hiện sao chép SSH Key tự động... (sshpass)")
			
			// Kiểm tra và cài đặt sshpass nếu không có sẵn
			_, err := exec.LookPath("sshpass")
			if err != nil {
				fmt.Println(ui.RenderWarning("Không tìm thấy lệnh 'sshpass'. Đang tự động cài đặt cho bạn..."))
				installCmd := ""
				
				// Phán đoán HĐH
				if _, err := exec.LookPath("apt-get"); err == nil {
					installCmd = "sudo apt-get update && sudo apt-get install -y sshpass"
				} else if _, err := exec.LookPath("yum"); err == nil {
					installCmd = "sudo yum install -y sshpass"
				} else if _, err := exec.LookPath("brew"); err == nil {
					installCmd = "brew install hudochenkov/sshpass/sshpass"
				}
				
				if installCmd != "" {
					exec.Command("sh", "-c", installCmd).Run()
				}
				
				// Kiểm tra lại sau cài
				if _, err := exec.LookPath("sshpass"); err != nil {
					return fmt.Errorf("cần cài đặt thư viện 'sshpass' trên máy của bạn để dùng tính năng --pass. Vui lòng chạy thủ công: sudo apt install sshpass (hoặc tương đương tuỳ hệ điều hành)")
				}
			}
			
			sshCopyCmd := fmt.Sprintf("sshpass -p '%s' ssh-copy-id -o StrictHostKeyChecking=no -i %s %s@%s > /dev/null 2>&1", initPass, initSSHKey, initSSHUser, initMasterIP)
			exec.Command("sh", "-c", sshCopyCmd).Run()
		}

		if err := ssh.CheckConnectivity(initMasterIP, initSSHUser, initSSHKey); err != nil {
			return fmt.Errorf("SSH connection thất bại: %w\n\n  Nguyên nhân: Khóa SSH chưa được cài trên máy chủ.\n  Gợi ý 1: Truyền cờ --pass \"MatKhauRoot\"\n  Gợi ý 2: Chạy lệnh thủ công: ssh-copy-id -i %s %s@%s\n  Gợi ý 3: ssh -i %s %s@%s",
				err, initSSHKey, initSSHUser, initMasterIP, initSSHKey, initSSHUser, initMasterIP)
		}
		fmt.Println(ui.RenderSuccess("Kết nối SSH thành công"))
		fmt.Println()

		// ── Step 3: Check/Install Ansible ──
		fmt.Println(ui.RenderStep(2, 6, "Kiểm tra Ansible..."))
		if !ansible.IsInstalled() {
			fmt.Println(ui.RenderWarning("Ansible chưa được cài. Đang cài đặt..."))
			if err := ansible.InstallAnsible(); err != nil {
				return fmt.Errorf("không thể cài Ansible: %w", err)
			}
		}
		fmt.Println(ui.RenderSuccess("Ansible sẵn sàng"))
		fmt.Println()

		// ── Step 4: Run Ansible - Install Docker & Init Swarm ──
		fmt.Println(ui.RenderStep(3, 6, "Cài đặt Docker và khởi tạo Swarm..."))
		runner := ansible.NewRunner(getPlaybooksDir()).
			WithHost(initMasterIP, initSSHUser, initSSHKey).
			WithVar("data_root", initDataRoot).
			WithVar("domain", initDomain).
			WithVar("node_role", "manager")

		if verbose {
			runner.Verbose = true
		}

		if err := runner.RunPlaybook("cluster-init.yml"); err != nil {
			return fmt.Errorf("cluster init thất bại: %w", err)
		}
		fmt.Println(ui.RenderSuccess("Docker Swarm đã được khởi tạo"))
		fmt.Println()

		// ── Step 5: Deploy Tier 1 (Infrastructure) ──
		fmt.Println(ui.RenderStep(4, 6, "Deploy Tier 1: Traefik + Portainer..."))
		if err := runner.RunPlaybook("deploy-tier1.yml"); err != nil {
			return fmt.Errorf("deploy tier 1 thất bại: %w", err)
		}
		fmt.Println(ui.RenderSuccess("Traefik và Portainer đang chạy"))
		fmt.Println()

		// ── Step 6: Deploy Tier 2 (Platform) ──
		fmt.Println(ui.RenderStep(5, 6, "Deploy Tier 2: MinIO + Database + Monitoring..."))
		if err := runner.RunPlaybook("deploy-tier2.yml"); err != nil {
			return fmt.Errorf("deploy tier 2 thất bại: %w", err)
		}
		fmt.Println(ui.RenderSuccess("Platform services đang chạy"))
		fmt.Println()

		// ── Step 7: Save config ──
		fmt.Println(ui.RenderStep(6, 6, "Lưu cấu hình cluster..."))
		cfg, _ := config.Load()
		cfg.AddCluster(config.Cluster{
			Name:      initClusterName,
			MasterIP:  initMasterIP,
			SSHKey:    initSSHKey,
			SSHUser:   initSSHUser,
			Domain:    initDomain,
			DataRoot:  initDataRoot,
			CreatedAt: time.Now().Format(time.RFC3339),
		})
		if err := cfg.Save(); err != nil {
			return fmt.Errorf("lưu config thất bại: %w", err)
		}
		fmt.Println(ui.RenderSuccess("Config đã được lưu tại ~/.swarm-ctl/config.yml"))
		fmt.Println()

		// ── Done ──
		fmt.Println(ui.SuccessBox.Render(fmt.Sprintf(`
✅ Cluster "%s" đã được khởi tạo thành công!

  🌐 Traefik Dashboard : https://traefik.%s
  📊 Portainer         : https://portainer.%s
  📈 Grafana           : https://grafana.%s
  🗄️  MinIO Console    : https://minio.%s

Tiếp theo:
  swarm-ctl node add --ip <worker-ip> --key %s --role worker
  swarm-ctl service deploy appwrite
  swarm-ctl dashboard
`,
			initClusterName, initDomain, initDomain,
			initDomain, initDomain, initSSHKey)))

		return nil
	},
}

// ──────────────────────────────────────────────
// swarm-ctl cluster status
// ──────────────────────────────────────────────
var clusterStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Xem trạng thái tổng quan của cluster",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		cluster, err := cfg.GetCurrentCluster()
		if err != nil {
			return err
		}

		fmt.Println(ui.Banner.Render(fmt.Sprintf("🐳 Cluster: %s", cluster.Name)))
		fmt.Printf("  Master: %s  |  Domain: %s\n\n", cluster.MasterIP, cluster.Domain)

		// SSH vào master và lấy thông tin
		client := ssh.NewClient(cluster.MasterIP, cluster.SSHUser, cluster.SSHKey)
		if err := client.Connect(); err != nil {
			return fmt.Errorf("không thể kết nối tới master: %w", err)
		}
		defer client.Close()

		// Lấy node list
		fmt.Println(ui.SectionHeader.Render(" NODES "))
		nodeOutput, err := client.Run("docker node ls --format '{{.Hostname}}\t{{.Status}}\t{{.Availability}}\t{{.ManagerStatus}}'")
		if err != nil {
			fmt.Println(ui.RenderWarning("Không thể lấy danh sách nodes: " + err.Error()))
		} else {
			fmt.Println(nodeOutput)
		}

		// Lấy service list
		fmt.Println(ui.SectionHeader.Render(" SERVICES "))
		svcOutput, err := client.Run("docker service ls --format '{{.Name}}\t{{.Mode}}\t{{.Replicas}}\t{{.Image}}'")
		if err != nil {
			fmt.Println(ui.RenderWarning("Không thể lấy danh sách services: " + err.Error()))
		} else {
			fmt.Println(svcOutput)
		}

		return nil
	},
}

// ──────────────────────────────────────────────
// swarm-ctl cluster upgrade
// ──────────────────────────────────────────────
var clusterUpgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Nâng cấp Docker Engine trên tất cả nodes (zero-downtime)",
	Long: `Upgrade Docker Engine trên từng node theo thứ tự:
  1. Drain node (chuyển tasks sang node khác)
  2. Upgrade Docker
  3. Rejoin Swarm
  4. Chờ healthy rồi chuyển sang node tiếp theo`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println(ui.RenderInfo("Tính năng này đang được phát triển (v1.1)"))
		return nil
	},
}

// ──────────────────────────────────────────────
// swarm-ctl cluster destroy
// ──────────────────────────────────────────────
var clusterDestroyForce bool

var clusterDestroyCmd = &cobra.Command{
	Use:   "destroy",
	Short: "Dọn dẹp TRẮNG TINH mọi thứ, đưa Server về trạng thái ban đầu",
	Long: `Lệnh này sẽ ĐẬP ĐI LÀM LẠI mạng lưới của bạn:
  1. Xóa toàn bộ Services & Containers.
  2. Gỡ bỏ Swarm Cluster (Leave --force).
  3. Gỡ cài đặt Docker Engine và xóa Data.
  4. Trả lại con VPS sạch sẽ như mới Mua.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if !clusterDestroyForce {
			fmt.Println(ui.RenderWarning(`CẢNH BÁO NGUY HIỂM TỘT ĐỘ:
Lệnh này sẽ XÓA SẠCH toàn bộ Database, Web, Appwrite và mọi thiết lập của bạn.
Mọi máy chủ VPS của bạn sẽ bị Format sạch sẽ trở lại như lúc mới mua.

Nếu bạn thực sự chắc chắn muốn XÓA BÌNH LÀM LẠI, hãy chạy thêm cờ --force:
  swarm-ctl cluster destroy --force`))
			return nil
		}

		cfg, err := config.Load()
		if err != nil {
			return err
		}
		cluster, err := cfg.GetCurrentCluster()
		if err != nil {
			return err
		}

		fmt.Println(ui.Banner.Render(`💥 CLUSTER SELF-DESTRUCT SEQUENCE INITIATED 💥`))

		client := ssh.NewClient(cluster.MasterIP, cluster.SSHUser, cluster.SSHKey)
		if err := client.Connect(); err != nil {
			return err
		}
		defer client.Close()

		fmt.Println(ui.RenderStep(1, 4, "Đang Xóa toàn bộ Cụm Swarm và Services..."))
		client.Run("docker swarm leave --force")

		fmt.Println(ui.RenderStep(2, 4, "Đang Gỡ cài đặt Docker & Dọn Container..."))
		client.Run("apt-get purge -y docker-ce docker-ce-cli containerd.io && apt-get autoremove -y")

		fmt.Println(ui.RenderStep(3, 4, "Đang Xóa Volume và Mạng Nội bộ (Storage)..."))
		client.Run("rm -rf /var/lib/docker && rm -rf /opt/swarm-ctl* && rm -rf " + cluster.DataRoot)

		// Tháo cluster ra khỏi máy Local
		fmt.Println(ui.RenderStep(4, 4, "Xóa Thông tin Cluster trên máy vận hành (Locally)..."))
		configPath := filepath.Join(os.Getenv("HOME"), ".swarm-ctl", "config.yml")
		if err := os.Remove(configPath); err != nil && !os.IsNotExist(err) {
			return err
		}

		fmt.Println()
		fmt.Println(ui.RenderSuccess("QUÁ TRÌNH PHỦI BỤI HOÀN TẤT. Hệ thống đã trở về cát bụi."))
		fmt.Println("Bây giờ bạn có thể thử chạy `swarm-ctl cluster init` để Setup lại Cụm Mới!")

		return nil
	},
}

func init() {
	// cluster init flags
	clusterInitCmd.Flags().StringVarP(&initMasterIP, "master", "m", "", "IP address của master node (bắt buộc)")
	clusterInitCmd.Flags().StringVarP(&initSSHKey, "key", "k", "~/.ssh/id_ed25519", "Đường dẫn SSH private key")
	clusterInitCmd.Flags().StringVarP(&initSSHUser, "user", "u", "root", "SSH username")
	clusterInitCmd.Flags().StringVarP(&initDomain, "domain", "d", "", "Domain chính của cluster (bắt buộc)")
	clusterInitCmd.Flags().StringVar(&initDataRoot, "data-root", "/opt/data", "Thư mục lưu dữ liệu")
	clusterInitCmd.Flags().StringVarP(&initClusterName, "name", "n", "production", "Tên cluster")
	clusterInitCmd.Flags().StringVarP(&initPass, "pass", "p", "", "Mật khẩu máy chủ (Dùng để tự động Copy SSH Key một lần duy nhất)")

	clusterInitCmd.MarkFlagRequired("master")
	clusterInitCmd.MarkFlagRequired("domain")

	// Đăng ký subcommands
	clusterCmd.AddCommand(clusterInitCmd)
	clusterCmd.AddCommand(clusterStatusCmd)
	clusterCmd.AddCommand(clusterUpgradeCmd)
	clusterCmd.AddCommand(clusterDestroyCmd)
	
	clusterDestroyCmd.Flags().BoolVarP(&clusterDestroyForce, "force", "f", false, "Đồng ý xóa dữ liệu ngay lập tức")
}

// getPlaybooksDir trả về đường dẫn tới Ansible playbooks
// (được bundle cùng binary hoặc có thể override qua env)
func getPlaybooksDir() string {
	if dir := os.Getenv("SWARM_CTL_PLAYBOOKS"); dir != "" {
		return dir
	}
	// Default: cùng thư mục với binary
	exe, err := os.Executable()
	if err != nil {
		return "./ansible/playbooks"
	}
	return fmt.Sprintf("%s/../ansible/playbooks", exe)
}
