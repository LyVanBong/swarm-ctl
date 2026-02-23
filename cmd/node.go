package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"

	"github.com/LyVanBong/swarm-ctl/internal/ansible"
	"github.com/LyVanBong/swarm-ctl/internal/config"
	"github.com/LyVanBong/swarm-ctl/internal/ssh"
	"github.com/LyVanBong/swarm-ctl/internal/ui"
	"github.com/spf13/cobra"
)

var nodeCmd = &cobra.Command{
	Use:   "node",
	Short: "Quản lý nodes trong cluster",
}

// ──────────────────────────────────────────────
// swarm-ctl node add
// ──────────────────────────────────────────────
var (
	nodeAddIP      string
	nodeAddKey     string
	nodeAddUser    string
	nodeAddRole    string
	nodeAddLabel   []string
	nodeAddPass    string // Password root
)

var nodeAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Thêm node mới vào cluster",
	Long: `Provision và thêm node mới vào Docker Swarm cluster.

Tool sẽ tự động:
  1. Kiểm tra kết nối SSH tới node mới
  2. Cài đặt Docker Engine
  3. Tạo thư mục data (DATA_ROOT)
  4. Mount shared storage (nếu có)
  5. Gia nhập Swarm cluster
  6. Gán node labels

Ví dụ:
  swarm-ctl node add --ip 10.0.0.5 --role worker
  swarm-ctl node add --ip 10.0.0.5 --role worker --label tier=app
  swarm-ctl node add --ip 10.0.0.5 --role manager`,

	RunE: func(cmd *cobra.Command, args []string) error {
		// Load cluster config
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		cluster, err := cfg.GetCurrentCluster()
		if err != nil {
			return err
		}

		// Dùng SSH key của cluster nếu không chỉ định
		keyPath := nodeAddKey
		if keyPath == "" {
			keyPath = cluster.SSHKey
		}
		sshUser := nodeAddUser
		if sshUser == "" {
			sshUser = cluster.SSHUser
		}

		fmt.Println(ui.Banner.Render(fmt.Sprintf("➕ ADD NODE: %s", nodeAddIP)))
		fmt.Printf("  Role   : %s\n", ui.Info.Render(nodeAddRole))
		fmt.Printf("  User   : %s\n", ui.Info.Render(sshUser))
		fmt.Printf("  Key    : %s\n", ui.Info.Render(keyPath))
		if len(nodeAddLabel) > 0 {
			fmt.Printf("  Labels : %s\n", ui.Info.Render(strings.Join(nodeAddLabel, ", ")))
		}
		fmt.Println()

		steps := []string{
			"Kiểm tra kết nối SSH",
			"Cài đặt Docker Engine",
			"Tạo thư mục dữ liệu",
			"Gia nhập Swarm cluster",
			"Gán node labels",
			"Xác nhận node healthy",
		}
		totalSteps := len(steps)

		// ── Step 1: SSH Connectivity ──
		fmt.Println(ui.RenderStep(1, totalSteps, steps[0]+"..."))
		
		// Tự động giải quyết nếu có --pass (Sao chép key sang tự động)
		if nodeAddPass != "" {
			fmt.Println("  Phát hiện cờ --pass, đang thực hiện sao chép SSH Key tự động... (sshpass)")
			
			// Kiểm tra và yêu cầu sshpass
			_, err := exec.LookPath("sshpass")
			if err != nil {
				return fmt.Errorf("cần cài đặt 'sshpass' trên máy của bạn để dùng tính năng --pass. Vui lòng chạy: sudo apt install sshpass (Linux) hoặc brew install hudochenkov/sshpass/sshpass (Mac)")
			}
			
			sshCopyCmd := fmt.Sprintf("sshpass -p '%s' ssh-copy-id -o StrictHostKeyChecking=no -i %s %s@%s > /dev/null 2>&1", nodeAddPass, keyPath, sshUser, nodeAddIP)
			exec.Command("sh", "-c", sshCopyCmd).Run()
		}

		if err := ssh.CheckConnectivity(nodeAddIP, sshUser, keyPath); err != nil {
			return fmt.Errorf(`
❌ Không thể kết nối SSH tới %s

   Nguyên nhân có thể:
   → Máy này chưa được Copy Key (Có thể cài tự động bằng cách thêm cờ: --pass "MatKhauMayChu")
   → SSH key không đúng hoặc Server chưa bật SSH service
   → Firewall đang chặn port 22

   Gợi ý Test thủ công: ssh -i %s %s@%s`, nodeAddIP, keyPath, sshUser, nodeAddIP)
		}
		fmt.Println(ui.RenderSuccess("Kết nối SSH thành công"))

		// ── Step 2-5: Ansible Playbook ──
		fmt.Println(ui.RenderStep(2, totalSteps, steps[1]+"..."))
		runner := ansible.NewRunner(getPlaybooksDir()).
			WithHost(nodeAddIP, sshUser, keyPath).
			WithVar("node_role", nodeAddRole).
			WithVar("master_ip", cluster.MasterIP).
			WithVar("data_root", cluster.DataRoot).
			WithVar("node_labels", strings.Join(nodeAddLabel, ","))

		if verbose {
			runner.Verbose = true
		}

		if err := runner.RunPlaybook("node-add.yml"); err != nil {
			return fmt.Errorf("node provision thất bại: %w\n\n  Xem logs chi tiết: swarm-ctl node add ... --verbose", err)
		}

		// ── Step 6: Verify node is healthy ──
		fmt.Println(ui.RenderStep(6, totalSteps, steps[5]+"..."))
		masterClient := ssh.NewClient(cluster.MasterIP, cluster.SSHUser, cluster.SSHKey)
		if err := masterClient.Connect(); err == nil {
			defer masterClient.Close()
			// Đợi node xuất hiện trong swarm
			time.Sleep(3 * time.Second)
			output, err := masterClient.Run(fmt.Sprintf(
				"docker node ls --filter 'name=%s' --format '{{.Status}}'", nodeAddIP))
			if err == nil && strings.TrimSpace(output) == "Ready" {
				fmt.Println(ui.RenderSuccess("Node đã sẵn sàng trong cluster"))
			}
		}

		fmt.Println()
		fmt.Println(ui.SuccessBox.Render(fmt.Sprintf(`
✅ Node %s đã được thêm thành công!

   Role   : %s
   Labels : %s

   Tiếp theo:
   → swarm-ctl node list       (xem danh sách nodes)
   → swarm-ctl service scale   (tận dụng node mới)
`, nodeAddIP, nodeAddRole, strings.Join(nodeAddLabel, ", "))))

		return nil
	},
}

// ──────────────────────────────────────────────
// swarm-ctl node remove
// ──────────────────────────────────────────────
var (
	nodeRemoveIP    string
	nodeRemoveForce bool
)

var nodeRemoveCmd = &cobra.Command{
	Use:   "remove",
	Short: "Xóa node khỏi cluster (an toàn)",
	Long: `Drain node trước khi xóa để đảm bảo không mất service.

Quá trình:
  1. Drain node (chuyển tasks sang node khác)
  2. Đợi tất cả tasks được reschedule
  3. Leave swarm
  4. Remove khỏi cluster

Ví dụ:
  swarm-ctl node remove --ip 10.0.0.5`,

	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		cluster, err := cfg.GetCurrentCluster()
		if err != nil {
			return err
		}

		fmt.Println(ui.Banner.Render(fmt.Sprintf("🗑️  REMOVE NODE: %s", nodeRemoveIP)))
		fmt.Println()
		fmt.Println(ui.WarnBox.Render(`
⚠️  CẢNH BÁO: Thao tác này sẽ:
   1. Drain tất cả services khỏi node
   2. Xóa node khỏi Swarm cluster
   
   Services sẽ được tự động reschedule sang nodes khác.
   Đảm bảo cluster có đủ capacity trước khi tiếp tục!`))
		fmt.Println()

		if !nodeRemoveForce {
			fmt.Print("Nhập IP node để xác nhận xóa: ")
			var confirm string
			fmt.Scanln(&confirm)
			if confirm != nodeRemoveIP {
				return fmt.Errorf("IP không khớp — đã hủy thao tác")
			}
		}

		masterClient := ssh.NewClient(cluster.MasterIP, cluster.SSHUser, cluster.SSHKey)
		if err := masterClient.Connect(); err != nil {
			return fmt.Errorf("không thể kết nối master: %w", err)
		}
		defer masterClient.Close()

		// Step 1: Drain node
		fmt.Println(ui.RenderStep(1, 3, "Drain node (chuyển tasks sang nodes khác)..."))
		output, err := masterClient.Run(fmt.Sprintf(
			"docker node update --availability drain $(docker node ls --filter 'addr=%s' -q)", nodeRemoveIP))
		if err != nil {
			return fmt.Errorf("không thể drain node: %w\n%s", err, output)
		}
		fmt.Println(ui.RenderSuccess("Node đã được drain"))

		// Step 2: Đợi tasks reschedule
		fmt.Println(ui.RenderStep(2, 3, "Đợi tasks reschedule (tối đa 60s)..."))
		time.Sleep(10 * time.Second)

		// Step 3: Remove node
		fmt.Println(ui.RenderStep(3, 3, "Xóa node khỏi cluster..."))
		output, err = masterClient.Run(fmt.Sprintf(
			"docker node rm $(docker node ls --filter 'addr=%s' -q) --force", nodeRemoveIP))
		if err != nil {
			return fmt.Errorf("không thể xóa node: %w\n%s", err, output)
		}

		fmt.Println()
		fmt.Println(ui.RenderSuccess(fmt.Sprintf("Node %s đã được xóa khỏi cluster", nodeRemoveIP)))
		return nil
	},
}

// ──────────────────────────────────────────────
// swarm-ctl node list
// ──────────────────────────────────────────────
var nodeListCmd = &cobra.Command{
	Use:   "list",
	Short: "Xem danh sách tất cả nodes",
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
			return fmt.Errorf("không thể kết nối master: %w", err)
		}
		defer client.Close()

		fmt.Println(ui.SectionHeader.Render(fmt.Sprintf(" NODES — Cluster: %s ", cluster.Name)))

		output, err := client.Run(`docker node ls --format "table {{.Hostname}}\t{{.Status}}\t{{.Availability}}\t{{.ManagerStatus}}\t{{.EngineVersion}}"`)
		if err != nil {
			return err
		}

		// Format output với màu sắc
		lines := strings.Split(strings.TrimSpace(output), "\n")
		for i, line := range lines {
			if i == 0 {
				fmt.Println(ui.Bold.Render(line))
			} else {
				if strings.Contains(line, "Ready") {
					fmt.Println(strings.Replace(line, "Ready", ui.Success.Render("Ready"), 1))
				} else if strings.Contains(line, "Down") {
					fmt.Println(strings.Replace(line, "Down", ui.Danger.Render("Down"), 1))
				} else {
					fmt.Println(line)
				}
			}
		}
		return nil
	},
}

// ──────────────────────────────────────────────
// swarm-ctl node ssh
// ──────────────────────────────────────────────
var nodeSSHCmd = &cobra.Command{
	Use:   "ssh [ip-or-hostname]",
	Short: "SSH trực tiếp vào node",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		cluster, err := cfg.GetCurrentCluster()
		if err != nil {
			return err
		}

		targetIP := args[0]
		// Exec ssh trực tiếp để user có interactive shell
		execPath, err := exec.LookPath("ssh")
		if err != nil {
			return fmt.Errorf("ssh không có trong PATH")
		}

		sshArgs := []string{
			"ssh",
			"-i", cluster.SSHKey,
			"-o", "StrictHostKeyChecking=no",
			fmt.Sprintf("%s@%s", cluster.SSHUser, targetIP),
		}

		fmt.Printf("🔗 SSH vào %s@%s ...\n", cluster.SSHUser, targetIP)
		return syscall.Exec(execPath, sshArgs, os.Environ())
	},
}

// ──────────────────────────────────────────────
// swarm-ctl node label
// ──────────────────────────────────────────────
var nodeLabelCmd = &cobra.Command{
	Use:   "label",
	Short: "Quản lý node labels",
}

var nodeLabelAddCmd = &cobra.Command{
	Use:   "add [node-ip] [key=value...]",
	Short: "Thêm labels vào node",
	Args:  cobra.MinimumNArgs(2),
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

		targetIP := args[0]
		labels := args[1:]

		// Lấy node ID
		nodeID, err := client.Run(fmt.Sprintf(
			"docker node ls --filter 'addr=%s' --format '{{.ID}}'", targetIP))
		if err != nil || strings.TrimSpace(nodeID) == "" {
			// Thử tìm theo hostname
			nodeID, err = client.Run(fmt.Sprintf(
				"docker node ls --filter 'name=%s' --format '{{.ID}}'", targetIP))
			if err != nil || strings.TrimSpace(nodeID) == "" {
				return fmt.Errorf("không tìm thấy node '%s'", targetIP)
			}
		}
		nodeID = strings.TrimSpace(nodeID)

		// Thêm thi labels
		labelArgs := strings.Join(func() []string {
			var result []string
			for _, l := range labels {
				result = append(result, "--label-add "+l)
			}
			return result
		}(), " ")

		output, err := client.Run(fmt.Sprintf(
			"docker node update %s %s", labelArgs, nodeID))
		if err != nil {
			return fmt.Errorf("thêm label thất bại: %w\n%s", err, output)
		}

		fmt.Println(ui.RenderSuccess(fmt.Sprintf(
			"Labels [%s] đã được thêm vào node %s",
			strings.Join(labels, ", "), targetIP)))
		return nil
	},
}

var nodeLabelRemoveCmd = &cobra.Command{
	Use:   "remove [node-ip] [key...]",
	Short: "Xóa labels khỏi node",
	Args:  cobra.MinimumNArgs(2),
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

		targetIP := args[0]
		labelKeys := args[1:]

		nodeID, _ := client.Run(fmt.Sprintf(
			"docker node ls --filter 'addr=%s' --format '{{.ID}}'", targetIP))
		nodeID = strings.TrimSpace(nodeID)
		if nodeID == "" {
			return fmt.Errorf("không tìm thấy node '%s'", targetIP)
		}

		labelArgs := strings.Join(func() []string {
			var result []string
			for _, k := range labelKeys {
				result = append(result, "--label-rm "+k)
			}
			return result
		}(), " ")

		output, err := client.Run(fmt.Sprintf(
			"docker node update %s %s", labelArgs, nodeID))
		if err != nil {
			return fmt.Errorf("xóa label thất bại: %w\n%s", err, output)
		}

		fmt.Println(ui.RenderSuccess(fmt.Sprintf("Labels đã được xóa khỏi node %s", targetIP)))
		return nil
	},
}

func init() {
	// node add flags
	nodeAddCmd.Flags().StringVarP(&nodeAddIP, "ip", "i", "", "IP address của node mới (bắt buộc)")
	nodeAddCmd.Flags().StringVarP(&nodeAddKey, "key", "k", "", "SSH private key (mặc định: key của cluster)")
	nodeAddCmd.Flags().StringVarP(&nodeAddUser, "user", "u", "", "SSH username (mặc định: user của cluster)")
	nodeAddCmd.Flags().StringVarP(&nodeAddRole, "role", "r", "worker", "Node role: worker | manager")
	nodeAddCmd.Flags().StringArrayVarP(&nodeAddLabel, "label", "l", []string{}, "Node labels (vd: tier=app)")
	nodeAddCmd.Flags().StringVarP(&nodeAddPass, "pass", "p", "", "Mật khẩu máy chủ (Dùng để tự động Copy SSH Key một lần duy nhất)")
	nodeAddCmd.MarkFlagRequired("ip")

	// node remove flags
	nodeRemoveCmd.Flags().StringVarP(&nodeRemoveIP, "ip", "i", "", "IP address của node cần xóa (bắt buộc)")
	nodeRemoveCmd.Flags().BoolVarP(&nodeRemoveForce, "force", "f", false, "Bỏ qua xác nhận")
	nodeRemoveCmd.MarkFlagRequired("ip")

	// Register subcommands
	nodeCmd.AddCommand(nodeAddCmd)
	nodeCmd.AddCommand(nodeRemoveCmd)
	nodeCmd.AddCommand(nodeListCmd)
	nodeCmd.AddCommand(nodeSSHCmd)

	// node label subcommands
	nodeLabelCmd.AddCommand(nodeLabelAddCmd)
	nodeLabelCmd.AddCommand(nodeLabelRemoveCmd)
	nodeCmd.AddCommand(nodeLabelCmd)
}
