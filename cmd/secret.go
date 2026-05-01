package cmd

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/LyVanBong/swarm-ctl/internal/config"
	"github.com/LyVanBong/swarm-ctl/internal/ssh"
	"github.com/LyVanBong/swarm-ctl/internal/ui"
	"github.com/spf13/cobra"
)

var secretCmd = &cobra.Command{
	Use:   "secret",
	Short: "Quản lý Docker Secrets",
}

// ──────────────────────────────────────────────
// swarm-ctl secret add
// ──────────────────────────────────────────────
var secretAddCmd = &cobra.Command{
	Use:   "add [name] [value]",
	Short: "Tạo Docker secret mới",
	Args:  cobra.ExactArgs(2),
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

		secretName, secretValue := args[0], args[1]
		output, err := client.Run(fmt.Sprintf(
			"printf '%%s' '%s' | docker secret create %s -",
			strings.ReplaceAll(secretValue, "'", "'\\''"), secretName))
		if err != nil {
			return fmt.Errorf("tạo secret thất bại: %w\n%s", err, output)
		}
		fmt.Println(ui.RenderSuccess(fmt.Sprintf("Secret '%s' đã được tạo", secretName)))
		return nil
	},
}

// ──────────────────────────────────────────────
// swarm-ctl secret list
// ──────────────────────────────────────────────
var secretListCmd = &cobra.Command{
	Use:     "list",
	Short:   "Xem danh sách secrets",
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

		fmt.Println(ui.SectionHeader.Render(" DOCKER SECRETS "))
		output, err := client.Run(`docker secret ls --format "table {{.Name}}\t{{.CreatedAt}}\t{{.UpdatedAt}}"`)
		if err != nil {
			return err
		}
		lines := strings.Split(strings.TrimSpace(output), "\n")
		for i, line := range lines {
			if i == 0 {
				fmt.Println(ui.Bold.Render(line))
			} else {
				fmt.Println(line)
			}
		}
		return nil
	},
}

// ──────────────────────────────────────────────
// swarm-ctl secret remove
// ──────────────────────────────────────────────
var secretRemoveCmd = &cobra.Command{
	Use:     "remove [name]",
	Short:   "Xóa Docker secret",
	Aliases: []string{"rm"},
	Args:    cobra.ExactArgs(1),
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

		output, err := client.Run(fmt.Sprintf("docker secret rm %s", args[0]))
		if err != nil {
			return fmt.Errorf("xóa secret thất bại: %w\n%s", err, output)
		}
		fmt.Println(ui.RenderSuccess(fmt.Sprintf("Secret '%s' đã được xóa", args[0])))
		return nil
	},
}

// ──────────────────────────────────────────────
// swarm-ctl secret rotate
// ──────────────────────────────────────────────
var secretRotateCmd = &cobra.Command{
	Use:   "rotate [name] [new-value]",
	Short: "Rotate secret an toàn (zero-downtime)",
	Long: `Rotate Docker secret không gây downtime:

Quy trình:
  1. Tạo secret mới tên: <name>_v2
  2. Update tất cả services đang dùng secret cũ → dùng secret mới
  3. Chờ services rolling update xong
  4. Xóa secret cũ
  5. Rename secret mới thành tên gốc (tạo alias)

Ví dụ:
  swarm-ctl secret rotate db_password "newPassw0rd!"`,

	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		secretName, newValue := args[0], args[1]

		cfg, err := config.Load()
		if err != nil {
			return err
		}
		cluster, err := cfg.GetCurrentCluster()
		if err != nil {
			return err
		}

		fmt.Println(ui.Banner.Render(fmt.Sprintf("🔐 ROTATE SECRET: %s", secretName)))
		fmt.Println()

		client := ssh.NewClient(cluster.MasterIP, cluster.SSHUser, cluster.SSHKey)
		if err := client.Connect(); err != nil {
			return err
		}
		defer client.Close()

		newSecretName := secretName + "_new"

		// Step 1: Tạo secret mới
		fmt.Println(ui.RenderStep(1, 4, fmt.Sprintf("Tạo secret mới: %s...", newSecretName)))
		output, err := client.Run(fmt.Sprintf(
			"printf '%%s' '%s' | docker secret create %s -",
			strings.ReplaceAll(newValue, "'", "'\\''"), newSecretName))
		if err != nil {
			return fmt.Errorf("tạo secret mới thất bại: %w\n%s", err, output)
		}
		fmt.Println(ui.RenderSuccess("Secret mới đã được tạo"))

		// Step 2: Tìm services đang dùng secret cũ
		fmt.Println(ui.RenderStep(2, 4, "Tìm services đang dùng secret..."))
		servicesOut, _ := client.Run(fmt.Sprintf(
			`docker service ls -q | xargs -I{} docker service inspect {} --format '{{.Spec.Name}} {{range .Spec.TaskTemplate.ContainerSpec.Secrets}}{{.SecretName}} {{end}}' | grep %s | awk '{print $1}'`,
			secretName))

		affectedServices := []string{}
		for _, s := range strings.Split(strings.TrimSpace(servicesOut), "\n") {
			if s = strings.TrimSpace(s); s != "" {
				affectedServices = append(affectedServices, s)
			}
		}

		if len(affectedServices) == 0 {
			fmt.Println(ui.RenderWarning("Không tìm thấy services nào dùng secret này"))
		} else {
			fmt.Printf("  Services ảnh hưởng: %s\n", ui.Info.Render(strings.Join(affectedServices, ", ")))
		}

		// Step 3: Update services (remove old secret, add new)
		fmt.Println(ui.RenderStep(3, 4, "Update services (rolling update)..."))
		for _, svc := range affectedServices {
			updateCmd := fmt.Sprintf(
				"docker service update --secret-rm %s --secret-add %s %s",
				secretName, newSecretName, svc)
			if _, err := client.Run(updateCmd); err != nil {
				fmt.Println(ui.RenderWarning(fmt.Sprintf("Update %s thất bại: %s", svc, err.Error())))
			} else {
				fmt.Println(ui.RenderSuccess(fmt.Sprintf("Service %s đã được update", svc)))
			}
		}

		// Step 4: Xóa secret cũ và đổi tên secret mới
		fmt.Println(ui.RenderStep(4, 4, "Cleanup secret cũ..."))
		if _, err := client.Run(fmt.Sprintf("docker secret rm %s", secretName)); err != nil {
			fmt.Println(ui.RenderWarning("Không thể xóa secret cũ (có thể đang được dùng): " + err.Error()))
			fmt.Println(ui.RenderInfo(fmt.Sprintf("Secret mới đang dùng tên: %s", newSecretName)))
			fmt.Println(ui.RenderInfo("Xóa thủ công sau khi chắc chắn không còn dùng secret cũ"))
		} else {
			// Tạo secret với tên gốc chứa giá trị mới
			client.Run(fmt.Sprintf(
				"printf '%%s' '%s' | docker secret create %s -",
				strings.ReplaceAll(newValue, "'", "'\\''"), secretName))
			// Xóa secret tạm
			client.Run(fmt.Sprintf("docker secret rm %s", newSecretName))
			fmt.Println(ui.RenderSuccess("Secret rotation hoàn tất"))
		}

		fmt.Println()
		fmt.Println(ui.SuccessBox.Render(fmt.Sprintf(`
✅ Secret "%s" đã được rotate thành công!

   Services đã được rolling update.
   Containers mới sẽ dùng secret mới.
`, secretName)))

		return nil
	},
}

// ──────────────────────────────────────────────
// swarm-ctl secret sync
// ──────────────────────────────────────────────
var (
	syncProject string
	syncEnv     string
)

var secretSyncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Đồng bộ biến môi trường từ Infisical vào Docker Secrets",
	Long: `Lệnh này sẽ dùng Infisical CLI nội bộ để tải JSON variables và đẩy thẳng lên Swarm Secrets,
bảo mật tuyệt đối không lưu ra file vật lý.

Ví dụ:
  swarm-ctl secret sync --project "proj_xyz" --env "prod"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		cluster, err := cfg.GetCurrentCluster()
		if err != nil {
			return err
		}

		if syncProject == "" || syncEnv == "" {
			return fmt.Errorf("cần cung cấp --project và --env")
		}

		fmt.Println(ui.Banner.Render("🔐 INFISICAL SYNC"))
		fmt.Println()

		client := ssh.NewClient(cluster.MasterIP, cluster.SSHUser, cluster.SSHKey)
		if err := client.Connect(); err != nil {
			return err
		}
		defer client.Close()

		fmt.Println(ui.RenderStep(1, 3, "Tải biến môi trường từ Infisical..."))
		
		infisicalCmd := exec.Command("infisical", "export", "--projectId", syncProject, "--env", syncEnv, "--format", "json")
		output, err := infisicalCmd.Output()
		if err != nil {
			return fmt.Errorf("lỗi tải dữ liệu từ infisical (bạn đã login infisical chưa?): %w", err)
		}

		var secrets []map[string]string
		if err := json.Unmarshal(output, &secrets); err != nil {
			// Thử cấu trúc map string string (Tùy phiên bản infisical)
			var secretMap map[string]string
			if err2 := json.Unmarshal(output, &secretMap); err2 != nil {
				return fmt.Errorf("không thể parse JSON từ Infisical: %w", err)
			}
			// Chuyển về struct đồng nhất
			for k, v := range secretMap {
				secrets = append(secrets, map[string]string{"key": k, "value": v})
			}
		}

		fmt.Println(ui.RenderStep(2, 3, fmt.Sprintf("Tìm thấy %d biến, đang bơm vào Swarm...", len(secrets))))

		for _, s := range secrets {
			key, value := s["key"], s["value"]
			if key == "" { // Fallback if simple map was used
				continue
			}

			// Kiểm tra xem secret đã tồn tại chưa
			existsOutput, _ := client.Run(fmt.Sprintf("docker secret ls --filter name=^%s$ -q", key))
			if strings.TrimSpace(existsOutput) != "" {
				// Xóa cũ tạo mới
				client.Run(fmt.Sprintf("docker secret rm %s", key))
			}

			// Tạo mới
			_, err := client.Run(fmt.Sprintf("printf '%%s' '%s' | docker secret create %s -", strings.ReplaceAll(value, "'", "'\\''"), key))
			if err != nil {
				fmt.Println(ui.RenderWarning(fmt.Sprintf("Lỗi đẩy secret %s: %v", key, err)))
			} else {
				fmt.Println(ui.Muted.Render(fmt.Sprintf("  ↳ %s (Sync OK)", key)))
			}
		}

		fmt.Println(ui.RenderStep(3, 3, "Hoàn tất!"))
		fmt.Println(ui.RenderSuccess("Dữ liệu Infisical đã được bơm thành công vào Docker Swarm Secrets!"))

		return nil
	},
}

func init() {
	secretCmd.AddCommand(secretAddCmd)
	secretCmd.AddCommand(secretListCmd)
	secretCmd.AddCommand(secretRemoveCmd)
	secretCmd.AddCommand(secretRotateCmd)
	
	secretSyncCmd.Flags().StringVar(&syncProject, "project", "", "Infisical Project ID")
	secretSyncCmd.Flags().StringVar(&syncEnv, "env", "prod", "Infisical Environment (dev, prod)")
	secretCmd.AddCommand(secretSyncCmd)
}
