package cmd

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/LyVanBong/swarm-ctl/internal/config"
	"github.com/LyVanBong/swarm-ctl/internal/ssh"
	"github.com/LyVanBong/swarm-ctl/internal/ui"
	"github.com/spf13/cobra"
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Kiểm tra và chẩn đoán vấn đề trong cluster",
	Long: `Tự động kiểm tra toàn bộ hệ thống và đưa ra gợi ý sửa lỗi.

Kiểm tra:
  ✓ Kết nối SSH tới master
  ✓ Docker Swarm status
  ✓ Overlay networks
  ✓ Services health
  ✓ Disk space trên các nodes
  ✓ Ansible đã cài chưa
  ✓ Config file hợp lệ không`,

	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println(ui.Banner.Render("🏥 SWARM-CTL DOCTOR"))
		fmt.Println()

		issues := 0

		// ── Check 1: Config ──────────────────────────────────
		fmt.Println(ui.SectionHeader.Render(" CONFIGURATION "))
		cfg, err := config.Load()
		if err != nil {
			fmt.Println(ui.RenderError("Config file lỗi: " + err.Error()))
			issues++
		} else {
			fmt.Println(ui.RenderSuccess("Config file hợp lệ"))
		}

		cluster, err := cfg.GetCurrentCluster()
		if err != nil {
			fmt.Println(ui.RenderError("Không có active cluster — chạy: swarm-ctl cluster init"))
			issues++
			printSummary(issues)
			return nil
		}
		fmt.Printf("  Active cluster: %s (%s)\n", ui.Bold.Render(cluster.Name), cluster.MasterIP)
		fmt.Println()

		// ── Check 2: Ansible ─────────────────────────────────
		fmt.Println(ui.SectionHeader.Render(" TOOLS "))
		if _, err := exec.LookPath("ansible-playbook"); err != nil {
			fmt.Println(ui.RenderWarning("Ansible chưa cài — node provisioning sẽ không hoạt động"))
			fmt.Println(ui.Muted.Render("   Sửa: pip3 install ansible"))
			issues++
		} else {
			out, _ := exec.Command("ansible", "--version").Output()
			version := strings.Split(string(out), "\n")[0]
			fmt.Println(ui.RenderSuccess("Ansible: " + version))
		}
		fmt.Println()

		// ── Check 3: SSH Connectivity ─────────────────────────
		fmt.Println(ui.SectionHeader.Render(" SSH CONNECTIVITY "))
		fmt.Printf("  Đang kết nối tới %s...\n", cluster.MasterIP)

		client := ssh.NewClient(cluster.MasterIP, cluster.SSHUser, cluster.SSHKey)
		if err := client.Connect(); err != nil {
			fmt.Println(ui.RenderError("Không thể SSH tới master: " + err.Error()))
			fmt.Println(ui.Muted.Render("   Kiểm tra: SSH key, firewall, server đang chạy"))
			issues++
			printSummary(issues)
			return nil
		}
		defer client.Close()
		fmt.Println(ui.RenderSuccess("SSH connection OK"))
		fmt.Println()

		// ── Check 4: Docker Swarm ─────────────────────────────
		fmt.Println(ui.SectionHeader.Render(" DOCKER SWARM "))
		swarmState, err := client.Run("docker info --format '{{.Swarm.LocalNodeState}}'")
		if err != nil || strings.TrimSpace(swarmState) != "active" {
			fmt.Println(ui.RenderError("Docker Swarm không active!"))
			fmt.Println(ui.Muted.Render("   Sửa: swarm-ctl cluster init"))
			issues++
		} else {
			fmt.Println(ui.RenderSuccess("Docker Swarm: active"))
		}

		// Node count
		nodeCount, _ := client.Run("docker node ls -q | wc -l")
		fmt.Printf("  Nodes: %s\n", ui.Info.Render(strings.TrimSpace(nodeCount)))
		fmt.Println()

		// ── Check 5: Networks ─────────────────────────────────
		fmt.Println(ui.SectionHeader.Render(" NETWORKS "))
		for _, network := range []string{"proxy_public", "app_internal", "data_net"} {
			out, err := client.Run(fmt.Sprintf("docker network ls --filter name=%s -q", network))
			if err != nil || strings.TrimSpace(out) == "" {
				fmt.Printf("  %s %s\n", ui.RenderError(""), network+" — KHÔNG TỒN TẠI")
				fmt.Printf("       Sửa: docker network create --driver overlay --attachable %s\n", network)
				issues++
			} else {
				fmt.Printf("  %s %s\n", ui.StatusOK, network)
			}
		}
		fmt.Println()

		// ── Check 6: Services health ──────────────────────────
		fmt.Println(ui.SectionHeader.Render(" SERVICES HEALTH "))
		svcOutput, err := client.Run(`docker service ls --format "{{.Name}} {{.Replicas}}"`)
		if err != nil {
			fmt.Println(ui.RenderWarning("Không thể lấy service list"))
		} else {
			lines := strings.Split(strings.TrimSpace(svcOutput), "\n")
			for _, line := range lines {
				if line == "" {
					continue
				}
				parts := strings.Fields(line)
				if len(parts) >= 2 {
					replicas := parts[1] // e.g. "3/3" or "0/3"
					repParts := strings.Split(replicas, "/")
					if len(repParts) == 2 && repParts[0] != repParts[1] {
						fmt.Printf("  %s %-35s [%s]\n", ui.StatusWarn, parts[0],
							ui.Warning.Render(replicas))
						issues++
					} else {
						fmt.Printf("  %s %-35s [%s]\n", ui.StatusOK, parts[0],
							ui.Success.Render(replicas))
					}
				}
			}
		}
		fmt.Println()

		// ── Check 7: Disk space ───────────────────────────────
		fmt.Println(ui.SectionHeader.Render(" DISK SPACE "))
		diskOut, err := client.Run("df -h /opt/data 2>/dev/null | tail -1")
		if err != nil || diskOut == "" {
			diskOut, _ = client.Run("df -h / | tail -1")
		}
		fmt.Println(" " + strings.TrimSpace(diskOut))

		// Kiểm tra > 80% full
		usageOut, _ := client.Run("df /opt/data 2>/dev/null || df / | tail -1 | awk '{print $5}' | tr -d '%'")
		usage := strings.TrimSpace(usageOut)
		if usage >= "80" {
			fmt.Println(ui.RenderWarning("Disk usage cao (>80%) — cân nhắc dọn dẹp hoặc mở rộng"))
			issues++
		}
		fmt.Println()

		// ── Summary ───────────────────────────────────────────
		printSummary(issues)
		return nil
	},
}

func printSummary(issues int) {
	fmt.Println(ui.SectionHeader.Render(" KẾT QUẢ "))
	if issues == 0 {
		fmt.Println(ui.SuccessBox.Render("✅ Tất cả kiểm tra đều pass! Cluster hoạt động bình thường."))
	} else {
		fmt.Println(ui.WarnBox.Render(fmt.Sprintf(
			"⚠️  Phát hiện %d vấn đề cần xử lý.\n   Xem gợi ý sửa lỗi ở trên.", issues)))
	}
}

func init() {
	rootCmd.AddCommand(doctorCmd)
}
